package main

import "net/http"

func ApiStats(w http.ResponseWriter, r *http.Request) (any, int, string, error) {
	rows, err := GlobalContext.Database.Queryx(`
	SHOW TABLE STATUS WHERE name IN ('domains', 'http_services', 'ssl_certificates', 'dns_records')`)
	AssertError(err)

	{
		tables_stats := make(map[string]map[string]any)
		for rows.Next() {
			row := make(map[string]any)
			table_stats := make(map[string]any)
			AssertError(rows.MapScan(row))
			table_stats["count"] = row["Rows"]
			table_stats["size"] = row["Data_length"]
			tables_stats[string(row["Name"].([]uint8))] = table_stats
		}
		return tables_stats, 200, "", nil
	}
}
