package main

import "database/sql"

type HttpServiceRow struct {
	Id         int64          `db:"id"`
	DomainId   int64          `db:"domain"`
	IsActive   int8           `db:"is_active"`
	Domain     string         `db:"domain"`
	RevDomain  string         `db:"rev_domain"`
	Secure     int8           `db:"secure"` // https
	Port       uint16         `db:"port"`
	PageTitle  sql.NullString `db:"page_title"`
	StatusCode uint16         `db:"status_code"`
	ActualPath string         `db:"actual_path"`
	RawStatus  string         `db:"raw_result"`
}

type HttpDocumentMetaRow struct {
	Id        int64  `db:"id"`
	ServiceId int64  `db:"service_id"`
	IsActive  int8   `db:"is_active"`
	Property  string `db:"property"`
	Content   string `db:"content"`
}

type HttpHeaderRow struct {
	Id        int64  `db:"id"`
	ServiceId int64  `db:"service_id"`
	IsActive  int8   `db:"is_active"`
	Key       string `db:"key"`
	Value     string `db:"value"`
}
