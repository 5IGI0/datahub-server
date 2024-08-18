package main

import (
	"encoding/json"
	"fmt"
)

func TaskDiscoursePosts2Domains() {
	domains := make([]string, 0, MAX_SQLX_PLACEHOLDERS)
	var total_domain = 0

	rows, err := GlobalContext.Database.Queryx(
		"SELECT raw_data FROM `discourse_posts`")
	AssertError(err)
	defer rows.Close()

	for rows.Next() {
		var row DiscoursePostRow
		var tmp struct {
			Links []struct {
				Internal bool   `json:"internal"`
				Url      string `json:"url"`
			} `json:"link_counts"`
		}
		AssertError(rows.StructScan(&row))
		AssertError(json.Unmarshal([]byte(row.RawData), &tmp))

		for _, link := range tmp.Links {
			if domain, e := ExtractDomainFromLink(link.Url); e && !link.Internal {
				domains = append(domains, domain)
				if len(domains) == MAX_SQLX_PLACEHOLDERS {
					total_domain += MAX_SQLX_PLACEHOLDERS
					InsertDomains(domains)
					domains = domains[:0]
					fmt.Print("[individual_discourse_posts_2_domains] Processed ", total_domain, " domains\r")
				}
			}
		}
	}

	if len(domains) != 0 {
		InsertDomains(domains)
	}
}
