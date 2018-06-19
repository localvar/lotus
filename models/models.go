package models

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/localvar/go-utils/config"
	_ "github.com/mattn/go-sqlite3"
)

var db *sqlx.DB

type Setting struct {
	Name  string `db:"name"`
	Value string `db:"value"`
}

func Init(debug bool) error {
	driver := config.String("/database/driver")
	dsn := config.String("/database/dsn")
	xdb, e := sqlx.Connect(driver, dsn)
	if e != nil {
		return e
	}
	db = xdb
	return nil
}

func Uninit() error {
	db.Close()
	return nil
}

// isExist checks whether a record is exist or not
// 'qs' must be like: SELECT 1 FROM table WHERE col1=:col1 LIMIT 1
func isExist(tx *sqlx.Tx, qs string, arg interface{}) (bool, error) {
	var (
		e     error
		dummy int
	)

	if tx == nil {
		e = db.Get(&dummy, qs, arg)
	} else {
		e = tx.Get(&dummy, qs, arg)
	}

	if e == sql.ErrNoRows {
		return false, nil
	}

	if e != nil {
		return false, e
	}

	return true, nil
}
