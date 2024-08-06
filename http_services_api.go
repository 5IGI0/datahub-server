package main

import (
	"encoding/json"
	"errors"
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
			"active":      {Generator: ToggleCondGenerator, Field: "is_active", Default: "true"},
		},
	)

	if e != nil {
		return nil, http.StatusBadRequest, "BAD_REQUEST", e
	}

	page, page_size := Req2Page(r)
	q, v, _ := squirrel.
		Select("domain", "raw_result", "secure", "port", "is_active").
		From("http_services").
		Limit(uint64(page_size)).
		Offset(uint64((page - 1) * page_size)).
		Where(c).ToSql()

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

func ApiHttpServicesSearchByHeader(w http.ResponseWriter, r *http.Request) (any, int, string, error) {
	if !r.URL.Query().Has("key") {
		return nil, http.StatusBadRequest, "MISSING_PARAM", errors.New("the parameter `key` is mandatory")
	}

	c, e := GetQuery2SqlConds(
		r.URL.Query(),
		map[string]Query2SqlCond{
			/* domain-related */
			"status_code":    {Generator: EqualCondGenerator, Field: "`http_services`.`status_code`"},
			"port":           {Generator: EqualCondGenerator, Field: "`http_services`.`port`"},
			"title":          {Generator: BeginsWithCondGenerator, Field: "`http_services`.`title`"},
			"path":           {Generator: BeginsWithCondGenerator, Field: "`http_services`.`actuel_path`"},
			"secure":         {Generator: BoolCondGenerator, Field: "`http_services`.`secure`"},
			"domain":         {Generator: SubDomainCondGenerator, Field: "`http_services`.`rev_domain`"},
			"service_active": {Generator: ToggleCondGenerator, Field: "`http_services`.`is_active`", Default: "true"},

			/* header related */
			"active": {Generator: ToggleCondGenerator, Field: "`http_headers`.`is_active`", Default: "true"},
			"key":    {Generator: LikeCondGenerator, Field: "`http_headers`.`key`"},
			"val":    {Generator: BeginsWithCondGenerator, Field: "`http_headers`.`value`"},
		},
	)

	if e != nil {
		return nil, http.StatusBadRequest, "BAD_REQUEST", e
	}

	page, page_size := Req2Page(r)
	q, v, _ := squirrel.
		Select(
			"`http_services`.`domain`",
			"`http_services`.`raw_result`",
			"`http_services`.`secure`",
			"`http_services`.`port`",
			"`http_services`.`is_active`").
		From("`http_headers`").
		Limit(uint64(page_size)).
		Offset(uint64((page - 1) * page_size)).
		Join("`http_services` ON `http_services`.`id`=`http_headers`.`service_id`").
		Where(c).ToSql()

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

func ApiHttpServicesSearchByMeta(w http.ResponseWriter, r *http.Request) (any, int, string, error) {
	if !r.URL.Query().Has("property") {
		return nil, http.StatusBadRequest, "MISSING_PARAM", errors.New("the parameter `property` is mandatory")
	}

	c, e := GetQuery2SqlConds(
		r.URL.Query(),
		map[string]Query2SqlCond{
			/* domain-related */
			"status_code":    {Generator: EqualCondGenerator, Field: "`http_services`.`status_code`"},
			"port":           {Generator: EqualCondGenerator, Field: "`http_services`.`port`"},
			"title":          {Generator: BeginsWithCondGenerator, Field: "`http_services`.`title`"},
			"path":           {Generator: BeginsWithCondGenerator, Field: "`http_services`.`actuel_path`"},
			"secure":         {Generator: BoolCondGenerator, Field: "`http_services`.`secure`"},
			"domain":         {Generator: SubDomainCondGenerator, Field: "`http_services`.`rev_domain`"},
			"service_active": {Generator: ToggleCondGenerator, Field: "`http_services`.`is_active`", Default: "true"},

			/* header related */
			"active":   {Generator: ToggleCondGenerator, Field: "`http_document_meta`.`is_active`", Default: "true"},
			"property": {Generator: LikeCondGenerator, Field: "`http_document_meta`.`property`"},
			"content":  {Generator: BeginsWithCondGenerator, Field: "`http_document_meta`.`content`"},
		},
	)

	if e != nil {
		return nil, http.StatusBadRequest, "BAD_REQUEST", e
	}

	page, page_size := Req2Page(r)
	q, v, _ := squirrel.
		Select(
			"`http_services`.`domain`",
			"`http_services`.`raw_result`",
			"`http_services`.`secure`",
			"`http_services`.`port`",
			"`http_services`.`is_active`").
		From("`http_document_meta`").
		Limit(uint64(page_size)).
		Offset(uint64((page - 1) * page_size)).
		Join("`http_services` ON `http_services`.`id`=`http_document_meta`.`service_id`").
		Where(c).ToSql()

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

func ApiHttpServicesSearchByRobotsTxt(w http.ResponseWriter, r *http.Request) (any, int, string, error) {
	if !r.URL.Query().Has("directive") {
		return nil, http.StatusBadRequest, "MISSING_PARAM", errors.New("the parameter `directive` is mandatory")
	}

	c, e := GetQuery2SqlConds(
		r.URL.Query(),
		map[string]Query2SqlCond{
			/* domain-related */
			"status_code":    {Generator: EqualCondGenerator, Field: "`http_services`.`status_code`"},
			"port":           {Generator: EqualCondGenerator, Field: "`http_services`.`port`"},
			"title":          {Generator: BeginsWithCondGenerator, Field: "`http_services`.`title`"},
			"path":           {Generator: BeginsWithCondGenerator, Field: "`http_services`.`actuel_path`"},
			"secure":         {Generator: BoolCondGenerator, Field: "`http_services`.`secure`"},
			"domain":         {Generator: SubDomainCondGenerator, Field: "`http_services`.`rev_domain`"},
			"service_active": {Generator: ToggleCondGenerator, Field: "`http_services`.`is_active`", Default: "true"},

			/* header related */
			"active":    {Generator: ToggleCondGenerator, Field: "`http_robots_txt`.`is_active`", Default: "true"},
			"useragent": {Generator: BeginsWithCondGenerator, Field: "`http_robots_txt`.`useragent`"},
			"directive": {Generator: BeginsWithCondGenerator, Field: "`http_robots_txt`.`directive`"},
			"val":       {Generator: BeginsWithCondGenerator, Field: "`http_robots_txt`.`value`"},
		},
	)

	if e != nil {
		return nil, http.StatusBadRequest, "BAD_REQUEST", e
	}

	page, page_size := Req2Page(r)
	q, v, _ := squirrel.
		Select(
			"`http_services`.`domain`",
			"`http_services`.`raw_result`",
			"`http_services`.`secure`",
			"`http_services`.`port`",
			"`http_services`.`is_active`").
		From("`http_robots_txt`").
		Limit(uint64(page_size)).
		Offset(uint64((page - 1) * page_size)).
		Join("`http_services` ON `http_services`.`id`=`http_robots_txt`.`service_id`").
		Where(c).ToSql()

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

func ApiHttpServicesSearchByCert(w http.ResponseWriter, r *http.Request) (any, int, string, error) {
	c, e := GetQuery2SqlConds(
		r.URL.Query(),
		map[string]Query2SqlCond{
			/* domain-related */
			"status_code":    {Generator: EqualCondGenerator, Field: "`http_services`.`status_code`"},
			"port":           {Generator: EqualCondGenerator, Field: "`http_services`.`port`"},
			"title":          {Generator: BeginsWithCondGenerator, Field: "`http_services`.`title`"},
			"path":           {Generator: BeginsWithCondGenerator, Field: "`http_services`.`actuel_path`"},
			"secure":         {Generator: BoolCondGenerator, Field: "`http_services`.`secure`"},
			"domain":         {Generator: SubDomainCondGenerator, Field: "`http_services`.`rev_domain`"},
			"service_active": {Generator: ToggleCondGenerator, Field: "`http_services`.`is_active`", Default: "true"},

			/* header related */
			"issuer_name":  {Generator: BeginsWithCondGenerator, Field: "`ssl_certificates`.`issuer_name`"},
			"issuer_orga":  {Generator: BeginsWithCondGenerator, Field: "`ssl_certificates`.`issuer_orga`"},
			"subject_name": {Generator: BeginsWithCondGenerator, Field: "`ssl_certificates`.`subject_name`"},
			"subject_orga": {Generator: BeginsWithCondGenerator, Field: "`ssl_certificates`.`subject_orga`"},
		},
	)

	if e != nil {
		return nil, http.StatusBadRequest, "BAD_REQUEST", e
	}

	page, page_size := Req2Page(r)
	q, v, _ := squirrel.
		Select(
			"`http_services`.`domain`",
			"`http_services`.`raw_result`",
			"`http_services`.`secure`",
			"`http_services`.`port`",
			"`http_services`.`is_active`").
		From("`ssl_certificates`").
		Limit(uint64(page_size)).
		Offset(uint64((page - 1) * page_size)).
		Join("`http_services` ON `http_services`.`certificate_id`=`ssl_certificates`.`id`").
		Where(c).ToSql()

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
