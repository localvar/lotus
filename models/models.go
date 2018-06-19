package models

import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/jmoiron/sqlx"
	"github.com/localvar/go-utils/config"
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
