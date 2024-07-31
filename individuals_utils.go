package main

import (
	"slices"
)

func GetIndividualsSources(individuals []map[string]any) {
	ids := make([]int64, 0, len(individuals))

	for i := range individuals {
		ids = append(ids, individuals[i][INDIVIDUAL_ID_KEY].(int64))
	}

	for i := 0; i < len(ids); i += MAX_SQLX_PLACEHOLDERS {
		num := MAX_SQLX_PLACEHOLDERS
		if len(ids)-i < MAX_SQLX_PLACEHOLDERS {
			num = len(ids) - i
		}

		// TODO: string builder
		placeholders := make([]any, 0, num)
		query := `SELECT individual_id, source FROM individual_sources
		WHERE individual_id IN (`
		for y := 0; y < num; y++ {
			if y != 0 {
				query += ",?"
			} else {
				query += "?"
			}
			placeholders = append(placeholders, ids[i+y])
		}
		query += ")"

		rows, _ := GlobalContext.Database.Queryx(query, placeholders...) // TODO: error
		defer rows.Close()

		for rows.Next() {
			var row struct {
				Id     int64  `db:"individual_id"`
				Source string `db:"source"`
			}

			rows.StructScan(&row) // TODO: error

			for y := range individuals {
				sources, _ := JsonAny2StringList(individuals[y]["sources"])
				if individuals[y]["sources"] == nil {
					individuals[y]["sources"] = []any{}
				}

				if individuals[y][INDIVIDUAL_ID_KEY] == row.Id && !slices.Contains(sources, row.Source) {
					individuals[y]["sources"] = append(individuals[y]["sources"].([]any), any(row.Source))
				}
			}
		}
	}
}
