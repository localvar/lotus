package app

import (
	"github.com/localvar/go-utils/rpc"
	"github.com/localvar/lotus/models"
	"net/http"
)

func tagRenderList(ctx *viewContext) error {
	tags, e := models.ListTag()
	if e != nil {
		return e
	}
	ctx.data["tags"] = tags
	return nil
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

func onDeleteTag(r *http.Request, arg *rpc.IDArg) error {
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

func tagInit() error {
	viewAddRoute("/tag/list.html", tagRenderList, viewRequireOAuth)
	rpc.Add("add-tag", onAddTag)
	rpc.Add("delete-tag", onDeleteTag)
	return nil
}
