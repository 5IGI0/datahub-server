package main

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/Masterminds/squirrel"
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
	var Records []DNSRecordRow
	var PresentRecords []DNSRecordRow

	FillRecord := func(Type uint16, records []any) {
		for _, record := range records {
			str_record, _ := record.(string)
			Records = append(Records, DNSRecordRow{
				Type: Type,
				Addr: sql.NullString{Valid: true, String: str_record}})
		}
	}

	if a_records, e := records["A"].([]any); e {
		FillRecord(1, a_records)
	}

	if a_records, e := records["AAAA"].([]any); e {
		FillRecord(28, a_records)
	}

	if mx_records, e := records["MX"].([]any); e {
		for _, record := range mx_records {
			str_record, _ := record.(string)

			split := strings.Split(str_record, " ")
			if len(split) != 2 {
				continue
			}
			priority, _ := strconv.Atoi(split[0])

			Records = append(Records, DNSRecordRow{
				Type:     28,
				Addr:     sql.NullString{Valid: true, String: split[1]},
				Priority: sql.NullInt32{Valid: true, Int32: int32(priority)},
			})
		}
	}

	for i := range Records {
		Records[i].DomainId = domain_id
		Records[i].HashId = Records[i].CompHashId()
	}

	if len(Records) == 0 {
		GlobalContext.Database.MustExec("UPDATE dns_records SET is_active=0 WHERE domain_id=?", domain_id)
		return
	}

	// reuse records that already exist
	// TODO: chunk query, it's unlikely to exceed the placeholder limit, but not impossible
	{
		var conds squirrel.Or

		for _, r := range Records {
			conds = append(conds, squirrel.Eq{"hash_id": r.HashId})
		}

		q, v := squirrel.
			Select("id", "hash_id").
			From("dns_records").
			Where(conds).MustSql()

		rows, err := GlobalContext.Database.Queryx(q, v...)
		AssertError(err)

		for rows.Next() {
			var row DNSRecordRow
			AssertError(rows.StructScan(&row))
			PresentRecords = append(PresentRecords, row)
		}
	}

	/* update is_active */
	{
		var conds squirrel.And

		for _, r := range PresentRecords {
			conds = append(conds, squirrel.NotEq{"id": r.Id})
		}
		q, v := squirrel.
			Update("dns_records").
			Set("is_active", 0).
			Where(conds).MustSql()
		GlobalContext.Database.MustExec(q, v...)
	}

	for _, record := range Records {
		is_present := false
		for _, PresentRecord := range PresentRecords {
			if PresentRecord.HashId == record.HashId {
				is_present = true
				break
			}
		}
		if is_present {
			continue
		}

		SetMap := make(map[string]any)
		SetMap["domain_id"] = domain_id
		SetMap["is_active"] = 1
		SetMap["type"] = record.Type

		if record.Addr.Valid {
			SetMap["addr"] = record.Addr.String
		} else {
			SetMap["addr"] = nil
		}

		if record.Priority.Valid {
			SetMap["priority"] = record.Priority.Int32
		} else {
			SetMap["priority"] = nil
		}

		q, v := squirrel.
			Insert("dns_records").
			SetMap(SetMap).
			MustSql()
		GlobalContext.Database.MustExec(q, v...)
	}
}
