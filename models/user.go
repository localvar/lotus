package models

import "time"

const (
	DisabledUser  = 0
	GeneralUser   = 1
	ContentEditor = 2
	SystemAdmin   = 10
)

type User struct {
	ID       uint64    `db:"id"`
	WechatID string    `db:"wechat_id"`
	Role     uint8     `db:"role"`
	NickName string    `db:"nick_name"`
	Avatar   string    `db:"avatar"`
	SignUpAt time.Time `db:"sign_up_at"`
}

/*
func GetUserByNickName(nickName string) (*User, error) {
	u := &User{NickName: nickName}
	if has, e := db.Get(u); e != nil {
		return nil, e
	} else if !has {
		return nil, nil
	}
	return u, nil
}

func GetUserByWechatID(wxID string) (*User, error) {
	u := &User{WechatID: wxID}
	if has, e := db.Get(u); e != nil {
		return nil, e
	} else if !has {
		return nil, nil
	}
	return u, nil
}

func SetUserRole(wxID string, role uint8) error {
	_, e := db.Update(&User{Role: role}, &User{WechatID: wxID})
	return e
}

func SetUserAvatar(wxID string, avatar string) error {
	_, e := db.Update(&User{Avatar: avatar}, &User{WechatID: wxID})
	return e
}
*/
