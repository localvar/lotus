package app

import (
	"net/http"

	"github.com/localvar/go-utils/rpc"
	"github.com/localvar/lotus/models"
	"strconv"
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

func userRenderIndex(ctx *viewContext) error {
	sid := ctx.r.URL.Query().Get("id")
	id, e := strconv.ParseInt(sid, 10, 64)
	if e != nil {
		return e
	}

	ue, e := models.GetUserExByID(id)
	if e != nil {
		return e
	} else if ue == nil {
		return errUserNotExist
	}

	ctx.data["user"] = ue

	u, e := viewGetUser(ctx)
	if e != nil {
		return e
	}

	ctx.data["isAdmin"] = u.Role == models.SystemAdmin
	return nil
}

func userInit() error {
	viewAddRoute("/user/index.html", userRenderIndex, viewRequireOAuth, 0)
	viewAddRoute("/user/list.html", viewRenderNoop, viewRequireOAuth, makeRoleMask(models.SystemAdmin))
	rpc.Add("find-user", onFindUser)
	rpc.Add("set-user-role", onSetUserRole)
	return nil
}
