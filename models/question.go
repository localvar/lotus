package models

import (
	"time"
)

type Tag struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
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
	AnsweredBy int64     `db:"answered_by`
	AnsweredAt time.Time `db:"answered_at"`
	DeletedAt  time.Time `db:"deleted_at"`
}

type QuestionTag struct {
	TagID      int64     `db:"tag_id"`
	QuestionID int64     `db:"question_id"`
	TaggedAt   time.Time `db:"tagged_at"`
	TaggedBy   string    `db:"tagged_by"`
}

func InsertTag(tag *Tag) (*Tag, error) {
	res, e := db.Exec("INSERT INTO `tag`(`name`) VALUES (?)", tag.Name)
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

	_, e = tx.Exec("DELETE FROM `question_tag` WHERE `tag_id`=?", id)
	if e != nil {
		return e
	}

	_, e = tx.Exec("DELETE FROM `tag` WHERE `id`=?", id)
	if e != nil {
		return e
	}

	return tx.Commit()
}

func RenameTag(id int64, name string) error {
	_, e := db.Exec("UPDATE `tag` SET `name`=? WHERE `id`=?", name, id)
	return e
}

func InsertQuestionTag(qt *QuestionTag) error {
	tx, e := db.Beginx()
	if e != nil {
		return e
	}

	qs := "SELECT 1 FROM `question_tag` WHERE `question_id`=:question_id AND `tag_id`=:tag_id LIMIT 1"
	if has, e := isExist(tx, qs, qt); e != nil {
		return e
	} else if has {
		return nil
	}

	qs = "INSERT INTO `question_tag`(`question_id`, `tag_id`, `tagged_at`, `tagged_by`)" +
		" VALUES(:question_id, :tag_id, :tagged_at, :tagged_by)"
	qt.TaggedAt = time.Now()
	if _, e = tx.NamedExec(qs, qt); e != nil {
		return e
	}

	return tx.Commit()
}

func DeleteQuestionTag(qt *QuestionTag) error {
	const qs = "DELETE FROM `question_tag` WHERE `question_id`=:question_id AND `tag_id`=:tag_id"
	_, e := db.NamedExec(qs, qt)
	return e
}

func InsertQuestion(q *Question) error {
	tx, e := db.Beginx()
	if e != nil {
		return e
	}

	return tx.Commit()
}
