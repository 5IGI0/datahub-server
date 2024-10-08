package main

import (
	"fmt"
)

func TaskIndividualEmails2Domains() {
	domains := make([]string, 0, MAX_SQLX_PLACEHOLDERS)
	var total_domain = 0

	rows, err := GlobalContext.Database.Queryx(
		"SELECT DISTINCT REVERSE(rev_host) AS domain FROM individual_emails")
	AssertError(err)
	defer rows.Close()

	for rows.Next() {
		var domain string
		AssertError(rows.Scan(&domain))
		domains = append(domains, domain)
		if len(domains) == MAX_SQLX_PLACEHOLDERS {
			total_domain += MAX_SQLX_PLACEHOLDERS
			InsertDomains(domains)
			domains = domains[:0]
			fmt.Print("[individual_emails_2_domains] Processed ", total_domain, " domains\r")
		}
	}

	if len(domains) != 0 {
		InsertDomains(domains)
	}
}
