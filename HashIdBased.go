package main

import (
	"github.com/Masterminds/squirrel"
)

type HashIdBased interface {
	GetId() int64
	GetHashId() string
}

func GetPresentHashIdBasedRows[T HashIdBased](NewRows []T, table string) []T {
	var PresentRows []T

	{
		var conds squirrel.Or
		for _, r := range NewRows {
			conds = append(conds, squirrel.Eq{
				"hash_id": r.GetHashId()})
		}

		q, v := squirrel.
			Select("id", "hash_id").
			From(table).
			Where(conds).MustSql()

		rows, err := GlobalContext.Database.Queryx(q, v...)
		AssertError(err)
		defer rows.Close()

		for rows.Next() {
			var Row T
			AssertError(rows.StructScan(&Row))
			PresentRows = append(PresentRows, Row)
		}
	}

	return PresentRows
}

func UpdateActiveHashIdBasedRows[T HashIdBased](OldRows []T, table string, cond squirrel.Sqlizer) {
	/* set is_active = 0 for focused rows */
	{
		conds := squirrel.And{cond}

		for _, d := range OldRows {
			conds = append(conds, squirrel.NotEq{"id": d.GetId()})
		}

		q, v := squirrel.
			Update(table).
			Set("is_active", 0).
			Where(conds).
			MustSql()
		GlobalContext.Database.MustExec(q, v...)
	}

	if len(OldRows) == 0 {
		return
	}

	/* set is_active = 1 for focused rows */
	{
		conds := squirrel.Or{}

		for _, d := range OldRows {
			conds = append(conds, squirrel.Eq{"id": d.GetId()})
		}

		q, v := squirrel.
			Update(table).
			Set("is_active", 1).
			Where(conds).
			MustSql()
		GlobalContext.Database.MustExec(q, v...)
	}
}

func InsertHashIdBasedRows[T HashIdBased](
	NewRows []T,
	table string,
	cond squirrel.Sqlizer,
	OnNew func(r T) map[string]interface{},
	OnDup func(r T, index int64) map[string]interface{}) int64 {
	var PresentRows []T

	if len(NewRows) == 0 && cond != nil {
		q, v := squirrel.
			Update(table).Set("is_active", 0).Where(cond).MustSql()
		GlobalContext.Database.MustExec(q, v...)
		return 0
	}

	PresentRows = GetPresentHashIdBasedRows(NewRows, table)
	if cond != nil {
		UpdateActiveHashIdBasedRows(PresentRows, table, cond)
	}

	last_inserted_id := int64(0)
	for _, nr := range NewRows {
		is_present := false
		for _, pr := range PresentRows {
			if pr.GetHashId() == nr.GetHashId() {
				if OnDup != nil {
					SetMap := OnDup(nr, pr.GetId())

					if SetMap != nil {
						q, v := squirrel.
							Update(table).
							SetMap(SetMap).
							Where(squirrel.Eq{"hash_id": nr.GetHashId()}).MustSql()
						GlobalContext.Database.MustExec(q, v...)
					}
				}

				is_present = true
				break
			}
		}
		if is_present {
			continue
		}

		SetMap := OnNew(nr)
		if SetMap != nil {
			q, v := squirrel.
				Insert(table).SetMap(SetMap).MustSql()
			r := GlobalContext.Database.MustExec(q, v...)
			last_inserted_id, _ = r.LastInsertId()
		}
		PresentRows = append(PresentRows, nr)
	}

	return last_inserted_id
}
