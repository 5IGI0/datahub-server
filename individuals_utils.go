package main

import (
	"slices"
)

const (
	MAX_SQLX_PLACEHOLDERS = 250
)

func GetIndividualsSources(individuals []JsonIndividual) {
	ids := make([]int64, 0, len(individuals))

	for i := range individuals {
		ids = append(ids, individuals[i].Id)
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
				if individuals[y].Id == row.Id && !slices.Contains(individuals[y].Sources, UnicodeEscape(row.Source)) {
					individuals[y].Sources = append(individuals[y].Sources, UnicodeEscape(row.Source))
				}
			}
		}
	}
}

// TODO: remove duplicate
func _IndividualSortList(list []UnicodeEscape) []UnicodeEscape {
	output := make([]UnicodeEscape, len(list))
	copy(output, list)

	slices.SortFunc(output, func(a UnicodeEscape, b UnicodeEscape) int {
		for i := range []byte(a) {
			if len(b) <= i {
				return int([]byte(a)[i])
			}
			tmp := int([]byte(a)[i]) - int([]byte(b)[i])
			if tmp != 0 {
				return tmp
			}
		}

		if len(a) == len(b) {
			return 0
		}

		return -int([]byte(b)[len([]byte(b))-1])
	})

	return output
}
