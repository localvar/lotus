package models

import (
	"database/sql"
	"strings"
	"time"
)

type Tag struct {
	ID        int64     `db:"id" json:"id" dbx:"<-"`
	Name      string    `db:"name" json:"name"`
	CreatedAt time.Time `db:"created_at" json:"-"`
	CreatedBy int64     `db:"created_by" json:"-"`
}

type Question struct {
	ID          int64     `db:"id" json:"id" dbx:"<-"`
	Urgent      bool      `db:"urgent" json:"urgent"`
	Private     bool      `db:"private" json:"private"`
	Featured    bool      `db:"featured" json:"featured"`
	Asker       int64     `db:"asker" json:"asker"`
	AskerName   string    `db:"asker_name" json:"askerName" dbx:"<-"`
	AskedAt     time.Time `db:"asked_at" json:"askedAt"`
	Content     string    `db:"content" json:"content"`
	Reply       string    `db:"reply" json:"reply"`
	Replier     int64     `db:"replier" json:"replier"`
	ReplierName string    `db:"replier_name" json:"replierName" dbx:"<-"`
	RepliedAt   time.Time `db:"replied_at" json:"repliedAt"`
	DeletedAt   time.Time `db:"deleted_at" json:"-"`
	Tags        []Tag     `db:"-" json:"tags,omitempty"`
}

type QuestionTag struct {
	QuestionID int64     `db:"question_id"`
	TagID      int64     `db:"tag_id"`
	TaggedAt   time.Time `db:"tagged_at"`
	TaggedBy   int64     `db:"tagged_by"`
}

func ListTag() ([]Tag, error) {
	res := make([]Tag, 0, 32)
	if e := db.Select(&res, "SELECT * FROM `tag`"); e == sql.ErrNoRows {
		return nil, nil
	} else if e != nil {
		return nil, e
	}
	return res, nil
}

func InsertTag(tag *Tag) (*Tag, error) {
	qs := buildInsertTyped("tag", tag)

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

func UpdateQuestion(q *Question) error {
	const qs = "UPDATE `question` SET `urgent`=:urgent, `private`=:private, `content`=:content WHERE `id`=:id;"
	_, e := db.NamedExec(qs, q)
	return e
}

func ReplyQuestion(q *Question) error {
	const qs = "UPDATE `question` SET `replier`=:replier, `reply`=:reply, `replied_at`=:replied_at WHERE `id`=:id;"
	_, e := db.NamedExec(qs, q)
	return e
}

const sqlSelectQuestion = "SELECT q.*, a.`nick_name` AS `asker_name`, IFNULL(r.`nick_name`,'') AS `replier_name` FROM `question` AS q" +
	" LEFT JOIN `user` AS a ON q.`asker`=a.`id`" +
	" LEFT JOIN `user` AS r ON q.`replier`=r.`id`"

func GetQuestionByID(id int64, wantTags bool) (*Question, error) {
	var q Question
	e := db.Get(&q, sqlSelectQuestion+" WHERE q.`id`=?;", id)
	if e == sql.ErrNoRows {
		return nil, nil
	} else if e != nil {
		return nil, e
	}

	if wantTags {
		// ignore error in tags query
		db.Select(&q.Tags, "SELECT * FROM `tag` WHERE `id` IN (SELECT `tag_id` FROM `question_tag` WHERE `question_id`=?)", id)
	}
	return &q, nil
}

type FindQuestionArg struct {
	UserID        int64 `json:"userId"`
	TagID         int64 `json:"tagId"`
	Urgent        bool  `json:"urgent"`
	PublicOnly    bool  `json:"publicOnly"`
	FeaturedOnly  bool  `json:"featuredOnly"`
	Replied       bool  `json:"replied"`
	Deleted       bool  `json:"deleted"`
	FeaturedFirst bool  `json:"featuredFirst"`
	Ascending     bool  `json:"ascending"`
	Offset        int64 `json:"offset"`
	Count         int64 `json:"count"`
}

type FindQuestionResult struct {
	Total     int64      `json:"total"`
	Questions []Question `json:"questions"`
}

func FindQuestion(fqa *FindQuestionArg) (*FindQuestionResult, error) {
	var args []interface{}
	var wheres []string

	if fqa.UserID > 0 {
		wheres = append(wheres, "q.`asker`=?")
		args = append(args, fqa.UserID)
	}

	if fqa.Urgent {
		wheres = append(wheres, "q.`urgent`<>0")
	}

	if fqa.FeaturedOnly {
		wheres = append(wheres, "q.`featured`<>0")
	}

	if fqa.Replied {
		wheres = append(wheres, "q.`replier`>0")
	} else {
		wheres = append(wheres, "q.`replier`=0")
	}

	if fqa.PublicOnly {
		wheres = append(wheres, "q.`private`=0")
	}

	if fqa.Deleted {
		wheres = append(wheres, "q.`deleted_at`=?")
	} else {
		wheres = append(wheres, "q.`deleted_at`<>?")
	}
	args = append(args, time.Time{})

	if fqa.TagID > 0 {
		wheres = append(wheres, "q.`id` IN (SELECT `question_id` FROM `question_tag` WHERE `tag_id`=?)")
		args = append(args, fqa.TagID)
	}

	where := strings.Join(wheres, " AND ")
	var sb strings.Builder
	sb.WriteString("SELECT COUNT(1) FROM `question` AS q WHERE ")
	sb.WriteString(where)
	sb.WriteByte(';')

	var result FindQuestionResult
	if e := db.Get(&result.Total, sb.String(), args...); e != nil {
		return nil, e
	}

	sb.Reset()
	sb.WriteString(sqlSelectQuestion)
	sb.WriteString(where)

	var orderby string
	if fqa.Replied {
		if fqa.Ascending {
			orderby = "q.`replied_at`"
		} else {
			orderby = "q.`replied_at` DESC"
		}
	} else {
		if fqa.Ascending {
			orderby = "q.`asked_at`"
		} else {
			orderby = "q.`asked_at` DESC"
		}
	}

	if fqa.FeaturedFirst && !fqa.FeaturedOnly {
		orderby = "q.`featured` DESC," + orderby
	}

	sb.WriteString(" ORDER BY ")
	sb.WriteString(orderby)

	if fqa.Count > 0 {
		if fqa.Offset > result.Total {
			return &result, nil
		}
		if fqa.Offset+fqa.Count > result.Total {
			fqa.Count = result.Total - fqa.Offset
		}
		sb.WriteString(" LIMIT ?, ?")
		args = append(args, fqa.Offset, fqa.Count)
	}
	sb.WriteByte(';')

	if e := db.Select(&result.Questions, sb.String(), args...); e != nil {
		return nil, e
	}

	return &result, nil
}

func RemoveQuestion(id int64) error {
	_, e := db.Exec("UPDATE `question` SET `deleted_at`=? WHERE `id`=?;", time.Now(), id)
	return e
}
