package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

type ApiAddrInfoResponse struct {
	KnownDomains   []string `json:"known_domains"`
	KnownOldDomain []string `json:"known_old_domains"`
}

func ApiAddrInfo(w http.ResponseWriter, r *http.Request) (any, int, string, error) {
	var response ApiAddrInfoResponse
	addr := mux.Vars(r)["addr"]

	/* get known (old) domains */
	rows, err := GlobalContext.Database.Queryx(`
	SELECT domains.domain, dns_records.is_active FROM dns_records
	JOIN domains ON domains.id=dns_records.domain_id
	WHERE dns_records.addr LIKE ?
	`, SQLEscapeStringLike(addr))
	if err != nil {
		return response, http.StatusInternalServerError, "SQL_ERROR", err
	}
	defer rows.Close()

	for rows.Next() {
		var row DomainRow
		AssertError(rows.StructScan(&row))
		if row.IsActive != 0 { // actually points to dns_records' is_active
			response.KnownDomains = append(response.KnownDomains, row.Domain)
		} else {
			response.KnownOldDomain = append(response.KnownOldDomain, row.Domain)
		}
	}

	return response, 200, "", nil
}
