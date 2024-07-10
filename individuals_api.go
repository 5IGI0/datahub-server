package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"
)

// TODO: pagination

// TODO: ?subdomain=false
func ApiIndividualByDomain(w http.ResponseWriter, r *http.Request) (any, int, string, error) {
	// TODO: encode utf-8 chars to valid-domain
	domain := reverse_str(mux.Vars(r)["domain"])

	// TODO: err
	rows, _ := GlobalContext.Database.Queryx(`
	SELECT individuals.* FROM individual_emails
	JOIN individuals ON individuals.id=individual_emails.individual_id
	WHERE individual_emails.rev_host LIKE ? OR individual_emails.rev_host LIKE ?`,
		SQLEscapeStringLike(domain)+"%", SQLEscapeStringLike(domain))
	defer rows.Close()

	var Individuals []JsonIndividual
	for rows.Next() {
		var IndividualRow TableIndividual
		var IndividualJson JsonIndividual
		rows.StructScan(&IndividualRow) // TODO: error
		IndividualJson.FromRow(&IndividualRow)
		Individuals = append(Individuals, IndividualJson)
	}

	GetIndividualsSources(Individuals)

	return Individuals, 200, "", nil
}

// ?strict=true
func ApiIndividualByUsername(w http.ResponseWriter, r *http.Request) (any, int, string, error) {
	// TODO: encode utf-8 chars to valid-domain
	username := mux.Vars(r)["username"]

	// TODO: err
	rows, _ := GlobalContext.Database.Queryx(`
	SELECT individuals.* FROM individual_emails
	JOIN individuals ON individuals.id=individual_emails.individual_id
	WHERE individual_emails.san_user LIKE ? OR individual_emails.san_user LIKE ?`,
		SQLEscapeStringLike(username)+"%", SQLEscapeStringLike(username))
	defer rows.Close()

	var Individuals []JsonIndividual
	for rows.Next() {
		var IndividualRow TableIndividual
		var IndividualJson JsonIndividual
		rows.StructScan(&IndividualRow) // TODO: error
		IndividualJson.FromRow(&IndividualRow)
		Individuals = append(Individuals, IndividualJson)
	}

	GetIndividualsSources(Individuals)

	return Individuals, 200, "", nil
}

// TODO: ?strict=true
func ApiIndividualByEmail(w http.ResponseWriter, r *http.Request) (any, int, string, error) {
	// TODO: encode utf-8 chars to valid-domain
	username := mux.Vars(r)["username"]
	domain := mux.Vars(r)["domain"]

	// TODO: err
	rows, _ := GlobalContext.Database.Queryx(`
	SELECT individuals.* FROM individual_emails
	JOIN individuals ON individuals.id=individual_emails.individual_id
	WHERE individual_emails.san_user=LOWER(?) AND individual_emails.rev_host=LOWER(?)`,
		alnumify(username), reverse_str(domain))
	defer rows.Close()

	var Individuals []JsonIndividual
	for rows.Next() {
		var IndividualRow TableIndividual
		var IndividualJson JsonIndividual
		rows.StructScan(&IndividualRow) // TODO: error
		IndividualJson.FromRow(&IndividualRow)
		Individuals = append(Individuals, IndividualJson)
	}

	GetIndividualsSources(Individuals)

	return Individuals, 200, "", nil
}

func ApiIndividualAdd(w http.ResponseWriter, r *http.Request) (any, int, string, error) {
	var input []JsonIndividual

	{
		tmp, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(tmp, &input); err != nil {
			return nil, http.StatusBadRequest, "BAD_JSON", err
		}
	}

	rows := make([]TableIndividual, 0, len(input))

	for i := range input {
		var row TableIndividual
		if err := row.FromJson(&input[i]); err != nil {
			return nil, http.StatusBadRequest, "INVALID_DATA", err
		}
		rows = append(rows, row)
	}

	InsertIndividuals(rows) // TODO: errors

	return nil, http.StatusCreated, "", nil
}
