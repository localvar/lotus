package app

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/localvar/go-utils/config"
	"github.com/localvar/go-utils/log"
	"github.com/localvar/go-utils/rpc"
	"github.com/localvar/lotus/models"
)

const (
	cookieUserID = "user-id"
)

type IDArg struct {
	ID int64 `json:"id"`
}

func Init(debug bool) error {
	if e := wechatInit(); e != nil {
		return e
	}

	if e := viewInit(debug); e != nil {
		return e
	}

	http.HandleFunc("/", serveHTTP)
	return nil
}

func Uninit() error {
	return nil
}

// if the server is behind NGINX, a request with URL
//    https://domain.com/aa.html
// may become
//    http://127.0.0.1/aa.html
// or
//    /aa.html
// this function is used to get the correct full request URL
// note that fragement is excluded from the result
func fullRequestURL(r *http.Request) string {
	url := r.URL.String()

	if pos := strings.LastIndexByte(url, '#'); pos >= 0 {
		url = url[:pos]
	}

	server := config.GetString("/app/server", "")
	if server == "" {
		return url
	}
	if l := len(server); server[l-1] == '/' {
		server = server[:l-1]
	}

	pos := strings.Index(url, "://")
	if pos == -1 {
		return server + url
	}

	url = url[pos+3:]
	if pos = strings.IndexByte(url, '/'); pos >= 0 {
		return server + url[pos:]
	}

	return server
}

func userIDFromCookie(r *http.Request) (int64, error) {
	if render.Debug {
		id, e := strconv.ParseInt(r.URL.Query().Get("uid"), 10, 64)
		if e == nil {
			return id, nil
		}
	}

	c, e := r.Cookie(cookieUserID)
	if e != nil {
		return 0, e
	}

	id, e := strconv.ParseInt(c.Value, 10, 64)
	if e != nil {
		return 0, e
	}

	return id, nil
}

func userFromCookie(r *http.Request) (*models.User, error) {
	id, e := userIDFromCookie(r)
	if e != nil {
		return nil, e
	}
	u, e := models.GetUserByID(id)
	if e != nil {
		return nil, e
	}
	if u == nil {
		return nil, errUserNotExist
	}
	return u, nil
}

func serveHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	path := r.URL.Path

	defer func() {
		duration := time.Now().Sub(start)
		log.Debugf("[%s]\t%s\t%s", r.Method, path, duration)
	}()

	if prefix := "/api/"; strings.HasPrefix(path, prefix) {
		rpc.ServeHTTP(prefix, w, r)
		return
	}

	if strings.HasPrefix(path, "/wechat/") {
		wechatServeHTTP(w, r)
		return
	}

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if strings.HasPrefix(path, "/static/") || // image/js/css
		strings.HasPrefix(path, "/MP_verify_") || // wechat
		path == "/robots.txt" {
		http.ServeFile(w, r, r.URL.Path[1:])
		return
	}

	if strings.HasSuffix(path, ".html") {
		viewServeHTTP(w, r)
		return
	}

	w.WriteHeader(http.StatusNotFound)
}
