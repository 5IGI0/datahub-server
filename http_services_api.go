package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type ApiHttpServiceResponse struct {
	Domain string         `json:"domain"`
	Data   map[string]any `json:"data"`
}

func ApiHttpServicesSearch(w http.ResponseWriter, r *http.Request) (any, int, string, error) {
	var vals []any
	query_buff := bytes.NewBufferString("SELECT domain, raw_result FROM http_services WHERE ")
	has_condition := false

	if r.URL.Query().Has("title") {
		vals = append(vals, SQLEscapeStringLike(r.URL.Query().Get("title"))+"%")
		query_buff.WriteString(" page_title LIKE ? ")
		has_condition = true
	}

	if r.URL.Query().Has("status_code") {
		query_buff.WriteString(Ternary(has_condition, " AND ", ""))
		vals = append(vals, r.URL.Query().Get("status_code"))
		query_buff.WriteString(" status_code=? ")
		has_condition = true
	}

	if r.URL.Query().Has("domain") {
		// TODO: idna
		query_buff.WriteString(Ternary(has_condition, " AND ", ""))
		vals = append(vals, "%"+SQLEscapeStringLike(r.URL.Query().Get("domain")))
		query_buff.WriteString(" rev_domain LIKE REVERSE(?) ")
		has_condition = true
	}

	if r.URL.Query().Has("port") {
		query_buff.WriteString(Ternary(has_condition, " AND ", ""))
		vals = append(vals, r.URL.Query().Get("port"))
		query_buff.WriteString(" port=? ")
		has_condition = true
	}

	if r.URL.Query().Has("secure") {
		if r.URL.Query().Get("secure") == "true" {
			query_buff.WriteString(" AND secure=1 ")
		} else {
			query_buff.WriteString(" AND secure=0 ")
		}
	}

	// NOTE: keep it at the end of conditions
	if !r.URL.Query().Has("allow_inactive") ||
		strings.ToLower(r.URL.Query().Get("allow_inactive")) == "false" {
		query_buff.WriteString(" AND is_active=1 ")
	}

	if !has_condition {
		return nil, http.StatusBadRequest, "NO_CONDITION", errors.New("please provide at least one condition")
	}

	rows, err := GlobalContext.Database.Queryx(query_buff.String(), vals...)
	if err != nil {
		return nil, http.StatusInternalServerError, "SQL_ERROR", err
	}
	defer rows.Close()

	var ret []ApiHttpServiceResponse

	for rows.Next() {
		var row HttpServiceRow
		AssertError(rows.StructScan(&row))
		json_tmp := ApiHttpServiceResponse{
			Domain: row.Domain,
		}
		json.Unmarshal([]byte(row.RawResult), &json_tmp.Data)
		ret = append(ret, json_tmp)
	}

	return ret, 200, "", nil
}
