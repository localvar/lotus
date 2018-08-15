package app

import (
	"net/http"

	"github.com/localvar/go-utils/rpc"
	"github.com/localvar/lotus/models"
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
	u, e := userFromCookie(r)
	if e != nil {
		return e
	}
	if u.Role != models.SystemAdmin {
		return errPermissionDenied
	}

	return models.DeleteTag(arg.ID)
}

func onUpdateTag(r *http.Request, arg *models.Tag) error {
	u, e := userFromCookie(r)
	if e != nil {
		return e
	}
	if u.Role != models.SystemAdmin {
		return errPermissionDenied
	}

	return models.UpdateTag(arg)
}

func tagInit() error {
	viewAddRoute("/tag.html", viewRenderNoop, viewRequireOAuth, makeRoleMask(models.SystemAdmin))
	rpc.Add("list-tags", onListTags)
	rpc.Add("add-tag", onAddTag)
	rpc.Add("update-tag", onUpdateTag)
	rpc.Add("delete-tag", onDeleteTag)
	return nil
}
