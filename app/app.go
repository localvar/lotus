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
// this function is used to get the correct full request URL
// note that fragement is excluded from the result
func fullRequestURL(r *http.Request) string {
	url := r.URL.String()

	if pos := strings.IndexByte(url, '#'); pos >= 0 {
		url = url[:pos]
	}

	server := config.GetString("/app/server", "")
	if server == "" {
		return url
	}
	if l := len(server); server[l-1] == '/' {
		server = server[:l-1]
	}

	if pos := strings.IndexByte(url[8:], '/'); pos >= 0 {
		return server + url[pos+8:]
	}

	return server
}

func userIDFromCookie(r *http.Request) (int64, error) {
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
	return models.GetUserByID(id)
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
