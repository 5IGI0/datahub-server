package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/Masterminds/squirrel"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"golang.org/x/net/idna"
)

func _IndividualApiRows2Json(rows *sqlx.Rows) []map[string]any {
	Individuals := []map[string]any{}
	for rows.Next() {
		var row TableIndividual
		AssertError(rows.StructScan(&row)) // TODO: error
		Individuals = append(Individuals, row.ToMap())
	}

	GetIndividualsSources(Individuals)
	return Individuals
}

func _ApiIndividual_Strictness2Conds(strictness string, username string) squirrel.Sqlizer {
	switch strictness {
	case "permissive":
		return squirrel.Or{
			squirrel.Like{"san_user": alnumify(username) + "%"},
			squirrel.Like{"email": SQLEscapeStringLike(username) + "%"},
		}
	case "lenient":
		return squirrel.Or{
			squirrel.Like{"san_user": alnumify(username)},
			squirrel.Like{"email": SQLEscapeStringLike(username) + "+%"},
			squirrel.Like{"email": SQLEscapeStringLike(username) + "@%"},
		}
	case "moderate":
		return squirrel.Like{"email": SQLEscapeStringLike(username) + "%"}
	case "strict":
		return squirrel.Or{
			squirrel.Like{"email": SQLEscapeStringLike(username) + "+%"},
			squirrel.Like{"email": SQLEscapeStringLike(username) + "@%"},
		}
	case "exact":
		return squirrel.Like{"email": SQLEscapeStringLike(username) + "@%"}
	}

	return squirrel.Or{
		squirrel.Like{"san_user": alnumify(username) + "%"},
		squirrel.Like{"email": SQLEscapeStringLike(username) + "%"},
	}
}

func ApiIndividualByEmail(w http.ResponseWriter, r *http.Request) (any, int, string, error) {
	username := mux.Vars(r)["username"]
	domain, _ := idna.ToASCII(mux.Vars(r)["domain"])

	conds := squirrel.And{
		_ApiIndividual_Strictness2Conds(r.URL.Query().Get("strictness"), username),
		squirrel.Expr("rev_host LIKE REVERSE(?)", "%"+domain),
	}

	page, page_size := Req2Page(r)
	q, v, _ := squirrel.
		Select("`individuals`.*").
		From("`individual_emails`").
		Join("`individuals` ON `individuals`.`id`=`individual_emails`.`individual_id`").
		Where(conds).
		Limit(uint64(page_size)).
		Offset(uint64((page - 1) * page_size)).
		ToSql()

	rows, err := GlobalContext.Database.Queryx(q, v...)

	if err != nil {
		return nil, http.StatusInternalServerError, "SQL_ERROR", err
	}

	defer rows.Close()

	return _IndividualApiRows2Json(rows), 200, "", nil
}

func ApiIndividualAdd(w http.ResponseWriter, r *http.Request) (any, int, string, error) {
	var input []map[string]any

	{
		tmp, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(tmp, &input); err != nil {
			return nil, http.StatusBadRequest, "BAD_JSON", err
		}
	}

	rows := make([]TableIndividual, 0, len(input))

	for i := range input {
		var row TableIndividual
		if err := row.FromMap(input[i]); err != nil {
			return nil, http.StatusBadRequest, "INVALID_DATA", err
		}
		rows = append(rows, row)
	}

	InsertIndividuals(rows) // TODO: errors

	return nil, http.StatusCreated, "", nil
}
