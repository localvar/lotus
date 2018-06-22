package models

import (
	"database/sql"
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

func FindUserByNickName(name string) ([]User, error) {
	res := make([]User, 0, 256)
	name = "%" + name + "%"
	e := db.Select(&res, "SELECT * FROM `user` WHERE `nick_name` LIKE ?'", name)
	if e != nil {
		return nil, e
	}
	return res, nil
}
