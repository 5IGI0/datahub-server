package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"golang.org/x/net/idna"
)

// TODO: pagination

func _IndividualApiRows2Json(rows *sqlx.Rows) []map[string]any {
	Individuals := []map[string]any{}
	for rows.Next() {
		var row TableIndividual
		rows.StructScan(&row) // TODO: error
		Individuals = append(Individuals, row.ToMap())
	}

	GetIndividualsSources(Individuals)
	return Individuals
}

func ApiIndividualByDomain(w http.ResponseWriter, r *http.Request) (any, int, string, error) {
	domain, _ := idna.ToASCII(mux.Vars(r)["domain"])
	domain = SQLEscapeStringLike(domain)

	if r.URL.Query().Get("subdomain") != "false" {
		domain = "%" + domain
	}

	rows, err := GlobalContext.Database.Queryx(`
	SELECT individuals.* FROM individual_emails
	JOIN individuals ON individuals.id=individual_emails.individual_id
	WHERE individual_emails.rev_host LIKE REVERSE(?)`,
		SQLEscapeStringLike(domain))

	if err != nil {
		return nil, http.StatusInternalServerError, "SQL_ERROR", err
	}

	defer rows.Close()

	return _IndividualApiRows2Json(rows), 200, "", nil
}

func _ApiIndividual_Strictness2Conds(strictness string, username string) (string, []any) {
	switch strictness {
	case "permissive":
		return "email LIKE ? OR san_user LIKE ?", append([]any{},
			SQLEscapeStringLike(username)+"%", alnumify(username)+"%")
	case "lenient":
		return "email LIKE ? OR email LIKE ? OR san_user = LOWER(?)", append([]any{},
			SQLEscapeStringLike(username)+"+%",
			SQLEscapeStringLike(username)+"@%",
			alnumify(username))
	case "moderate":
		return "email LIKE ?", append([]any{},
			SQLEscapeStringLike(username)+"%")
	case "strict":
		return "email LIKE ? OR email LIKE ?", append([]any{},
			SQLEscapeStringLike(username)+"+%",
			SQLEscapeStringLike(username)+"@%")
	case "exact":
		return "email LIKE ?", append([]any{}, SQLEscapeStringLike(username)+"@%")
	}

	return "email LIKE ? OR san_user LIKE ?", append([]any{},
		SQLEscapeStringLike(username)+"%", alnumify(username)+"%")
}

func ApiIndividualByUsername(w http.ResponseWriter, r *http.Request) (any, int, string, error) {
	username := mux.Vars(r)["username"]

	usr_query, usr_vals := _ApiIndividual_Strictness2Conds(r.URL.Query().Get("strictness"), username)

	rows, err := GlobalContext.Database.Queryx(`
	SELECT individuals.* FROM individual_emails
	JOIN individuals ON individuals.id=individual_emails.individual_id
	WHERE `+usr_query,
		usr_vals...)

	if err != nil {
		return nil, http.StatusInternalServerError, "SQL_ERROR", err
	}

	defer rows.Close()

	return _IndividualApiRows2Json(rows), 200, "", nil
}

func ApiIndividualByEmail(w http.ResponseWriter, r *http.Request) (any, int, string, error) {
	username := mux.Vars(r)["username"]
	domain, _ := idna.ToASCII(mux.Vars(r)["domain"])

	query, vals := _ApiIndividual_Strictness2Conds(r.URL.Query().Get("strictness"), username)
	query = "(" + query + ") AND rev_host LIKE REVERSE(?)"
	if r.URL.Query().Get("subdomain") != "false" {
		vals = append(vals, "%"+domain)
	}

	rows, err := GlobalContext.Database.Queryx(`
	SELECT individuals.* FROM individual_emails
	JOIN individuals ON individuals.id=individual_emails.individual_id
	WHERE `+query, vals...)

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
