package main

import (
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"fmt"
)

type HttpServiceRow struct {
	Id         int64          `db:"id"`
	DomainId   int64          `db:"domain_id"`
	IsActive   int8           `db:"is_active"`
	Domain     string         `db:"domain"`
	RevDomain  string         `db:"rev_domain"`
	Secure     int8           `db:"secure"` // https
	Port       uint16         `db:"port"`
	PageTitle  sql.NullString `db:"page_title"`
	StatusCode uint16         `db:"status_code"`
	ActualPath string         `db:"actual_path"`
	RawResult  string         `db:"raw_result"`
}

type HttpDocumentMetaRow struct {
	Id        int64  `db:"id"`
	ServiceId int64  `db:"service_id"`
	IsActive  int8   `db:"is_active"`
	Property  string `db:"property"`
	Content   string `db:"content"`
	HashId    string `db:"hash_id"`
}

func (r HttpDocumentMetaRow) GetId() int64 { return r.Id }
func (r HttpDocumentMetaRow) GetHashId() string {
	if r.HashId == "" {
		return r.CompHashId()
	}
	return r.HashId
}
func (r HttpDocumentMetaRow) CompHashId() string {
	h := sha1.Sum([]byte(
		fmt.Sprintf(
			"%v:%v:%v",
			r.ServiceId,
			r.Property,
			r.Content)))
	return hex.EncodeToString(h[:])
}

type HttpHeaderRow struct {
	Id        int64  `db:"id"`
	ServiceId int64  `db:"service_id"`
	IsActive  int8   `db:"is_active"`
	Key       string `db:"key"`
	Value     string `db:"value"`
	HashId    string `db:"hash_id"`
}

func (r HttpHeaderRow) GetId() int64 { return r.Id }
func (r HttpHeaderRow) GetHashId() string {
	if r.HashId == "" {
		return r.CompHashId()
	}
	return r.HashId
}
func (r HttpHeaderRow) CompHashId() string {
	h := sha1.Sum([]byte(
		fmt.Sprintf(
			"%v:%v:%v",
			r.ServiceId,
			r.Key,
			r.Value)))
	return hex.EncodeToString(h[:])
}

type HttpRobotsTxtRow struct {
	Id        int64  `db:"id"`
	ServiceId int64  `db:"service_id"`
	IsActive  int8   `db:"is_active"`
	UserAgent string `db:"useragent"`
	Directive string `db:"directive"`
	Value     string `db:"value"`
	HashId    string `db:"hash_id"`
}

func (r HttpRobotsTxtRow) GetId() int64 { return r.Id }
func (r HttpRobotsTxtRow) GetHashId() string {
	if r.HashId == "" {
		return r.CompHashId()
	}
	return r.HashId
}
func (r HttpRobotsTxtRow) CompHashId() string {
	h := sha1.Sum([]byte(
		fmt.Sprintf(
			"%v:%v:%v:%v",
			r.ServiceId,
			r.UserAgent,
			r.Directive,
			r.Value)))
	return hex.EncodeToString(h[:])
}

type MatrixClientVersionRow struct {
	Id        int64  `db:"id"`
	ServiceId int64  `db:"service_id"`
	IsActive  int8   `db:"is_active"`
	Version   string `db:"version"`
	HashId    string `db:"hash_id"`
}

func (r MatrixClientVersionRow) GetId() int64 { return r.Id }
func (r MatrixClientVersionRow) GetHashId() string {
	if r.HashId == "" {
		return r.CompHashId()
	}
	return r.HashId
}
func (r MatrixClientVersionRow) CompHashId() string {
	h := sha1.Sum([]byte(
		fmt.Sprintf(
			"%v:%v",
			r.ServiceId,
			r.Version)))
	return hex.EncodeToString(h[:])
}

type MatrixClientFeatureRow struct {
	Id        int64  `db:"id"`
	ServiceId int64  `db:"service_id"`
	IsActive  int8   `db:"is_active"`
	Feature   string `db:"version"`
	HashId    string `db:"hash_id"`
}

func (r MatrixClientFeatureRow) GetId() int64 { return r.Id }
func (r MatrixClientFeatureRow) GetHashId() string {
	if r.HashId == "" {
		return r.CompHashId()
	}
	return r.HashId
}
func (r MatrixClientFeatureRow) CompHashId() string {
	h := sha1.Sum([]byte(
		fmt.Sprintf(
			"%v:%v",
			r.ServiceId,
			r.Feature)))
	return hex.EncodeToString(h[:])
}
