package models

import (
	"database/sql"
	"strings"
	"time"
)

const (
	GeneralUser   = 0
	ContentEditor = 1
	SystemAdmin   = 10
)

type User struct {
	ID        int64     `db:"id" json:"id" dbx:"<-"`
	WxOpenID  string    `db:"wx_open_id" json:"-"`
	WxUnionID string    `db:"wx_union_id" json:"-"`
	Role      uint8     `db:"role" json:"role"`
	NickName  string    `db:"nick_name" json:"nickName"`
	Avatar    string    `db:"avatar" json:"avatar"`
	SignUpAt  time.Time `db:"sign_up_at" json:"signUpAt"`
	FoulCount uint32    `db:"foul_count" json:"foulCount"`
	BlockedAt time.Time `db:"blocked_at" json:"blockedAt"`
}

func InsertUser(u *User) (*User, error) {
	qs := buildInsertTyped("user", u)

	res, e := db.NamedExec(qs, u)
	if e != nil {
		return nil, e
	}

	id, e := res.LastInsertId()
	if e != nil {
		return nil, e
	}

	u.ID = id
	return u, nil
}

func GetUserByID(id int64) (*User, error) {
	var u User
	e := db.Get(&u, "SELECT * FROM `user` WHERE `id`=?", id)
	if e == sql.ErrNoRows {
		return nil, nil
	}
	return &u, nil
}

func GetUserByWxOpenID(id string) (*User, error) {
	var u User
	e := db.Get(&u, "SELECT * FROM `user` WHERE `wx_open_id`=?", id)
	if e == sql.ErrNoRows {
		return nil, nil
	}
	return &u, nil
}

func SetUserRole(id int64, role uint8) error {
	_, e := db.Exec("UPDATE `user` SET `role`=? WHERE `id`=?", role, id)
	return e
}

type FindUserArg struct {
	Name   string `json:"name"`
	Offset int64  `json:"offset"`
	Count  int64  `json:"count"`
}

type FindUserResult struct {
	Total int64  `json:"total"`
	Users []User `json:"users"`
}

func FindUser(fua *FindUserArg) (*FindUserResult, error) {
	var result FindUserResult
	var args []interface{}
	var sb strings.Builder

	tx, e := db.Beginx()
	if e != nil {
		return nil, e
	}
	defer tx.Rollback()

	sb.WriteString("SELECT COUNT(1) FROM `user`")
	if len(fua.Name) > 0 {
		sb.WriteString(" WHERE `nick_name` like ?")
		args = append(args, "%"+fua.Name+"%")
	}
	sb.WriteByte(';')

	if e := tx.Get(&result.Total, sb.String(), args...); e != nil {
		return nil, e
	}

	sb.Reset()
	sb.WriteString("SELECT * FROM `user`")
	if len(fua.Name) > 0 {
		sb.WriteString(" WHERE `nick_name` like ?")
	}

	sb.WriteString(" LIMIT ?, ?")
	args = append(args, fua.Offset, fua.Count)
	sb.WriteByte(';')

	if e := tx.Select(&result.Users, sb.String(), args...); e != nil {
		return nil, e
	}

	return &result, nil
}
