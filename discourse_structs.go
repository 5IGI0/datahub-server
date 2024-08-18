package main

import (
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"fmt"
)

type DiscourseInstanceRow struct {
	Id            int64  `db:"id"`
	Secure        int8   `db:"secure"`
	Host          string `db:"host"`
	RevDomain     string `db:"rev_domain"`
	Root          string `db:"root"`
	Title         string `db:"title"`
	Description   string `db:"description"`
	LoginRequired int8   `db:"login_required"`
	RawData       string `db:"raw_data"`
	HashId        string `db:"hash_id"`
}

func (r DiscourseInstanceRow) GetId() int64 { return r.Id }
func (r DiscourseInstanceRow) GetHashId() string {
	if r.HashId == "" {
		return r.CompHashId()
	}
	return r.HashId
}
func (r DiscourseInstanceRow) CompHashId() string {
	h := sha1.Sum([]byte(
		fmt.Sprintf(
			"%v:%v:%v",
			r.Secure,
			r.Host,
			r.Root)))
	return hex.EncodeToString(h[:])
}

type DiscourseCategoryRow struct {
	Id               int64         `db:"id"`
	InstanceId       int64         `db:"instance_id"`
	CategoryId       int64         `db:"category_id"`
	IsActive         int8          `db:"is_active"`
	Name             string        `db:"name"`
	Slug             string        `db:"slug"`
	Description      string        `db:"description"`
	RawData          string        `db:"raw_data"`
	ParentCategoryId sql.NullInt64 `db:"parent_category_id"`
	HashId           string        `db:"hash_id"`
}

func (r DiscourseCategoryRow) GetId() int64 { return r.Id }
func (r DiscourseCategoryRow) GetHashId() string {
	if r.HashId == "" {
		return r.CompHashId()
	}
	return r.HashId
}
func (r DiscourseCategoryRow) CompHashId() string {
	h := sha1.Sum([]byte(
		fmt.Sprintf(
			"%v:%v",
			r.InstanceId,
			r.CategoryId)))
	return hex.EncodeToString(h[:])
}

type DiscourseTopicRow struct {
	Id         int64         `db:"id"`
	InstanceId int64         `db:"instance_id"`
	TopicId    int64         `db:"topic_id"`
	Title      string        `db:"title"`
	CategoryId int64         `db:"category_id"`
	UserId     sql.NullInt64 `db:"user_id"`
	RawData    string        `db:"raw_data"`
	IsDataFull int8          `db:"is_data_full"`
	HashId     string        `db:"hash_id"`
}

func (r DiscourseTopicRow) GetId() int64 { return r.Id }
func (r DiscourseTopicRow) GetHashId() string {
	if r.HashId == "" {
		return r.CompHashId()
	}
	return r.HashId
}
func (r DiscourseTopicRow) CompHashId() string {
	h := sha1.Sum([]byte(
		fmt.Sprintf(
			"%v:%v",
			r.InstanceId,
			r.TopicId)))
	return hex.EncodeToString(h[:])
}

type DiscoursePostRow struct {
	Id         int64  `db:"id"`
	InstanceId int64  `db:"instance_id"`
	TopicId    int64  `db:"topic_id"`
	PostId     int64  `db:"post_id"`
	UserId     int64  `db:"user_id"`
	RawData    string `db:"raw_data"`
	HashId     string `db:"hash_id"`
}

func (r DiscoursePostRow) GetId() int64 { return r.Id }
func (r DiscoursePostRow) GetHashId() string {
	if r.HashId == "" {
		return r.CompHashId()
	}
	return r.HashId
}
func (r DiscoursePostRow) CompHashId() string {
	h := sha1.Sum([]byte(
		fmt.Sprintf(
			"%v:%v",
			r.InstanceId,
			r.PostId)))
	return hex.EncodeToString(h[:])
}

const (
	DISCOURSE_USER_ADMIN_FLAG     = 1
	DISCOURSE_USER_MODERATOR_FLAG = 2
)

type DiscourseUserRow struct {
	Id               int64          `db:"id"`
	InstanceId       int64          `db:"instance_id"`
	UserId           int64          `db:"user_id"`
	Username         string         `db:"username"`
	Name             string         `db:"name"`
	Title            string         `db:"title"`
	Flags            uint8          `db:"flags"`
	WebSiteDomain    sql.NullString `db:"website_domain"`
	RevWebSiteDomain sql.NullString `db:"rev_website_domain"`
	RawData          string         `db:"raw_data"`
	IsDataFull       int8           `db:"is_data_full"`
	HashId           string         `db:"hash_id"`
}

func (r DiscourseUserRow) GetId() int64 { return r.Id }
func (r DiscourseUserRow) GetHashId() string {
	if r.HashId == "" {
		return r.CompHashId()
	}
	return r.HashId
}
func (r DiscourseUserRow) CompHashId() string {
	h := sha1.Sum([]byte(
		fmt.Sprintf(
			"%v:%v",
			r.InstanceId,
			r.UserId)))
	return hex.EncodeToString(h[:])
}

type DiscourseTagRow struct {
	Id          int64          `db:"id"`
	InstanceId  int64          `db:"instance_id"`
	Name        string         `db:"name"`
	Description sql.NullString `db:"description"`
	HashId      string         `db:"hash_id"`
}

func (r DiscourseTagRow) GetId() int64 { return r.Id }
func (r DiscourseTagRow) GetHashId() string {
	if r.HashId == "" {
		return r.CompHashId()
	}
	return r.HashId
}
func (r DiscourseTagRow) CompHashId() string {
	h := sha1.Sum([]byte(
		fmt.Sprintf(
			"%v:%v",
			r.InstanceId,
			r.Name)))
	return hex.EncodeToString(h[:])
}
