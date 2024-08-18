package main

import "fmt"

func __InsertDiscourseInstances(instances []DiscourseInstanceRow) {
	InsertHashIdBasedRows(instances, "discourse_instances", nil,
		func(r DiscourseInstanceRow) map[string]any {
			return map[string]any{
				"host":           r.Host,
				"secure":         r.Secure,
				"root":           r.Root,
				"title":          r.Title,
				"description":    r.Description,
				"raw_data":       r.RawData,
				"login_required": r.LoginRequired,
			}
		}, nil)
}

func TaskHttpServices2Discourses() {
	total_discourses := 0
	rows, err := GlobalContext.Database.Queryx(
		"SELECT DISTINCT domain FROM http_robots_txt JOIN `http_services` ON http_services.id=http_robots_txt.service_id WHERE `directive` LIKE 'disallow' AND `value`='/*?*api_key*' AND `http_services`.`secure`=1;")
	AssertError(err)
	defer rows.Close()

	var instances []DiscourseInstanceRow
	for rows.Next() {
		var row HttpServiceRow
		var instance_row DiscourseInstanceRow

		AssertError(rows.StructScan(&row))
		instance_row.Root = ""
		instance_row.Secure = 1
		instance_row.Host = row.Domain
		instance_row.Title = "[NOT YET SCRAPED]"
		instance_row.Description = "[NOT YET SCRAPED]"
		instance_row.RawData = "null"
		instance_row.LoginRequired = 0
		instances = append(instances, instance_row)

		if len(instances) == MAX_SQLX_PLACEHOLDERS/4 {
			__InsertDiscourseInstances(instances)
			total_discourses += MAX_SQLX_PLACEHOLDERS / 4
			fmt.Print("[individual_http_services_2_discourses] Processed ", total_discourses, " discourses\r")
			instances = instances[:0]
		}
	}

	if len(instances) != 0 {
		__InsertDiscourseInstances(instances)
	}
}
