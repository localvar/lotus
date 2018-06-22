package app

import (
	"errors"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/localvar/go-utils/rpc"
	"github.com/localvar/lotus/models"
)

var (
	errQuestionNotExist = errors.New("question does not exist")
	errUserNotExist     = errors.New("user does not exist")
	errPermissionDenied = errors.New("permission denied")
	errContentTooShort  = errors.New("content too short")
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

	if u.Role != models.GeneralUser {
		return q, nil
	}

	if !q.DeletedAt.IsZero() {
		return nil, errQuestionNotExist
	}

	if q.Asker == u.ID {
		return q, nil
	}

	if q.Private || q.Replier == 0 {
		return nil, errPermissionDenied
	}

	return q, nil
}

func onEditQuestion(r *http.Request, q *models.Question) (interface{}, error) {
	if c := strings.TrimSpace(q.Content); len(c) < 20 || utf8.RuneCount([]byte(c)) < 10 {
		return nil, errContentTooShort
	}

	u, e := userFromCookie(r)
	if e != nil {
		return nil, e
	}

	if q.ID == 0 {
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

	if q1.Asker != q.Asker {
		return nil, errPermissionDenied
	}

	if q.Asker != u.ID {
		if u.Role == models.GeneralUser {
			q.Private = q1.Private
		} else {
			return nil, errPermissionDenied
		}
	}

	e = models.UpdateQuestion(q)
	if e != nil {
		return nil, e
	}
	return q, e
}

func onAnswerQuestion(r *http.Request, q *models.Question) (interface{}, error) {
	return nil, nil
}

func questionInit() error {
	viewAddRoute("/question/edit.html", viewRenderNoop, 0)
	rpc.Add("get-question-by-id", onGetQuestionByID)
	rpc.Add("edit-question", onEditQuestion)
	rpc.Add("answer-question", onAnswerQuestion)
	return nil
}
