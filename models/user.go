package models

import (
	"database/sql"
	"strings"
	"time"
)

const (
	BlockedUser   = 1
	GeneralUser   = 2
	ContentEditor = 3
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

func SetUserRole(ids []int64, role uint8) error {
	tx, e := db.Beginx()
	if e != nil {
		return e
	}

	for _, id := range ids {
		_, e := tx.Exec("UPDATE `user` SET `role`=? WHERE `id`=?", role, id)
		if e != nil {
			tx.Rollback()
			return e
		}
	}

	return tx.Commit()
}

type FindUserArg struct {
	Role       uint8  `json:"role"`
	NickName   string `json:"nickName"`
	PageSize   uint32 `json:"pageSize"`
	PageNumber uint32 `json:"pageNumber"`
}

type FindUserResult struct {
	Total      uint32 `json:"total"`
	PageSize   uint32 `json:"pageSize"`
	PageNumber uint32 `json:"pageNumber"`
	Users      []User `json:"users"`
}

func FindUser(fua *FindUserArg) (*FindUserResult, error) {
	var args []interface{}
	var wheres []string
	var sb strings.Builder
	var result FindUserResult

	if fua.PageSize == 0 {
		fua.PageSize = 1
	}
	result.PageSize = fua.PageSize

	tx, e := db.Beginx()
	if e != nil {
		return nil, e
	}
	defer tx.Rollback()

	if fua.Role > 0 {
		wheres = append(wheres, "`role`=?")
		args = append(args, fua.Role)
	}

	if len(fua.NickName) > 0 {
		wheres = append(wheres, "`nick_name` like ?")
		args = append(args, "%"+fua.NickName+"%")
	}

	sb.WriteString("SELECT COUNT(1) FROM `user`")
	if len(wheres) > 0 {
		where := "WHERE " + strings.Join(wheres, " AND ")
		sb.WriteString(where)
	}
	sb.WriteByte(';')

	if e := tx.Get(&result.Total, sb.String(), args...); e != nil {
		return nil, e
	}

	if result.Total == 0 {
		return &result, nil
	}

	if fua.PageSize*fua.PageNumber >= result.Total {
		fua.PageNumber = (result.Total+fua.PageSize-1)/fua.PageSize - 1
	}

	sb.Reset()
	sb.WriteString("SELECT * FROM `user`")
	if len(wheres) > 0 {
		where := "WHERE " + strings.Join(wheres, " AND ")
		sb.WriteString(where)
	}

	sb.WriteString(" LIMIT ?, ?")
	args = append(args, fua.PageNumber*fua.PageSize, fua.PageSize)
	sb.WriteByte(';')

	if e := tx.Select(&result.Users, sb.String(), args...); e != nil {
		return nil, e
	}

	result.PageNumber = fua.PageNumber
	return &result, nil
}
