package app

import (
	"fmt"
	"net/http"

	"github.com/localvar/go-utils/layout"
	"github.com/localvar/lotus/models"
)

//------------------------------------------------------------------------------
// view Context

type viewContext struct {
	w    http.ResponseWriter
	r    *http.Request
	user *models.User
	tmpl string // path of html template
	data map[string]interface{}
}

//------------------------------------------------------------------------------
// view routes

type viewHandlerFunc func(vc *viewContext) error

type viewRoute struct {
	handler  viewHandlerFunc
	attr     uint32
	roleMask uint32
}

func makeRoleMask(roles ...uint8) uint32 {
	var mask uint32
	for _, r := range roles {
		mask |= 1 << uint32(r)
	}
	return mask
}

func viewAddRoute(path string, handler viewHandlerFunc, attr uint32, roleMask uint32) {
	if _, ok := viewRoutes[path]; ok {
		panic(fmt.Errorf("route '%v' already registered", path))
	}
	viewRoutes[path] = &viewRoute{handler: handler, attr: attr, roleMask: roleMask}
}

func viewFindRoute(path string) *viewRoute {
	if r, ok := viewRoutes[path]; ok {
		return r
	}
	return nil
}

//------------------------------------------------------------------------------

const (
	viewRequireOAuth = 1 << iota
	viewUseWechatJSSDK
	viewCustomRender
)

var (
	viewRoutes = make(map[string]*viewRoute, 64)
	render     layout.Layout
)

func viewInit(debug bool) error {
	render.Build(layout.Option{
		Debug:      debug,
		Folder:     "views",
		LeftDelim:  "{[",
		RightDelim: "]}",
		Ext:        ".gohtml",
	})

	if e := questionInit(); e != nil {
		return e
	}

	if e := tagInit(); e != nil {
		return e
	}

	if e := userInit(); e != nil {
		return e
	}

	return nil
}

func viewRenderError(ctx *viewContext, err interface{}) {
	message := "未知错误"
	if err != nil {
		switch d := err.(type) {
		case string:
			message = d
		case error:
			message = d.Error()
		}
	}
	ctx.data["errmsg"] = message
	render.Render(ctx.w, "error.gohtml", ctx.data)
}

func viewRenderNoop(ctx *viewContext) error {
	return nil
}

func viewGetUserID(ctx *viewContext) (int64, error) {
	if ctx.user != nil {
		return ctx.user.ID, nil
	}
	return userIDFromCookie(ctx.r)
}

func viewGetUser(ctx *viewContext) (*models.User, error) {
	if ctx.user != nil {
		return ctx.user, nil
	}
	u, e := userFromCookie(ctx.r)
	if e != nil {
		return nil, e
	}
	ctx.user = u
	return u, nil
}

func viewServeHTTP(w http.ResponseWriter, r *http.Request) {
	vr := viewFindRoute(r.URL.Path)
	if vr == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	ctx := viewContext{
		w: w,
		r: r,
		data: map[string]interface{}{
			"debug": render.Debug,
		},
	}
	if vr.attr&viewRequireOAuth != 0 {
		if ok, e := wechatOAuth(&ctx); !ok {
			if e != nil {
				viewRenderError(&ctx, e)
			}
			return
		}
	}

	if vr.roleMask != 0 {
		u, e := viewGetUser(&ctx)
		if e != nil {
			viewRenderError(&ctx, e)
			return
		}
		if vr.roleMask&(1<<uint32(u.Role)) == 0 {
			viewRenderError(&ctx, errPermissionDenied)
			return
		}
	}

	if e := vr.handler(&ctx); e != nil {
		viewRenderError(&ctx, e)
		return
	}

	if vr.attr&viewUseWechatJSSDK != 0 {
		ctx.data["wxcfg"] = wechatNewJSSDKConfig(r)
	}

	path := ctx.tmpl
	if len(path) == 0 {
		path = r.URL.Path[1:]
	}
	path = path[:len(path)-len(".html")] + render.Ext // html ==> {Ext}
	render.Render(w, path, ctx.data)
}
