package main

import (
	"encoding/json"

	"github.com/Masterminds/squirrel"
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
	if v, e := data["robots_txt"]; e {
		HttpServiceInsertRobotTxt(service_id, v.([]any))
	}
}

func HttpServiceInsertHeader(service_id int64, headers map[string]any) {
	var Headers []HttpHeaderRow
	var PresentHeaders []HttpHeaderRow

	if len(headers) == 0 {
		GlobalContext.Database.Exec("UPDATE http_headers SET is_active=0 WHERE service_id=?")
		return
	}

	/* parse */
	for k, v := range headers {
		var Header HttpHeaderRow

		Header.ServiceId = service_id
		Header.Value = TruncateText(v.(string), 127)
		Header.Key = TruncateText(k, 127)
		Header.HashId = Header.CompHashId()

		Headers = append(Headers, Header)
	}

	/* get present directives */
	{
		var conds squirrel.Or
		for _, h := range Headers {
			conds = append(conds, squirrel.Eq{
				"hash_id": h.HashId})
		}

		q, v, _ := squirrel.
			Select("id", "hash_id").
			From("http_headers").
			Where(conds).ToSql()

		rows, err := GlobalContext.Database.Queryx(q, v...)
		AssertError(err)
		defer rows.Close()

		for rows.Next() {
			var Header HttpHeaderRow
			AssertError(rows.StructScan(&Header))
			PresentHeaders = append(PresentHeaders, Header)
		}
	}

	if len(PresentHeaders) == 0 {
		GlobalContext.Database.MustExec("UPDATE http_headers SET is_active=0 WHERE service_id=?", service_id)
	} else {
		conds := squirrel.And{squirrel.Eq{"service_id": service_id}}

		for _, h := range PresentHeaders {
			conds = append(conds, squirrel.NotEq{"id": h.Id})
		}

		q, v := squirrel.
			Update("http_headers").
			Set("is_active", 0).
			Where(conds).
			MustSql()
		GlobalContext.Database.MustExec(q, v...)
	}

	for _, Header := range Headers {
		is_present := false
		for _, PHeader := range PresentHeaders {
			if PHeader.HashId == Header.HashId {
				is_present = true
				break
			}
		}

		if is_present {
			continue
		}

		q, v := squirrel.
			Insert("http_headers").SetMap(map[string]interface{}{
			"`service_id`": Header.ServiceId,
			"`is_active`":  1,
			"`key`":        Header.Key,
			"`value`":      Header.Value,
		}).MustSql()
		GlobalContext.Database.MustExec(q, v...)
		PresentHeaders = append(PresentHeaders, Header)
	}
}
func HttpServiceInsertDocumentMeta(service_id int64, meta []any) {
	var MetaList []HttpDocumentMetaRow
	var PresentMeta []HttpDocumentMetaRow

	if len(meta) == 0 {
		GlobalContext.Database.MustExec("UPDATE `http_document_meta` SET is_active=0 WHERE service_id=?",
			service_id)
		return
	}

	/* parse */
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

	/* get present directives */
	{
		var conds squirrel.Or
		for _, m := range MetaList {
			conds = append(conds, squirrel.Eq{
				"hash_id": m.HashId})
		}

		q, v, _ := squirrel.
			Select("id", "hash_id").
			From("http_robots_txt").
			Where(conds).ToSql()

		rows, err := GlobalContext.Database.Queryx(q, v...)
		AssertError(err)
		defer rows.Close()

		for rows.Next() {
			var Meta HttpDocumentMetaRow
			AssertError(rows.StructScan(&Meta))
			PresentMeta = append(PresentMeta, Meta)
		}
	}

	/* update is_actives */
	if len(PresentMeta) == 0 {
		GlobalContext.Database.MustExec("UPDATE `http_robots_txt` SET is_active=0 WHERE service_id=?",
			service_id)
	} else {
		conds := squirrel.And{squirrel.Eq{"service_id": service_id}}

		for _, m := range PresentMeta {
			conds = append(conds, squirrel.NotEq{"id": m.Id})
		}

		q, v := squirrel.
			Update("http_robots_txt").
			Set("is_active", 0).
			Where(conds).
			MustSql()
		GlobalContext.Database.MustExec(q, v...)
	}

	/* insert */
	for _, m := range MetaList {
		is_present := false
		for _, pm := range PresentMeta {
			if pm.HashId == m.HashId {
				is_present = true
				break
			}
		}
		if is_present {
			continue
		}

		q, v := squirrel.
			Insert("http_document_meta").SetMap(map[string]interface{}{
			"service_id": service_id,
			"is_active":  1,
			"property":   m.Property,
			"content":    m.Content,
		}).MustSql()
		GlobalContext.Database.MustExec(q, v...)
		PresentMeta = append(PresentMeta, m)
	}
}

func HttpServiceInsertRobotTxt(service_id int64, directives []any) {
	var Directives []HttpRobotsTxtRow
	var PresentDirectives []HttpRobotsTxtRow

	if len(directives) == 0 {
		GlobalContext.Database.MustExec("UPDATE `http_robots_txt` SET is_active=0 WHERE service_id=?",
			service_id)
		return
	}

	/* parse */
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

	/* get present directives */
	{
		var conds squirrel.Or
		for _, d := range Directives {
			conds = append(conds, squirrel.Eq{
				"hash_id": d.CompHashId()})
		}

		q, v, _ := squirrel.
			Select("id", "hash_id").
			From("http_robots_txt").
			Where(conds).ToSql()

		rows, err := GlobalContext.Database.Queryx(q, v...)
		AssertError(err)
		defer rows.Close()

		for rows.Next() {
			var Directive HttpRobotsTxtRow
			AssertError(rows.StructScan(&Directive))
			PresentDirectives = append(PresentDirectives, Directive)
		}
	}

	/* update is_actives */
	if len(PresentDirectives) == 0 {
		GlobalContext.Database.MustExec("UPDATE `http_robots_txt` SET is_active=0 WHERE service_id=?",
			service_id)
	} else {
		conds := squirrel.And{squirrel.Eq{"service_id": service_id}}

		for _, d := range PresentDirectives {
			conds = append(conds, squirrel.NotEq{"id": d.Id})
		}

		q, v := squirrel.
			Update("http_robots_txt").
			Set("is_active", 0).
			Where(conds).
			MustSql()
		GlobalContext.Database.MustExec(q, v...)
	}

	/* insert */
	for _, d := range Directives {
		is_present := false
		d.HashId = d.CompHashId()
		for _, pd := range PresentDirectives {
			if pd.HashId == d.HashId {
				is_present = true
				break
			}
		}
		if is_present {
			continue
		}

		q, v := squirrel.
			Insert("http_robots_txt").SetMap(map[string]interface{}{
			"service_id": service_id,
			"is_active":  1,
			"useragent":  d.UserAgent,
			"directive":  d.Directive,
			"value":      d.Value,
		}).MustSql()
		GlobalContext.Database.MustExec(q, v...)
		PresentDirectives = append(PresentDirectives, d)
	}
}
