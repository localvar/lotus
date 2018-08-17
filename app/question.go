package app

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/localvar/go-utils/rpc"
	"github.com/localvar/lotus/models"
)

var (
	errInvalidArgument  = errors.New("参数无效")
	errQuestionNotExist = errors.New("问题不存在")
	errUserNotExist     = errors.New("用户不存在")
	errPermissionDenied = errors.New("权限不足")
)

func onGetQuestionByID(r *http.Request, arg *rpc.IDArg) (interface{}, error) {
	u, e := userFromCookie(r)
	if e != nil {
		return nil, e
	}

	q, e := models.GetQuestionByID(arg.ID, true)
	if e != nil {
		return nil, e
	}

	// admin & editor can see all questions
	if models.IsManager(u.Role) {
		return q, nil
	}

	// user can see all his/her questions
	if q.Asker == u.ID {
		return q, nil
	}

	// deleted question is invisible
	if !q.DeletedAt.IsZero() {
		return nil, errQuestionNotExist
	}

	// private & unreplied questions is invisible
	if q.Private || q.Replier == 0 {
		return nil, errPermissionDenied
	}

	// this question is visible
	return q, nil
}

func onEditQuestion(r *http.Request, q *models.Question) (interface{}, error) {
	if c := strings.TrimSpace(q.Content); utf8.RuneCount([]byte(c)) < 10 {
		return nil, errors.New("问题太短了")
	}

	u, e := userFromCookie(r)
	if e != nil {
		return nil, e
	}

	// blocked user is not allowed to ask a question
	if u.Role == models.BlockedUser {
		return nil, errPermissionDenied
	}

	if q.ID == 0 {
		if lq, _ := models.GetUserLastQuestion(u.ID); lq != nil {
			if time.Now().Sub(lq.AskedAt) < time.Hour {
				return nil, errors.New("提问过于频繁，请稍后再试")
			}
		}

		q.Asker = u.ID
		q.AskedAt = time.Now()
		q.Reply = ""
		q.Replier = 0
		q.RepliedAt = time.Time{}
		q.DeletedAt = q.RepliedAt

		return models.InsertQuestion(q)
	}

	q1, e := models.GetQuestionByID(q.ID, false)
	if e != nil {
		return nil, e
	}
	if q1 == nil || !q1.DeletedAt.IsZero() {
		return nil, errQuestionNotExist
	}

	// change question asker is not allowed
	if q1.Asker != q.Asker {
		return nil, errPermissionDenied
	}

	if q.Replier != 0 && !models.IsManager(u.Role) {
		return nil, errors.New("不允许修改已经得到回复的问题")
	}

	if q.Asker != u.ID {
		// only editor & admin is allowed to modify the other's question
		// but 'private' flag cannot be changed in this case
		if models.IsManager(u.Role) {
			q.Private = q1.Private
		} else {
			return nil, errPermissionDenied
		}
	}

	if e = models.UpdateQuestion(q); e != nil {
		return nil, e
	}
	return q, nil
}

func onReplyQuestion(r *http.Request, q *models.Question) (interface{}, error) {
	u, e := userFromCookie(r)
	if e != nil {
		return nil, e
	}
	if !models.IsManager(u.Role) {
		return nil, errPermissionDenied
	}

	q.RepliedAt = time.Now()
	q.Replier = u.ID
	if e = models.ReplyQuestion(q); e != nil {
		return nil, e
	}

	return q, nil
}

func onRemoveQuestion(r *http.Request, arg *IDArg) error {
	if arg.ID <= 0 {
		return nil
	}

	u, e := userFromCookie(r)
	if e != nil {
		return e
	}

	q, e := models.GetQuestionByID(arg.ID, false)
	if e != nil {
		return e
	}

	if u.ID != q.Asker && !models.IsManager(u.Role) {
		return errPermissionDenied
	}

	return models.RemoveQuestion(arg.ID)
}

func questionRenderList(ctx *viewContext) error {
	return nil
}

func questionRenderMine(ctx *viewContext) error {
	_, e := userIDFromCookie(ctx.r)
	if e != nil {
		return e
	}

	ctx.data["mode"] = "mine"
	ctx.tmpl = "question/list.html"
	return nil
}

func questionRenderReplied(ctx *viewContext) error {
	ctx.data["mode"] = "replied"
	ctx.tmpl = "question/list.html"
	return nil
}

func questionRenderFeatured(ctx *viewContext) error {
	ctx.data["mode"] = "featured"
	ctx.tmpl = "question/list.html"
	return nil
}

func onFindQuestion(r *http.Request, arg *models.FindQuestionArg) (interface{}, error) {
	refer, e := url.Parse(r.Referer())
	if e != nil {
		return nil, e
	}
	u, e := userFromCookie(r)
	if e != nil {
		return nil, e
	}

	switch strings.ToLower(refer.Path) {
	case "/question/mine.html":
		arg.Asker = u.ID
	case "/question/featured.html":
		arg.Featured = "yes"
		if u.ID != arg.Asker {
			arg.Private = "no"
		}
	}

	return models.FindQuestion(arg)
}

type setQuestionFlagArg struct {
	ID    int64  `json:"id"`
	Flag  string `json:"flag"`
	Value bool   `json:"value"`
}

func onSetQuestionFlag(r *http.Request, arg *setQuestionFlagArg) error {
	u, e := userFromCookie(r)
	if e != nil {
		return e
	}
	if u.Role == models.BlockedUser {
		return errPermissionDenied
	}

	q, e := models.GetQuestionByID(arg.ID, false)
	if e != nil {
		return e
	}
	if q == nil {
		return errQuestionNotExist
	}

	if arg.Flag == "featured" {
		if !models.IsManager(u.Role) {
			return errPermissionDenied
		}
		if q.Replier == 0 {
			return errors.New(`不能给未回复的问题添加"精华"标志`)
		}
	} else if arg.Flag == "urgent" {
		if q.Replier > 0 {
			return errors.New(`不能给已回复的问题添加"加急"标志`)
		} else if !models.IsManager(u.Role) && q.Asker != u.ID {
			return errPermissionDenied
		}
	} else if arg.Flag == "private" {
		if  q.Asker != u.ID {
			return errPermissionDenied
		}
	} else {
		return errInvalidArgument
	}

	return models.SetQuestionFlag(arg.ID, arg.Flag, arg.Value)
}

func questionInit() error {
	viewAddRoute("/question/list.html", questionRenderList, viewRequireOAuth, 0)
	viewAddRoute("/question/mine.html", questionRenderMine, viewRequireOAuth, 0)
	viewAddRoute("/question/replied.html", questionRenderReplied, viewRequireOAuth, 0)
	viewAddRoute("/question/featured.html", questionRenderFeatured, viewRequireOAuth, 0)

	viewAddRoute("/question/edit.html", viewRenderNoop, viewRequireOAuth, 0)
	viewAddRoute("/question/reply.html", viewRenderNoop, viewRequireOAuth, makeRoleMask(models.ContentEditor, models.SystemAdmin))

	rpc.Add("get-question-by-id", onGetQuestionByID)
	rpc.Add("set-question-flag", onSetQuestionFlag)
	rpc.Add("edit-question", onEditQuestion)
	rpc.Add("reply-question", onReplyQuestion)
	rpc.Add("remove-question", onRemoveQuestion)
	rpc.Add("find-question", onFindQuestion)
	return nil
}
