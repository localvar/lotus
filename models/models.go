package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/localvar/go-utils/config"
)

var db *sqlx.DB

type Setting struct {
	Name  string `db:"name"`
	Value string `db:"value"`
}

func Init(debug bool) error {
	cs := config.String("/app/database")
	xdb, e := sqlx.Connect("mysql", cs)
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
