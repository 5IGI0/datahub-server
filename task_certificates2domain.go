package main

import "fmt"

func TaskCertificate2Domain() {
	domains := make([]string, 0, MAX_SQLX_PLACEHOLDERS)
	var total_domain = 0

	rows, err := GlobalContext.Database.Queryx(
		"SELECT DISTINCT domain FROM ssl_certificate_dns_names WHERE domain NOT LIKE '*%'")
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
			fmt.Print("[individual_certificates_2_domains] Processed ", total_domain, " domains\r")
		}
	}

	if len(domains) != 0 {
		InsertDomains(domains)
	}
}
