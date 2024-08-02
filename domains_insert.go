package main

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"

	"golang.org/x/net/idna"
)

func DomainInsertScan(Scan map[string]any) error {
	// TODO: check validity + existence
	domain, _ := Scan["domain"].(string)
	tags, _ := Scan["tags"].([]any)
	services := Scan["services"].(map[string]any)
	check_time := Scan["meta"].(map[string]any)["started_at"].(string)
	dns_records := Scan["dns_records"].(map[string]any)

	is_active := 0
	if len(dns_records) != 0 {
		is_active = 1
	}

	flags := uint64(0)

	for _, tag := range tags {
		k, _ := tag.(string)
		if v, e := DomainScanTag2Flag[k]; e {
			flags |= v
		}
	}

	for k := range services {
		switch k {
		case "http":
			flags |= DOMAIN_HTTP_FLAG
		case "https":
			flags |= DOMAIN_HTTPS_FLAG
		}
	}

	domain, _ = idna.ToASCII(domain)

	var domain_id int64 = 0
	GlobalContext.Database.Get(&domain_id, "SELECT id FROM domains WHERE domain=?", domain)

	if domain_id == 0 {
		res := GlobalContext.Database.MustExec(`
		INSERT INTO domains(domain,is_active,cur_flags,last_check,check_ver) VALUE(?,?,?,?,?)`,
			domain, is_active, flags, check_time[:19], Scan["version"])
		domain_id, _ = res.LastInsertId()
	} else {
		GlobalContext.Database.MustExec(`
		UPDATE domains SET is_active=?, cur_flags=?, last_check=?, check_ver=?
		WHERE id=?`,
			is_active, flags, check_time[:19], Scan["version"], domain_id)
	}

	{
		json_str, _ := json.Marshal(Scan)
		GlobalContext.Database.MustExec(`INSERT INTO domain_scan_archives(domain_id,raw_result) VALUE(?,?)`,
			domain_id, json_str)
	}

	DomainInsertRecords(domain_id, dns_records)

	GlobalContext.Database.MustExec("UPDATE http_services SET is_active=0 WHERE domain_id=?", domain_id)
	for k, v := range services {
		switch k {
		case "http":
			HttpServiceInsert(domain_id, domain, 80, 0, v.(map[string]any))
		case "https":
			HttpServiceInsert(domain_id, domain, 443, 1, v.(map[string]any))
		}

	}

	return nil
}

func DomainInsertRecords(domain_id int64, records map[string]any) {
	type UnifiedRecord struct {
		Type     uint16
		Addr     *string
		Priority *uint16
	}
	var output_records []UnifiedRecord

	if a_records, e := records["A"].([]any); e {
		for _, record := range a_records {
			str_record, _ := record.(string)
			output_records = append(output_records, UnifiedRecord{
				Type: 1,
				Addr: &str_record})
		}
	}

	if aaaa_records, e := records["AAAA"].([]any); e {
		for _, record := range aaaa_records {
			str_record, _ := record.(string)
			output_records = append(output_records, UnifiedRecord{
				Type: 28,
				Addr: &str_record})
		}
	}

	if mx_records, e := records["MX"].([]any); e {
		for _, record := range mx_records {
			str_record, _ := record.(string)

			split := strings.Split(str_record, " ")
			if len(split) != 2 {
				continue
			}
			priority, _ := strconv.Atoi(split[0])
			uin16_prit := uint16(priority)

			output_records = append(output_records, UnifiedRecord{
				Type:     28,
				Addr:     &split[1],
				Priority: &uin16_prit})
		}
	}

	if len(output_records) == 0 {
		GlobalContext.Database.MustExec("UPDATE dns_records SET is_active=0 WHERE domain_id=?", domain_id)
		return
	}

	already_present_records := []DNSRecordRow{}

	// reuse records that already exist
	// TODO: chunk query, it's unlikely to exceed the placeholder limit, but not impossible
	{
		var vars = []any{domain_id}
		buff := bytes.NewBufferString("SELECT id, type, addr, priority FROM dns_records WHERE domain_id=? AND (")
		for i, record := range output_records {
			if i != 0 {
				buff.WriteString(" OR ")
			}
			if len(output_records) != 1 {
				buff.WriteByte('(')
			}

			buff.WriteString("type=? AND ")
			vars = append(vars, record.Type)
			if record.Addr == nil {
				buff.WriteString("addr IS NULL AND ")
			} else {
				buff.WriteString("addr=? AND ")
				vars = append(vars, *record.Addr)
			}
			if record.Priority == nil {
				buff.WriteString("priority IS NULL")
			} else {
				buff.WriteString("priority=?")
				vars = append(vars, *record.Priority)
			}

			if len(output_records) != 1 {
				buff.WriteByte(')')
			}
		}
		buff.WriteByte(')')
		rows, err := GlobalContext.Database.Queryx(buff.String(), vars...)
		AssertError(err)

		for rows.Next() {
			var row DNSRecordRow
			AssertError(rows.StructScan(&row))

			already_present_records = append(already_present_records, row)
		}
	}

	if len(already_present_records) != 0 {
		var vars = []any{domain_id}
		buff := bytes.NewBufferString("UPDATE dns_records SET is_active=0 WHERE domain_id=? AND id NOT IN (")
		for i, record := range already_present_records {
			if i != 0 {
				buff.WriteString(",?")
			} else {
				buff.WriteByte('?')
			}
			vars = append(vars, record.Id)
		}
		buff.WriteByte(')')
		GlobalContext.Database.MustExec(buff.String(), vars...)
	} else {
		GlobalContext.Database.MustExec("UPDATE dns_records SET is_active=0 WHERE domain_id=?", domain_id)
	}

	for _, record := range output_records {
		is_present := false
		for _, present_record := range already_present_records {
			/* check if it is the same */
			if present_record.Type != record.Type {
				continue
			} else if record.Addr == nil && present_record.Addr.Valid {
				continue
			} else if record.Addr != nil && *record.Addr != present_record.Addr.String {
				continue
			} else if record.Priority == nil && present_record.Priority.Valid {
				continue
			} else if record.Priority != nil && present_record.Priority.Int32 != int32(*record.Priority) {
				continue
			}

			is_present = true
			break
		}

		if is_present {
			continue
		}
		GlobalContext.Database.MustExec(`
		INSERT INTO dns_records(domain_id,is_active,type,addr,priority) VALUE(?,1,?,?,?)`,
			domain_id, record.Type, record.Addr, record.Priority)
	}
}
