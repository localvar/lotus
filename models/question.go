package models

import (
	"database/sql"
	"time"
)

type Tag struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	CreatedBy int64     `db:"created_by"`
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
	ID         int64     `db:"id"`
	Status     uint8     `db:"status"`
	AskedBy    int64     `db:"asked_by"`
	AskedAt    time.Time `db:"asked_at"`
	Content    string    `db:"content"`
	Answer     string    `db:"answer"`
	AnsweredBy int64     `db:"answered_by"`
	AnsweredAt time.Time `db:"answered_at"`
	DeletedAt  time.Time `db:"deleted_at"`
}

type QuestionTag struct {
	QuestionID int64     `db:"question_id"`
	TagID      int64     `db:"tag_id"`
	TaggedAt   time.Time `db:"tagged_at"`
	TaggedBy   string    `db:"tagged_by"`
}

func InsertTag(tag *Tag) (*Tag, error) {
	qs := buildInsertTyped("tag", tag)

	tag.CreatedAt = time.Now()
	res, e := db.NamedExec(qs, tag)
	if e != nil {
		return nil, e
	}

	id, e := res.LastInsertId()
	if e != nil {
		return nil, e
	}

	tag.ID = id
	return tag, nil
}

func DeleteTag(id int64) error {
	tx, e := db.Beginx()
	if e != nil {
		return e
	}

	_, e = tx.Exec("DELETE FROM `question_tag` WHERE `tag_id`=?;", id)
	if e != nil {
		return e
	}

	_, e = tx.Exec("DELETE FROM `tag` WHERE `id`=?;", id)
	if e != nil {
		return e
	}

	return tx.Commit()
}

func InsertQuestionTag(qt *QuestionTag) (*QuestionTag, error) {
	qs := buildInsertTyped("question_tag", qt)

	qt.TaggedAt = time.Now()
	if _, e := db.NamedExec(qs, qt); e != nil {
		return nil, e
	}

	return qt, nil
}

func DeleteQuestionTag(qt *QuestionTag) error {
	const qs = "DELETE FROM `question_tag` WHERE `question_id`=:question_id AND `tag_id`=:tag_id;"
	_, e := db.NamedExec(qs, qt)
	return e
}

func InsertQuestion(q *Question) (*Question, error) {
	qs := buildInsertTyped("question", q)

	q.AskedAt = time.Now()
	res, e := db.NamedExec(qs, q)
	if e != nil {
		return nil, e
	}

	id, e := res.LastInsertId()
	if e != nil {
		return nil, e
	}

	q.ID = id
	return q, nil
}

func GetQuestionByID(id int64) (*Question, error) {
	var q Question
	e := db.Get(&q, "SELECT * FROM `question` WHERE `id`=?;", id)
	if e == sql.ErrNoRows {
		return nil, nil
	}
	return &q, nil
}

func RemoveQuestion(id int64) error {
	// db.Exec("UPDATE `question` SET `deleted_at`=? WHERE `id`=?;")
	return nil
}
