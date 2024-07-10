package main

import (
	"crypto/sha1"
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"time"
)

type JsonIndividual struct {
	Id        int64           `json:"id"`
	Emails    []UnicodeEscape `json:"emails"`
	Usernames []UnicodeEscape `json:"usernames"`
	Realnames []UnicodeEscape `json:"realnames"`
	// in case we can't easily determine if it is a realname or an username
	Names     []UnicodeEscape `json:"names"`
	FirstSeen UnicodeEscape   `json:"first_seen"`
	LastSeen  UnicodeEscape   `json:"last_seen"`
	Sources   []UnicodeEscape
}

func (ind *JsonIndividual) FromRow(row *TableIndividual) {
	ind.Id = row.Id
	json.Unmarshal([]byte(row.Emails), &ind.Emails)
	json.Unmarshal([]byte(row.Usernames), &ind.Usernames)
	json.Unmarshal([]byte(row.Realnames), &ind.Realnames)
	json.Unmarshal([]byte(row.Names), &ind.Names)
	ind.FirstSeen = UnicodeEscape(time.Time(row.FirstSeen).Format(time.RFC3339))
	ind.LastSeen = UnicodeEscape(time.Time(row.LastSeen).Format(time.RFC3339))
	ind.Sources = append([]UnicodeEscape{}, row._sources...)
}

type TableIndividual struct {
	Id int64 `db:"id"`
	// json-encoded list.
	Emails    string `db:"emails"`
	_emails   []UnicodeEscape
	Usernames string `db:"usernames"`
	Realnames string `db:"realnames"`
	Names     string `db:"names"`
	HashId    string `db:"hash_id"`
	FirstSeen DBTime `db:"first_seen"`
	LastSeen  DBTime `db:"last_seen"`
	_sources  []UnicodeEscape
}

type DBTime time.Time

func (t DBTime) Value() (driver.Value, error) {
	if time.Time(t).IsZero() {
		return nil, nil
	}
	return time.Time(t).Format("2006-01-02T15:04:05"), nil
}

func (t *DBTime) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	tmp, err := time.Parse("2006-01-02 15:04:05", string(value.([]byte)))
	*t = DBTime(tmp)
	return err
}

func (row *TableIndividual) UpdateHash() {
	tmp := sha1.Sum([]byte(row.Emails + row.Usernames + row.Realnames + row.Names))
	row.HashId = hex.EncodeToString(tmp[:])
}

func (row *TableIndividual) Init() {
	row.Emails = "[]"
	row.Usernames = "[]"
	row.Realnames = "[]"
	row.Names = "[]"
	row.UpdateHash()
	row._sources = []UnicodeEscape{}
}

func (row *TableIndividual) FromJson(Individual *JsonIndividual) error {
	row.Init()

	if Individual.Emails != nil {
		for i := 0; i < len(Individual.Emails); i++ {
			Individual.Emails[i] = UnicodeEscape(SanitizeEmail(string(Individual.Emails[i])))
		}
		tmp, _ := json.Marshal(_IndividualSortList(Individual.Emails))
		row.Emails = string(tmp)
		row._emails = make([]UnicodeEscape, len(Individual.Emails))
		copy(row._emails, Individual.Emails)
	}

	if Individual.Names != nil {
		tmp, _ := json.Marshal(_IndividualSortList(Individual.Names))
		row.Names = string(tmp)
	}

	if Individual.Realnames != nil {
		tmp, _ := json.Marshal(_IndividualSortList(Individual.Realnames))
		row.Realnames = string(tmp)
	}

	if Individual.Usernames != nil {
		tmp, _ := json.Marshal(_IndividualSortList(Individual.Usernames))
		row.Usernames = string(tmp)
	}

	var err error
	tmp, err := time.Parse(time.RFC3339, string(Individual.FirstSeen))
	row.FirstSeen = DBTime(tmp)
	if err != nil {
		return err
	}
	tmp, err = time.Parse(time.RFC3339, string(Individual.LastSeen))
	row.LastSeen = DBTime(tmp)
	if err != nil {
		return err
	}

	row._sources = make([]UnicodeEscape, len(Individual.Sources))
	copy(row._sources, Individual.Sources)

	row.UpdateHash()

	return nil
}
