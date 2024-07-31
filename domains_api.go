package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"golang.org/x/net/idna"
)

func ApiDomainAdd(w http.ResponseWriter, r *http.Request) (any, int, string, error) {
	var input map[string]any

	{
		tmp, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(tmp, &input); err != nil {
			return nil, http.StatusBadRequest, "BAD_JSON", err
		}
	}

	if err := DomainInsertScan(input); err != nil {
		return nil, http.StatusInternalServerError, "INTERNAL_ERROR", err
	}

	return nil, http.StatusCreated, "", nil
}

func ApiDomainSubs(w http.ResponseWriter, r *http.Request) (any, int, string, error) {
	domain, _ := idna.ToASCII(mux.Vars(r)["domain"])

	ret := make([]string, 0)

	rows, err := GlobalContext.Database.Query("SELECT domain FROM domains WHERE rev_domain LIKE REVERSE(?)",
		"%."+SQLEscapeStringLike(domain))

	if err != nil {
		return nil, http.StatusInternalServerError, "SQL_ERROR", err
	}
	defer rows.Close()

	for rows.Next() {
		var tmp string
		rows.Scan(&tmp)
		ret = append(ret, tmp)
	}

	return ret, 200, "", nil
}

func ApiDomainScan(w http.ResponseWriter, r *http.Request) (any, int, string, error) {
	domain, _ := idna.ToASCII(mux.Vars(r)["domain"])

	var domain_id int64 = 0
	GlobalContext.Database.Get(&domain_id, "SELECT id FROM domains WHERE domain=LOWER(?)", domain)

	if domain_id == 0 {
		return nil, http.StatusNotFound, "DOMAIN_NOT_FOUND", fmt.Errorf("`%s` not found", domain)
	}

	var raw_result string = ""

	GlobalContext.Database.Get(&raw_result,
		"SELECT raw_result FROM domain_scan_archives WHERE domain_id=? ORDER BY id DESC LIMIT 1", domain_id)

	if raw_result == "" {
		return nil, http.StatusNotFound, "SCAN_NOT_FOUND", fmt.Errorf("no scan available for `%s`", domain)
	}

	var ret any

	json.Unmarshal([]byte(raw_result), &ret)

	return ret, 200, "", nil
}

func ApiDomainsOutdated(w http.ResponseWriter, r *http.Request) (any, int, string, error) {
	ret := make([]string, 0, 500)

	rows, err := GlobalContext.Database.Query(`
	SELECT domain FROM domains
	WHERE last_check IS NULL LIMIT 100`)
	if err != nil {
		return nil, http.StatusInternalServerError, "SQL_ERROR", err
	}
	defer rows.Close()

	for rows.Next() {
		var tmp string
		AssertError(rows.Scan(&tmp))
		ret = append(ret, tmp)
	}

	if len(ret) != 0 {
		return ret, 200, "", nil
	}

	// if nothing found, then scan old scanned domains
	rows, err = GlobalContext.Database.Query(`
	SELECT domain FROM domains
	WHERE last_check < SUBTIME(NOW(), '01 00 00:00:00') LIMIT 100`)
	if err != nil {
		return nil, http.StatusInternalServerError, "SQL_ERROR", err
	}
	defer rows.Close()

	for rows.Next() {
		var tmp string
		AssertError(rows.Scan(&tmp))
		ret = append(ret, tmp)
	}

	return ret, 200, "", nil
}
