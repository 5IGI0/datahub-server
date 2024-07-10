package main

import (
	"slices"
	"time"
)

const (
	_INSERT_INDIVIDUAL_BATCH = 500
)

func InsertIndividuals(individuals []TableIndividual) {
	IndividualsMap := make(map[string]*TableIndividual)

	// unduplicate individuals
	for i := range individuals {
		individuals[i].UpdateHash()
		indptr := IndividualsMap[individuals[i].HashId]

		if indptr == nil {
			copy := individuals[i]
			indptr = &copy
			IndividualsMap[individuals[i].HashId] = &copy
		}

		if time.Time(indptr.LastSeen).Before(time.Time(individuals[i].LastSeen)) {
			indptr.LastSeen = individuals[i].LastSeen
		}

		if time.Time(indptr.FirstSeen).After(time.Time(individuals[i].FirstSeen)) {
			indptr.FirstSeen = individuals[i].FirstSeen
		}

		for _, source := range individuals[i]._sources {
			if !slices.Contains(indptr._sources, source) {
				indptr._sources = append(indptr._sources, source)
			}
		}
	}

	IndividualsBatch := make([]TableIndividual, 0, _INSERT_INDIVIDUAL_BATCH)
	for _, ind := range IndividualsMap {
		IndividualsBatch = append(IndividualsBatch, *ind)

		if len(IndividualsBatch) == _INSERT_INDIVIDUAL_BATCH {
			_InsertIndividuals_worker(IndividualsBatch)
			IndividualsBatch = IndividualsBatch[:0]
		}
	}

	if len(IndividualsBatch) != 0 {
		_InsertIndividuals_worker(IndividualsBatch)
	}
}

func _IndividualsPopulateHashId2Id(individuals []TableIndividual, HashId2Id map[string]int64) {
	var hash_ids []any
	// TODO: string builder
	query := "SELECT id, hash_id FROM individuals WHERE hash_id IN ("
	for i := range individuals {
		if i != 0 {
			query += ",?"
		} else {
			query += "?"
		}
		individuals[i].UpdateHash()
		hash_ids = append(hash_ids, individuals[i].HashId)
	}
	query += ")"

	rows, err := GlobalContext.Database.Queryx(query, hash_ids...)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var row struct {
			Id     int64  `db:"id"`
			HashId string `db:"hash_id"`
		}
		rows.StructScan(&row) // TODO: error
		HashId2Id[row.HashId] = row.Id
	}
}

func _IndividualsPopulateId2Sources(individuals []TableIndividual, Id2Sources map[int64][]string, IdMap map[string]int64) {
	var ids []any

	query := "SELECT individual_id, source FROM individual_sources WHERE individual_id IN ("
	for i := range individuals {
		if i != 0 {
			query += ",?"
		} else {
			query += "?"
		}
		individuals[i].UpdateHash()
		ids = append(ids, IdMap[individuals[i].HashId])
	}
	query += ")"

	rows, err := GlobalContext.Database.Queryx(query, ids...)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var row struct {
			Id     int64  `db:"individual_id"`
			Source string `db:"hash_id"`
		}
		rows.StructScan(&row) // TODO: error
		Id2Sources[row.Id] = append(Id2Sources[row.Id], row.Source)
	}
}

func _InsertIndividuals_worker(individuals []TableIndividual) {
	// to not insert already inserted rows
	hashId2Id := make(map[string]int64)
	_IndividualsPopulateHashId2Id(individuals, hashId2Id)

	for i := range individuals {

		// i do this to avoid as many unused ids as possible.
		if hashId2Id[individuals[i].HashId] == 0 {
			// on duplicate key update -> avoid atomicity issues
			GlobalContext.Database.MustExec(`
			INSERT INTO individuals (emails,usernames,realnames,names,first_seen,last_seen)
			VALUES (?,?,?,?,?,?)
			ON DUPLICATE KEY UPDATE
			first_seen=LEAST(first_seen, VALUES(first_seen)), last_seen=GREATEST(last_seen, VALUES(last_seen))`,
				individuals[i].Emails, individuals[i].Usernames, individuals[i].Realnames,
				individuals[i].Names, individuals[i].FirstSeen, individuals[i].LastSeen)
		} else {
			GlobalContext.Database.MustExec(`
			UPDATE individuals
			SET first_seen=LEAST(first_seen, ?), last_seen=GREATEST(last_seen, ?)
			WHERE id=?`,
				individuals[i].FirstSeen, individuals[i].LastSeen, hashId2Id[individuals[i].HashId])
		}
	}

	newHashId2id := make(map[string]int64)
	_IndividualsPopulateHashId2Id(individuals, newHashId2id)
	Id2Sources := make(map[int64][]string)
	_IndividualsPopulateId2Sources(individuals, Id2Sources, newHashId2id)

	for i := range individuals {
		id := newHashId2id[individuals[i].HashId]
		if id == 0 {
			continue
		}

		// if new, add meta in the meta' tables
		if hashId2Id[individuals[i].HashId] == 0 {
			for _, email := range individuals[i]._emails {
				GlobalContext.Database.MustExec(`
				INSERT IGNORE INTO individual_emails(email,individual_id) VALUES(?,?)`,
					email, id)
			}
		}

		individual_sources := Id2Sources[hashId2Id[individuals[i].HashId]]
		for _, source := range individuals[i]._sources {
			if !slices.Contains(individual_sources, string(source)) {
				GlobalContext.Database.MustExec(`
					INSERT IGNORE INTO individual_sources(source,individual_id) VALUES(?,?)`,
					source, id)
			}
		}
	}
}
