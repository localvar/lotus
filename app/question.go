package app

import "net/http"

func questionRenderEdit(w http.ResponseWriter, r *http.Request) {
	viewRenderSimple(w, r)
}

func questionInit() error {
	viewAddRoute("/question/edit.html", questionRenderEdit, 0)
	return nil
}
