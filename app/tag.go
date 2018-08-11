package app

import (
	"github.com/localvar/go-utils/rpc"
	"github.com/localvar/lotus/models"
	"net/http"
)

func onListTags(r *http.Request) (interface{}, error) {
	return models.ListTag()
}

func onAddTag(r *http.Request, arg *models.Tag) (interface{}, error) {
	u, e := userFromCookie(r)
	if e != nil {
		return nil, e
	}
	if u.Role != models.SystemAdmin {
		return nil, errPermissionDenied
	}

	arg.CreatedBy = u.ID
	return models.InsertTag(arg)
}

func onDeleteTag(r *http.Request, arg *IDArg) error {
	if arg.ID <= 0 {
		return nil
	}

	u, e := userFromCookie(r)
	if e != nil {
		return e
	}

	if u.Role != models.SystemAdmin {
		return errPermissionDenied
	}

	return models.DeleteTag(arg.ID)
}

func tagRenderList(ctx *viewContext) error {
	u, e := viewGetUser(ctx)
	if e != nil {
		return e
	}
	ctx.data["isAdmin"] = u.Role == models.SystemAdmin
	return nil
}

func tagInit() error {
	viewAddRoute("/tag/list.html", tagRenderList, viewRequireOAuth)
	rpc.Add("list-tags", onListTags)
	rpc.Add("add-tag", onAddTag)
	rpc.Add("delete-tag", onDeleteTag)
	return nil
}
