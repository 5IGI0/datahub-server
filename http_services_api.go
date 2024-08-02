package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Masterminds/squirrel"
)

type ApiHttpServiceResponse struct {
	Domain   string         `json:"domain"`
	Secure   int8           `json:"secure"`
	Port     uint16         `json:"port"`
	IsActive int8           `json:"is_active"`
	Data     map[string]any `json:"data"`
}

func ApiHttpServicesSearch(w http.ResponseWriter, r *http.Request) (any, int, string, error) {
	c, e := GetQuery2SqlConds(
		r.URL.Query(),
		map[string]Query2SqlCond{
			"status_code": {Generator: EqualCondGenerator},
			"port":        {Generator: EqualCondGenerator},
			"title":       {Generator: BeginsWithCondGenerator},
			"path":        {Generator: BeginsWithCondGenerator, Field: "actual_path"},
			"secure":      {Generator: BoolCondGenerator},
			"domain":      {Generator: SubDomainCondGenerator, Field: "rev_domain"},
			"active":      {Generator: ToggleCondGenerator, Field: "is_active"},
		},
	)

	if e != nil {
		return nil, http.StatusBadRequest, "BAD_REQUEST", e
	}

	q, v, _ := squirrel.
		Select("domain", "raw_result", "secure", "port", "is_active").
		From("http_services").Where(c).ToSql()

	log.Println(q, v)
	rows, err := GlobalContext.Database.Queryx(q, v...)
	if err != nil {
		return nil, http.StatusInternalServerError, "SQL_ERROR", err
	}
	defer rows.Close()

	var ret []ApiHttpServiceResponse

	for rows.Next() {
		var row HttpServiceRow
		AssertError(rows.StructScan(&row))
		json_tmp := ApiHttpServiceResponse{
			Domain:   row.Domain,
			Secure:   row.Secure,
			Port:     row.Port,
			IsActive: row.IsActive,
		}
		json.Unmarshal([]byte(row.RawResult), &json_tmp.Data)
		ret = append(ret, json_tmp)
	}

	return ret, 200, "", nil
}
