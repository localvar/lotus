package app

import (
	"net/http"

	"github.com/localvar/go-utils/rpc"
	"github.com/localvar/lotus/models"
)

func onFindUser(r *http.Request, arg *models.FindUserArg) (interface{}, error) {
	u, e := userFromCookie(r)
	if e != nil {
		return nil, e
	}

	if u.Role != models.SystemAdmin {
		return nil, errPermissionDenied
	}

	return models.FindUser(arg)
}

type SetUserRoleArg struct {
	IDs  []int64 `json:"ids"`
	Role uint8   `json:"role"`
}

func onSetUserRole(r *http.Request, arg *SetUserRoleArg) error {
	u, e := userFromCookie(r)
	if e != nil {
		return e
	}

	if u.Role != models.SystemAdmin {
		return errPermissionDenied
	}

	return models.SetUserRole(arg.IDs, arg.Role)
}

func onGetUserByID(r *http.Request, arg *IDArg) (interface{}, error) {
	ue, e := models.GetUserExByID(arg.ID)
	if e != nil {
		return nil, e
	} else if ue == nil {
		return nil, errUserNotExist
	}
	return ue, nil
}

func userInit() error {
	viewAddRoute("/user/index.html", viewRenderNoop, viewRequireOAuth, 0)
	viewAddRoute("/user/list.html", viewRenderNoop, viewRequireOAuth, makeRoleMask(models.SystemAdmin))
	rpc.Add("find-user", onFindUser)
	rpc.Add("set-user-role", onSetUserRole)
	rpc.Add("get-user-by-id", onGetUserByID)
	return nil
}
