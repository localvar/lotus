package models

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/localvar/go-utils/config"
	_ "github.com/mattn/go-sqlite3"
)

var db *sqlx.DB

func Init(debug bool) error {
	driver := config.String("/database/driver")
	dsn := config.String("/database/dsn")

	xdb, e := sqlx.Connect(driver, dsn)
	if e != nil {
		return e
	}

	db = xdb
	return upgrade()
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

func getFieldList(o interface{}) []string {
	t := reflect.TypeOf(o).Elem()
	num := t.NumField()
	res := make([]string, 0, num)
	for i := 0; i < num; i++ {
		tag := t.Field(i).Tag
		if s := tag.Get("dbx"); !strings.Contains(s, "<-") {
			if s = tag.Get("db"); s != "-" {
				res = append(res, s)
			}
		}
	}
	return res
}

func buildInsert(table string, fields ...string) string {
	var sb strings.Builder
	sb.WriteString("INSERT INTO `")
	sb.WriteString(table)
	sb.WriteString("`(")
	for i, f := range fields {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteByte('`')
		sb.WriteString(f)
		sb.WriteByte('`')
	}
	sb.WriteString(") VALUES (")
	for i, f := range fields {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteByte(':')
		sb.WriteString(f)
	}
	sb.WriteString(");")
	return sb.String()
}

func buildInsertTyped(table string, o interface{}) string {
	fields := getFieldList(o)
	return buildInsert(table, fields...)
}

type schemaScript struct {
	ver   int
	stmts []string
}

func loadSchemaScript() ([]schemaScript, error) {
	const syntaxErrFmt = "expect ';' at line: %v"

	f, e := os.Open("models/schema.sql")
	if e != nil {
		if e == os.ErrNotExist {
			e = nil
		}
		return nil, e
	}
	defer f.Close()

	var (
		ln      int
		res     []schemaScript
		cur     schemaScript
		reVer   = regexp.MustCompile(`^-- +(?i:version) +(\d+)$`)
		sb      = strings.Builder{}
		scanner = bufio.NewScanner(f)
	)

	for scanner.Scan() {
		ln++
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 {
			continue
		}

		if !strings.HasPrefix(line, "--") {
			if sb.Len() > 0 {
				sb.WriteByte(' ')
			}
			sb.WriteString(line)

			if line[len(line)-1] == ';' {
				cur.stmts = append(cur.stmts, sb.String())
				sb.Reset()
			}
			continue
		}

		if m := reVer.FindStringSubmatch(line); len(m) == 3 {
			if sb.Len() > 0 {
				return nil, fmt.Errorf(syntaxErrFmt, ln)
			}
			if len(cur.stmts) > 0 {
				res = append(res, cur)
			}
			ver, _ := strconv.Atoi(m[3])
			if l := len(res); l > 0 && res[l-1].ver >= ver {
				return nil, fmt.Errorf("version %v appears after a higher version", ver)
			}
			cur.ver = ver
			cur.stmts = nil
		}
	}

	if e := scanner.Err(); e != nil {
		return nil, e
	}

	if sb.Len() > 0 {
		return nil, fmt.Errorf(syntaxErrFmt, ln)
	}

	if len(cur.stmts) > 0 {
		res = append(res, cur)
	}

	return res, nil
}

func upgrade() error {
	const schemaVersion = "schema_version"

	sss, e := loadSchemaScript()
	if e != nil {
		return e
	}
	if len(sss) == 0 {
		return nil
	}

	ver, e := GetOptionInt(nil, schemaVersion)
	if e != nil { // regards all errors as schema not created
		ver = -1 // set ver to -1 to execute all statements
	}

	toVer := sss[len(sss)-1].ver
	if ver >= toVer {
		return nil
	}

	tx, e := db.Beginx()
	if e != nil {
		return e
	}

	driver := db.DriverName()
	for _, ss := range sss {
		if ss.ver <= ver {
			continue
		}
		for _, s := range ss.stmts {
			if driver == "mysql" {
				s = strings.Replace(s, "WITHOUT ROWID", "", 1)
				s = strings.Replace(s, "INTEGER", "BIGINT(20)", 1)
			}
			if _, e = tx.Exec(s); e != nil {
				tx.Rollback()
				return e
			}
		}
	}

	if e = SetOption(tx, schemaVersion, toVer); e != nil {
		return e
	}

	return tx.Commit()
}
