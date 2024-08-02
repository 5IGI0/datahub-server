package main

import (
	"bytes"
	"encoding/json"
)

func HttpServiceInsert(domain_id int64, domain string, port uint16, secure int8, data map[string]any) {
	var service_id int64 = 0

	raw_result, _ := json.Marshal(data)

	GlobalContext.Database.Get(&service_id, "SELECT id FROM http_services WHERE domain_id=? AND secure=? AND port=?",
		domain_id, secure, port)

	if service_id == 0 {
		service_id, _ = GlobalContext.Database.MustExec(`INSERT INTO http_services(
			      is_active,domain_id,domain,secure,port,page_title,status_code,actual_path,raw_result)
			VALUE(1,?,?,?,?,?,?,?,?)`,
			domain_id, domain, secure, port, data["title"], data["status_code"], data["path"], raw_result).LastInsertId()
	} else {
		GlobalContext.Database.MustExec(`UPDATE http_services
		SET is_active=1,page_title=?,status_code=?,actual_path=?,raw_result=?
		WHERE id=?`,
			data["title"], data["status_code"], data["path"], raw_result, service_id)
	}

	HttpServiceInsertHeader(service_id, data["headers"].(map[string]any))
	HttpServiceInsertDocumentMeta(service_id, data["html_meta"].([]any))
}

func HttpServiceInsertHeader(service_id int64, headers map[string]any) {
	var PresentHeader []HttpHeaderRow

	if len(headers) == 0 {
		GlobalContext.Database.Exec("UPDATE http_headers SET is_active=0 WHERE service_id=?")
		return
	}

	{
		vars := []any{service_id}
		buff := bytes.NewBufferString("SELECT `id`, `key`, `value` FROM `http_headers` WHERE `service_id`=? AND (")
		for k, v := range headers {
			if len(vars) != 1 {
				buff.WriteString(" OR ")
			}
			if len(headers) != 1 {
				buff.WriteByte('(')
			}

			v = TruncateText(v.(string), 127)
			k = TruncateText(k, 127)

			buff.WriteString("`key`=? AND `value`=?")
			vars = append(vars, k, v)

			if len(headers) != 1 {
				buff.WriteByte(')')
			}
		}

		buff.WriteByte(')')
		rows, err := GlobalContext.Database.Queryx(buff.String(), vars...)
		AssertError(err)

		for rows.Next() {
			var row HttpHeaderRow
			AssertError(rows.StructScan(&row))
			PresentHeader = append(PresentHeader, row)
		}
	}

	if len(PresentHeader) == 0 {
		GlobalContext.Database.MustExec("UPDATE http_headers SET is_active=0 WHERE service_id=?", service_id)
	} else {
		vars := []any{service_id}
		buff := bytes.NewBufferString("UPDATE http_headers SET is_active=0 WHERE service_id=? AND id NOT IN (")
		for i, v := range PresentHeader {
			if i == 0 {
				buff.WriteByte('?')
			} else {
				buff.WriteString(",?")
			}
			vars = append(vars, v.Id)
		}
		buff.WriteByte(')')
		GlobalContext.Database.MustExec(buff.String(), vars...)
	}

	for k, v := range headers {

		v := TruncateText(v.(string), 127)
		k = TruncateText(k, 127)

		is_present := false
		for _, vk := range PresentHeader {
			if vk.Key == k && vk.Value == v {
				is_present = true
				break
			}
		}

		if is_present {
			continue
		}
		GlobalContext.Database.MustExec("INSERT INTO `http_headers`(`service_id`,`is_active`,`key`,`value`) VALUE(?,1,?,?)",
			service_id, k, v)
		PresentHeader = append(PresentHeader, HttpHeaderRow{Key: k, Value: v})
	}
}

func HttpServiceInsertDocumentMeta(service_id int64, meta []any) {
	var PresentMeta []HttpDocumentMetaRow

	if len(meta) == 0 {
		GlobalContext.Database.Exec("UPDATE http_document_meta SET is_active=0 WHERE service_id=?")
		return
	}

	{
		vars := []any{service_id}
		buff := bytes.NewBufferString("SELECT `id`, `property`, `content` FROM `http_document_meta` WHERE `service_id`=? AND (")
		for i, v := range meta {
			if i != 0 {
				buff.WriteString(" OR ")
			}
			if len(meta) != 1 {
				buff.WriteByte('(')
			}

			p := TruncateText(v.(map[string]any)["property"].(string), 127)
			c := TruncateText(v.(map[string]any)["content"].(string), 127)
			buff.WriteString("`property`=? AND `content`=?")
			vars = append(vars, p, c)

			if len(meta) != 1 {
				buff.WriteByte(')')
			}
		}

		buff.WriteByte(')')
		rows, err := GlobalContext.Database.Queryx(buff.String(), vars...)
		AssertError(err)

		for rows.Next() {
			var row HttpDocumentMetaRow
			AssertError(rows.StructScan(&row))
			PresentMeta = append(PresentMeta, row)
		}
	}

	if len(PresentMeta) == 0 {
		GlobalContext.Database.MustExec("UPDATE http_document_meta SET is_active=0 WHERE service_id=?", service_id)
	} else {
		vars := []any{service_id}
		buff := bytes.NewBufferString("UPDATE http_document_meta SET is_active=0 WHERE service_id=? AND id NOT IN (")
		for i, v := range PresentMeta {
			if i == 0 {
				buff.WriteByte('?')
			} else {
				buff.WriteString(",?")
			}
			vars = append(vars, v.Id)
		}
		buff.WriteByte(')')
		GlobalContext.Database.MustExec(buff.String(), vars...)
	}

	for _, v := range meta {
		v := v.(map[string]any)
		p := TruncateText(v["property"].(string), 127)
		c := TruncateText(v["content"].(string), 127)
		is_present := false
		for _, m := range PresentMeta {
			if m.Property == p && m.Content == c {
				is_present = true
				break
			}
		}

		if is_present {
			continue
		}

		GlobalContext.Database.MustExec("INSERT INTO `http_document_meta`(`service_id`,`is_active`,`property`,`content`) VALUE(?,1,?,?)",
			service_id, p, c)
		PresentMeta = append(PresentMeta, HttpDocumentMetaRow{Property: p, Content: c})
	}
}
