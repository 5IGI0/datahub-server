package main

import (
	"bytes"
	"fmt"
	"slices"
)

func InsertDomains(domains []string) int {
	known_domains := make([]string, 0, len(domains))
	query_str := bytes.NewBufferString("SELECT domain FROM domains WHERE domain IN (")
	vals := make([]any, 0, len(domains))
	for i, domain := range domains {
		if i == 0 {
			query_str.WriteByte('?')
		} else {
			query_str.WriteString(",?")
		}
		vals = append(vals, domain)
	}
	query_str.WriteByte(')')

	rows, err := GlobalContext.Database.Queryx(query_str.String(), vals...)
	AssertError(err)
	defer rows.Close()

	for rows.Next() {
		var domain string
		AssertError(rows.Scan(&domain))
		known_domains = append(known_domains, domain)
	}

	var total_added int = 0
	for _, domain := range domains {
		if !slices.Contains(known_domains, domain) {
			known_domains = append(known_domains, domain)
			_, err := GlobalContext.Database.Exec("INSERT IGNORE INTO domains(domain) VALUES(?)", domain)
			AssertError(err)
		}
	}

	return total_added
}

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
