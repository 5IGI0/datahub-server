package main

import (
	"encoding/json"

	"github.com/Masterminds/squirrel"
)

func HttpServiceInsert(domain_id int64, domain string, port uint16, secure int8, data map[string]any) {
	var service_id int64 = 0
	var cert_id any = nil

	raw_result, _ := json.Marshal(data)

	if certificate, e := data["certificate"].(map[string]any); e {
		var cert_row SSLCertificateRow
		AssertError(cert_row.FromMap(certificate))
		cert_id = SSLCertificateInsert(cert_row)
	}

	GlobalContext.Database.Get(&service_id, "SELECT id FROM http_services WHERE domain_id=? AND secure=? AND port=?",
		domain_id, secure, port)

	page_title := data["title"]
	if page_title != nil {
		page_title = TruncateText(page_title.(string), 255)
	}

	path := data["path"]
	if path != nil {
		path = TruncateText(path.(string), 255)
	}

	is_new_cert := true
	if service_id == 0 {
		service_id, _ = GlobalContext.Database.MustExec(`INSERT INTO http_services(
			      is_active,domain_id,domain,secure,port,page_title,status_code,actual_path,raw_result,certificate_id)
			VALUE(1,?,?,?,?,?,?,?,?,?)`,
			domain_id, domain, secure, port, page_title, data["status_code"], path, raw_result, cert_id).LastInsertId()

	} else {
		is_new_cert = false
		GlobalContext.Database.MustExec(`UPDATE http_services
		SET is_active=1,page_title=?,status_code=?,actual_path=?,raw_result=?,certificate_id=?
		WHERE id=?`,
			page_title, data["status_code"], path, raw_result, cert_id, service_id)
	}

	if cert_id != nil {
		if !is_new_cert {
			old_cert_id := 0
			GlobalContext.Database.Get(&old_cert_id, `
				SELECT certificate_id FROM http_certificate_history WHERE service_id=? ORDER BY observed_at DESC LIMIT 1
			`, service_id)

			if old_cert_id != cert_id {
				is_new_cert = true
			}
		}

		if is_new_cert {
			GlobalContext.Database.MustExec(`
					INSERT INTO http_certificate_history(service_id, certificate_id, observed_at)
					VALUE(?,?,NOW())`, service_id, cert_id)
		}
	}

	HttpServiceInsertHeader(service_id, data["headers"].(map[string]any))
	HttpServiceInsertDocumentMeta(service_id, data["html_meta"].([]any))
	if v, e := data["robots_txt"]; e {
		HttpServiceInsertRobotTxt(service_id, v.([]any))
	}
}

func HttpServiceInsertHeader(service_id int64, headers map[string]any) {
	var Headers []HttpHeaderRow

	for k, v := range headers {
		var Header HttpHeaderRow

		Header.ServiceId = service_id
		Header.Value = TruncateText(v.(string), 127)
		Header.Key = TruncateText(k, 127)
		Header.HashId = Header.CompHashId()

		Headers = append(Headers, Header)
	}

	InsertHashIdBasedRows(Headers, "http_headers", squirrel.Eq{"service_id": service_id}, func(row HttpHeaderRow) map[string]any {
		return map[string]any{
			"`service_id`": service_id,
			"`is_active`":  1,
			"`key`":        row.Key,
			"`value`":      row.Value}
	}, func(HttpHeaderRow, int64) map[string]any { return nil })
}

func HttpServiceInsertDocumentMeta(service_id int64, meta []any) {
	var MetaList []HttpDocumentMetaRow

	for _, m := range meta {
		var Meta HttpDocumentMetaRow

		m := m.(map[string]any)

		Meta.ServiceId = service_id
		Meta.Property, _ = m["property"].(string)
		Meta.Content, _ = m["content"].(string)

		Meta.Property = TruncateText(Meta.Property, 127)
		Meta.Content = TruncateText(Meta.Content, 127)
		Meta.HashId = Meta.CompHashId()

		MetaList = append(MetaList, Meta)
	}

	InsertHashIdBasedRows(MetaList, "http_document_meta", squirrel.Eq{"service_id": service_id}, func(row HttpDocumentMetaRow) map[string]any {
		return map[string]any{
			"service_id": service_id,
			"is_active":  1,
			"property":   row.Property,
			"content":    row.Content}
	}, func(HttpDocumentMetaRow, int64) map[string]any { return nil })
}

func HttpServiceInsertRobotTxt(service_id int64, directives []any) {
	var Directives []HttpRobotsTxtRow

	for _, d := range directives {
		var Directive HttpRobotsTxtRow

		d := d.(map[string]any)

		Directive.UserAgent, _ = d["useragent"].(string)
		Directive.Directive, _ = d["directive"].(string)
		Directive.Value, _ = d["data"].(string)

		Directive.ServiceId = service_id
		Directive.UserAgent = TruncateText(Directive.UserAgent, 63)
		Directive.Directive = TruncateText(Directive.Directive, 127)
		Directive.Value = TruncateText(Directive.Value, 512)

		Directives = append(Directives, Directive)
	}

	InsertHashIdBasedRows(Directives, "http_robots_txt", squirrel.Eq{"service_id": service_id}, func(row HttpRobotsTxtRow) map[string]any {
		return map[string]any{
			"service_id": service_id,
			"is_active":  1,
			"useragent":  row.UserAgent,
			"directive":  row.Directive,
			"value":      row.Value}
	}, func(HttpRobotsTxtRow, int64) map[string]any { return nil })
}
