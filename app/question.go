package app

func questionRenderEdit(ctx *viewContext) error {
	return nil
}

func questionInit() error {
	viewAddRoute("/question/edit.html", questionRenderEdit, 0)
	return nil
}
