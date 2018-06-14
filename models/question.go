package models

import "time"

type Tag struct {
	ID        uint64    `db:"id"`
	Name      string    `db:"name"`
	CreatedBy string    `db:"created_by"`
	CreatedAt time.Time `db:"created_at"`
}

const (
	StatusNormal = iota
	StatusUrgentRequested
	StatusUrgentApproved
	StatusAnswered
	StatusFeatured
	StatusLast
)

type Question struct {
	ID         uint64    `db:"id"`
	Status     uint8     `db:"status"`
	AskedBy    uint64    `db:"asked_by"`
	AskedAt    time.Time `db:"asked_at"`
	Content    string    `db:"content"`
	Answer     string    `db:"answer"`
	AnsweredBy uint64    `db:"answered_by`
	AnsweredAt time.Time `db:"answered_by"`
	DeletedAt  time.Time `db:"deleted_at"`
}

type QuestionTag struct {
	TagID      uint64    `db:"tag_id"`
	QuestionID uint64    `db:"question_id"`
	TaggedAt   time.Time `db:"tagged_at"`
	TaggedBy   string    `db:"tagged_by"`
}

/*
func DeleteTag(id uint32) error {
	tran := db.NewSession()
	defer tran.Close()

	if e := tran.Begin(); e != nil {
		return e
	}

	if _, e := tran.Delete(&QuestionTag{TagID: id}); e != nil {
		return e
	}

	if _, e := tran.Delete(&Tag{ID: id}); e != nil {
		return e
	}

	return tran.Commit()
}

func RemoveQuestionTag(qid, tid uint32) error {
	_, e := db.Delete(&QuestionTag{TagID: tid, QuestionID: qid})
	return e
}

func SetQuestionStatus(qid uint32, status uint8) error {
	_, e := db.ID(qid).MustCols("status").Update(&Question{Status: status})
	return e
}

func GetQuestion(qid uint32) (*Question, error) {
	q := &Question{ID: qid}
	if has, e := db.Get(q); e != nil {
		return nil, e
	} else if !has {
		return nil, nil
	}
	return q, nil
}

func UpdateQuestion(qid uint32, title, content string) error {
	_, e := db.ID(qid).MustCols("title", "content").Update(&Question{
		Title:   title,
		Content: content,
	})
	return e
}

func UpdateQuestionAnswer(qid uint32, answer string) error {
	_, e := db.ID(qid).MustCols("answer").Update(&Question{Answer: answer})
	return e
}

func AnswerQuestion(qid uint32, answer string) error {
	_, e := db.ID(qid).Update(&Question{
		Status:     StatusAnswered,
		Answer:     answer,
		AnsweredAt: time.Now(),
	})
	return e
}

func RemoveQuestion(qid uint32) error {
	_, e := db.Delete(&Question{ID: qid})
	return e
}
*/
