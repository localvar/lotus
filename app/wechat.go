package app

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/localvar/go-utils/config"
	"github.com/localvar/go-utils/wechat"
)

var wxclnt *wechat.Client

const (
	urlOAuth2 = "https://open.weixin.qq.com/connect/oauth2/authorize?appid=APPID&redirect_uri=REDIRECT_URI&response_type=code&scope=snsapi_base#wechat_redirect"
)

func wechatNewDataWithConfig(r *http.Request) map[string]interface{} {
	url := fullRequestURL(r)
	cfg := wxclnt.NewJSSDKConfig(true, render.Debug, url)
	return map[string]interface{}{"wxcfg": cfg}
}

func wechatOAuth(w http.ResponseWriter, r *http.Request) (bool, error) {
	if userIDFromCookie(r) != "" {
		return true, nil
	}

	code := r.URL.Query().Get("code")
	if len(code) == 0 {
		from := url.QueryEscape(fullRequestURL(r))
		url := strings.Replace(urlOAuth2, "APPID", wxclnt.AppID, 1)
		url = strings.Replace(url, "REDIRECT_URI", from, 1)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
		return false, nil
	}

	token, e := wxclnt.GetOAuth2AccessToken(code)
	if e != nil {
		return false, e
	}

	c := &http.Cookie{
		Path:     "/",
		HttpOnly: true,
		Name:     cookieUserID,
		Value:    token.OpenID,
	}
	http.SetCookie(w, c)
	return true, nil
}

func wechatInit() error {
	appID := config.String("/wechat/appid")
	secret := config.String("/wechat/appsecret")
	wxclnt = wechat.NewClient(appID, secret)
	return nil
}

func wechatServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		query := r.URL.Query()

		token := config.String("/wechat/token")
		ts := query.Get("timestamp")
		nonce := query.Get("nonce")
		signature := query.Get("signature")
		echostr := query.Get("echostr")

		if wechat.VerifySignature(token, ts, nonce, signature) {
			w.Write([]byte(echostr))
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
		return
	}

	w.WriteHeader(http.StatusBadRequest)
}
