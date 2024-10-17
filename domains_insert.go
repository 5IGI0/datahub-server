package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"golang.org/x/net/idna"
)

func DomainInsertScan(Scan map[string]any) error {
	// TODO: check validity + existence
	domain, _ := Scan["domain"].(string)
	tags, _ := Scan["tags"].([]any)
	services, _ := Scan["services"].(map[string]any)
	meta, _ := Scan["meta"].(map[string]any)
	check_time, e := meta["check_time"].(string)
	if !e {
		check_time = time.Now().Format("2006-01-02T15:04:05")
	}
	dns_records, _ := Scan["dns_records"].(map[string]any)

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

	InsertHashIdBasedRows(Records, "dns_records", squirrel.Eq{"domain_id": domain_id},
		func(r DNSRecordRow) map[string]interface{} {
			return map[string]any{
				"domain_id": domain_id,
				"is_active": 1,
				"type":      r.Type,
				"addr":      r.Addr,
				"priority":  r.Priority}
		}, nil)
}

func InsertDomains(domains []string) {
	for i := 0; i < len(domains); i += MAX_SQLX_PLACEHOLDERS {
		tmp := domains[i:]
		if len(domains)-i > MAX_SQLX_PLACEHOLDERS {
			tmp = domains[i : i+MAX_SQLX_PLACEHOLDERS]
		}

		_InsertDomains(tmp)
	}
}

func _InsertDomains(domains []string) {
	domains = SanitizeDomains(domains)
	if len(domains) == 0 {
		return
	}

	var q = bytes.NewBufferString("INSERT IGNORE INTO domains(domain) VALUE")
	var d = make([]any, 0, len(domains))

	for i, dd := range domains {
		if i != 0 {
			q.WriteByte(',')
		}
		q.WriteString("(?)")
		d = append(d, dd)
	}

	GlobalContext.Database.MustExec(q.String(), d...)
}
