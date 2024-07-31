package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"slices"
	"time"
)

type TableIndividual struct {
	Id int64 `db:"id"`
	// json-encoded data.
	Data      string
	HashId    string `db:"hash_id"`
	FirstSeen DBTime `db:"first_seen"`
	LastSeen  DBTime `db:"last_seen"`
	_emails   []string
	_sources  []string
}

func (row *TableIndividual) UpdateHash() {
	tmp := sha1.Sum([]byte(row.Data))
	row.HashId = hex.EncodeToString(tmp[:])
}

func (row *TableIndividual) Init() {
	row.Data = "{}"
	row.UpdateHash()
	row._sources = []string{}
}

func (row *TableIndividual) FromMap(Individual map[string]any) error {
	row.Init()

	if str, ok := Individual["first_seen"].(string); ok {
		tmp, err := time.Parse(time.RFC3339, str)
		row.FirstSeen = DBTime(tmp)
		if err != nil {
			return err
		}
	}

	if str, ok := Individual["last_seen"].(string); ok {
		tmp, err := time.Parse(time.RFC3339, str)
		row.LastSeen = DBTime(tmp)
		if err != nil {
			return err
		}
	}

	row._emails, _ = JsonAny2StringList(Individual["emails"])
	row._sources, _ = JsonAny2StringList(Individual["sources"])

	if len(row._emails) == 0 {
		return errors.New("individual has no searchable field")
	}

	if len(row._sources) == 0 {
		return errors.New("individual has no source")
	}

	map_copy := make(map[string]any)

	for k, v := range Individual {
		if !slices.Contains([]string{
			// things that should be ignored (for hash consistence)
			"sources",
			"first_seen",
			"last_seen",
		}, k) {
			map_copy[k] = v
		}
	}

	if tmp, err := json.Marshal(JsonSanitize(map_copy)); err != nil {
		return err
	} else {
		row.Data = string(tmp)
	}

	row.UpdateHash()

	return nil
}

func (row *TableIndividual) ToMap() map[string]any {
	ret := make(map[string]any)

	json.Unmarshal([]byte(row.Data), &ret)
	ret["first_seen"] = any(row.FirstSeen)
	ret["last_seen"] = any(row.LastSeen)
	ret[INDIVIDUAL_ID_KEY] = row.Id

	return ret
}
