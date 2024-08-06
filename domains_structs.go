package main

import (
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"fmt"
)

type DomainRow struct {
	Id           int64          `db:"id"`
	Domain       string         `db:"domain"`
	RevDomain    string         `db:"rev_domain"`
	IsActive     int8           `db:"is_active"`
	CurrentFlags uint64         `db:"cur_flags"`
	OldFlags     uint64         `db:"old_flags"`
	FirstSeen    sql.NullString `db:"first_seen"`
	LastSeen     sql.NullString `db:"last_seen"`
	LastCheck    sql.NullString `db:"last_check"`
	CheckVer     uint16         `db:"check_ver"`
}

type DomainScanArchiveRow struct {
	Id        int64  `db:"id"`
	DomainId  int64  `db:"domain_id"`
	RawResult string `db:"raw_result"`
}

type DNSRecordRow struct {
	Id       int64          `db:"id"`
	DomainId int64          `db:"domain_id"`
	IsActive int8           `db:"is_active"`
	Type     uint16         `db:"type"`
	Addr     sql.NullString `db:"addr"`
	Priority sql.NullInt32  `db:"priority"`
	HashId   string         `db:"hash_id"`
}

func (r DNSRecordRow) GetId() int64 { return r.Id }
func (r DNSRecordRow) GetHashId() string {
	if r.HashId == "" {
		return r.CompHashId()
	}
	return r.HashId
}
func (r DNSRecordRow) CompHashId() string {
	h := sha1.Sum([]byte(
		fmt.Sprintf(
			"%v:%v:%v:%v",
			r.DomainId,
			r.Type,
			r.Addr.String,
			r.Priority.Int32)))
	return hex.EncodeToString(h[:])
}

const (
	DOMAIN_IPV4_FLAG    = 1 << 0
	DOMAIN_IPV6_FLAG    = 1 << 1
	DOMAIN_MX_FLAG      = 1 << 2
	DOMAIN_HTTP_FLAG    = 1 << 3
	DOMAIN_HTTPS_FLAG   = 1 << 4
	DOMAIN_CRASHED_FLAG = 1 << 63
)

var DomainScanTag2Flag = map[string]uint64{
	"IPv4":         DOMAIN_IPV4_FLAG,
	"IPv6":         DOMAIN_IPV6_FLAG,
	"mail":         DOMAIN_MX_FLAG,
	"crashed-scan": DOMAIN_CRASHED_FLAG,
}
