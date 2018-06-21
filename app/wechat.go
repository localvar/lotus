package app

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/localvar/go-utils/config"
	"github.com/localvar/go-utils/wechat"
	"github.com/localvar/lotus/models"
)

var wxclnt *wechat.Client

const (
	urlOAuth2 = "https://open.weixin.qq.com/connect/oauth2/authorize?appid=APPID&redirect_uri=REDIRECT_URI&response_type=code&scope=snsapi_userinfo#wechat_redirect"
)

func wechatNewJSSDKConfig(r *http.Request) *wechat.JSSDKConfig {
	url := fullRequestURL(r)
	return wxclnt.NewJSSDKConfig(true, render.Debug, url)
}

func wechatOAuth(ctx *viewContext) (bool, error) {
	if _, e := userIDFromCookie(ctx.r); e == nil {
		return true, nil
	}

	code := ctx.r.URL.Query().Get("code")
	if len(code) == 0 {
		from := url.QueryEscape(fullRequestURL(ctx.r))
		url := strings.Replace(urlOAuth2, "APPID", wxclnt.AppID, 1)
		url = strings.Replace(url, "REDIRECT_URI", from, 1)
		http.Redirect(ctx.w, ctx.r, url, http.StatusTemporaryRedirect)
		return false, nil
	}

	ui, e := wxclnt.GetUserInfoViaOAuth2(code)
	if e != nil {
		return false, e
	}

	u, e := models.GetUserByWxOpenID(ui.OpenID)
	if e != nil {
		return false, e
	}

	if u == nil {
		u = &models.User{
			WxOpenID:  ui.OpenID,
			WxUnionID: ui.UnionID,
			Role:      models.GeneralUser,
			NickName:  ui.Nickname,
			Avatar:    ui.HeadImageURL,
			SignUpAt:  time.Now(),
		}
		if u, e = models.InsertUser(u); e != nil {
			return false, e
		}
	}

	c := &http.Cookie{
		Path:     "/",
		HttpOnly: true,
		Name:     cookieUserID,
		Value:    strconv.FormatInt(u.ID, 10),
	}
	http.SetCookie(ctx.w, c)
	ctx.user = u
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
