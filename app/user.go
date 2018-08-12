package app

import (
	"github.com/localvar/go-utils/rpc"
	"github.com/localvar/lotus/models"
	"net/http"
)

func onFindUser(r *http.Request, arg *models.FindUserArg) (*models.FindUserResult, error) {
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
	ID   int64 `json:"id"`
	Role uint8 `json:"role"`
}

func onSetUserRole(r *http.Request, arg *SetUserRoleArg) error {
	u, e := userFromCookie(r)
	if e != nil {
		return e
	}

	if u.Role != models.SystemAdmin {
		return errPermissionDenied
	}

	return models.SetUserRole(arg.ID, arg.Role)
}

func userInit() error {
	viewAddRoute("/user/list.html", viewRenderNoop, viewRequireOAuth)
	rpc.Add("find-user", onFindUser)
	rpc.Add("set-user-role", onSetUserRole)
	return nil
}
