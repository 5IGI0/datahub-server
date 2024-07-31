package main

import (
	"database/sql/driver"
	"time"
)

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
