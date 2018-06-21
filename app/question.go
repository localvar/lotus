package app

import (
	"errors"
	"net/http"

	"github.com/localvar/go-utils/rpc"
	"github.com/localvar/lotus/models"
)

var (
	errQuestionNotExist = errors.New("question does not exist")
)

func onGetQuestion(r *http.Request, arg *rpc.IDArg) (interface{}, error) {
	q, e := models.GetQuestionByID(arg.ID)
	if e != nil {
		return nil, e
	}
	if !q.DeletedAt.IsZero() {
		return nil, errQuestionNotExist
	}

	return q, nil
}

func questionInit() error {
	viewAddRoute("/question/edit.html", viewRenderNoop, 0)
	rpc.Add("get-question-by-id", onGetQuestion)
	return nil
}
