package models

import (
	"fmt"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
)

type Option struct {
	Name  string `db:"name"`
	Value string `db:"value"`
}

func GetOptionString(tx *sqlx.Tx, name string) (string, error) {
	const qs = "SELECT * FROM `options` WHERE `name`=?"

	var (
		e error
		o Option
	)

	if tx != nil {
		e = tx.Get(&o, qs, name)
	} else {
		e = db.Get(&o, qs, name)
	}

	if e != nil {
		return "", e
	}

	return o.Value, nil
}

func GetOptionBool(tx *sqlx.Tx, name string) (bool, error) {
	s, e := GetOptionString(tx, name)
	if e != nil {
		return false, e
	}
	return strconv.ParseBool(s)
}

func GetOptionInt(tx *sqlx.Tx, name string) (int, error) {
	s, e := GetOptionString(tx, name)
	if e != nil {
		return 0, e
	}
	v, e := strconv.ParseInt(s, 10, 64)
	return int(v), e
}

func GetOptionUint(tx *sqlx.Tx, name string) (uint, error) {
	s, e := GetOptionString(tx, name)
	if e != nil {
		return 0, e
	}
	v, e := strconv.ParseUint(s, 10, 64)
	return uint(v), e
}

func GetOptionUint32(tx *sqlx.Tx, name string) (uint32, error) {
	s, e := GetOptionString(tx, name)
	if e != nil {
		return 0, e
	}
	v, e := strconv.ParseUint(s, 10, 32)
	return uint32(v), e
}

func GetOptionUint64(tx *sqlx.Tx, name string) (uint64, error) {
	s, e := GetOptionString(tx, name)
	if e != nil {
		return 0, e
	}
	return strconv.ParseUint(s, 10, 64)
}

func GetOptionInt32(tx *sqlx.Tx, name string) (int32, error) {
	s, e := GetOptionString(tx, name)
	if e != nil {
		return 0, e
	}
	v, e := strconv.ParseInt(s, 10, 32)
	return int32(v), e
}

func GetOptionInt64(tx *sqlx.Tx, name string) (int64, error) {
	s, e := GetOptionString(tx, name)
	if e != nil {
		return 0, e
	}
	return strconv.ParseInt(s, 10, 64)
}

func GetOptionFloat32(tx *sqlx.Tx, name string) (float32, error) {
	s, e := GetOptionString(tx, name)
	if e != nil {
		return 0, e
	}
	v, e := strconv.ParseFloat(s, 32)
	return float32(v), e
}

func GetOptionFloat64(tx *sqlx.Tx, name string) (float64, error) {
	s, e := GetOptionString(tx, name)
	if e != nil {
		return 0, e
	}
	return strconv.ParseFloat(s, 64)
}

func GetOptionTime(tx *sqlx.Tx, name string) (time.Time, error) {
	s, e := GetOptionString(tx, name)
	if e != nil {
		return time.Time{}, e
	}
	return time.Parse(time.RFC3339, s)
}

func GetOptionDuration(tx *sqlx.Tx, name string) (time.Duration, error) {
	s, e := GetOptionString(tx, name)
	if e != nil {
		return 0, e
	}
	return time.ParseDuration(s)
}

func SetOption(tx *sqlx.Tx, name string, value interface{}) error {
	const qs = "REPLACE INTO `options`(`name`, `value`) VALUES(?, ?);"

	var (
		e  error
		sv string
	)

	switch v := value.(type) {
	case time.Time:
		sv = v.Format(time.RFC3339)
	default:
		sv = fmt.Sprint(value)
	}

	if tx != nil {
		_, e = tx.Exec(qs, name, sv)
	} else {
		_, e = db.Exec(qs, name, sv)
	}

	return e
}
