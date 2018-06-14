package app

import (
	"fmt"
	"net/http"

	"github.com/localvar/go-utils/layout"
)

//------------------------------------------------------------------------------
// view routes

type viewRoute struct {
	handler http.HandlerFunc
	attr    uint
}

func addViewRoute(path string, handler http.HandlerFunc, attr uint) {
	if _, ok := viewRoutes[path]; ok {
		panic(fmt.Errorf("route '%v' already registered", path))
	}
	viewRoutes[path] = &viewRoute{handler: handler, attr: attr}
}

func findViewRoute(path string) *viewRoute {
	if r, ok := viewRoutes[path]; ok {
		return r
	}
	return nil
}

//------------------------------------------------------------------------------

const (
	viewRequireOAuth = 0x00000001
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
	return nil
}

func viewRenderSimple(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[1:]
	path = path[:len(path)-4] + "gohtml" // html ==> gohtml
	render.Render(w, path, nil)
}

func viewRenderWechatSimple(w http.ResponseWriter, r *http.Request) {
	data := wechatNewDataWithConfig(r)
	path := r.URL.Path[1:]
	path = path[:len(path)-4] + "gohtml" // html ==> gohtml
	render.Render(w, path, data)
}

func viewRenderError(w http.ResponseWriter, r *http.Request, data interface{}) {
	message := "未知错误"
	if data != nil {
		switch d := data.(type) {
		case string:
			message = d
		case error:
			message = d.Error()
		}
	}
	render.Render(w, "error.gohtml", message)
}

func viewServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[1:]
	vr := findViewRoute(path)
	if vr == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if vr.attr&viewRequireOAuth != 0 {
		if ok, e := wechatOAuth(w, r); !ok {
			if e != nil {
				viewRenderError(w, r, e)
			}
			return
		}
	}

	vr.handler(w, r)
}
