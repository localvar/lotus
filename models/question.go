package models

import (
	"database/sql"
	"strconv"
	"strings"
	"time"
)

type Tag struct {
	ID        int64     `db:"id" json:"id" dbx:"<-"`
	Name      string    `db:"name" json:"name"`
	Color     string    `db:"color" json:"color"`
	CreatedAt time.Time `db:"created_at" json:"-"`
	CreatedBy int64     `db:"created_by" json:"-"`
}

type Question struct {
	ID            int64     `db:"id" json:"id" dbx:"<-"`
	Urgent        bool      `db:"urgent" json:"urgent"`
	Private       bool      `db:"private" json:"private"`
	Featured      bool      `db:"featured" json:"featured"`
	Asker         int64     `db:"asker" json:"asker"`
	AskerName     string    `db:"asker_name" json:"askerName" dbx:"<-"`
	AskerAvatar   string    `db:"asker_avatar" json:"askerAvatar" dbx:"<-"`
	AskedAt       time.Time `db:"asked_at" json:"askedAt"`
	Content       string    `db:"content" json:"content"`
	Reply         string    `db:"reply" json:"reply"`
	Replier       int64     `db:"replier" json:"replier"`
	ReplierName   string    `db:"replier_name" json:"replierName" dbx:"<-"`
	ReplierAvatar string    `db:"replier_avatar" json:"replierAvatar" dbx:"<-"`
	RepliedAt     time.Time `db:"replied_at" json:"repliedAt"`
	DeletedAt     time.Time `db:"deleted_at" json:"-"`
	Tags          []int64   `db:"-" json:"tags,omitempty"`
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

func UpdateTag(tag *Tag) error {
	const qs = "UPDATE `tag` SET `name`=:name, `color`=:color WHERE `id`=:id;"
	_, e := db.NamedExec(qs, tag)
	return e

}

func DeleteTag(id int64) error {
	tx, e := db.Beginx()
	if e != nil {
		return e
	}
	defer tx.Rollback()

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

func SetQuestionTag(uid, qid int64, added, removed []int64) error {
	tx, e := db.Beginx()
	if e != nil {
		return e
	}
	defer tx.Rollback()

	qs := "DELETE FROM `question_tag` WHERE `question_id`=? AND `tag_id`=?;"
	for _, tid := range removed {
		if _, e := db.Exec(qs, qid, tid); e != nil {
			return e
		}
	}

	qt := &QuestionTag{
		QuestionID: qid,
		TaggedBy:   uid,
		TaggedAt:   time.Now(),
	}
	qs = buildInsertTyped("question_tag", qt)
	for _, tid := range added {
		qt.TagID = tid
		if _, e := db.NamedExec(qs, qt); e != nil {
			return e
		}
	}

	return tx.Commit()
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
	tx, e := db.Beginx()
	if e != nil {
		return e
	}
	defer tx.Rollback()

	qs := "UPDATE `question` SET `urgent`=:urgent, `replier`=:replier, `reply`=:reply, `replied_at`=:replied_at WHERE `id`=:id;"
	_, e = tx.NamedExec(qs, q)
	if e != nil {
		return e
	}

	_, e = tx.Exec("DELETE FROM `question_tag` WHERE `question_id`=?;", q.ID)
	if e != nil {
		return e
	}

	qt := &QuestionTag{
		QuestionID: q.ID,
		TaggedBy:   q.Replier,
		TaggedAt:   time.Now(),
	}
	qs = buildInsertTyped("question_tag", qt)
	for _, tid := range q.Tags {
		qt.TagID = tid
		if _, e := tx.NamedExec(qs, qt); e != nil {
			return e
		}
	}

	return tx.Commit()
}

func SetQuestionFlag(id int64, flag string, value bool) error {
	switch flag {
	default:
		return nil
	case "featured", "urgent", "private":
	}
	qs := "UPDATE `question` SET `" + flag + "`=? WHERE `id`=?;"
	_, e := db.Exec(qs, value, id)
	return e
}

const sqlSelectQuestion = "SELECT q.*," +
	" IFNULL(a.`nick_name`,'匿名用户') AS `asker_name`," +
	" IFNULL(a.`avatar`,'') AS `asker_avatar`," +
	" IFNULL(r.`nick_name`,'') AS `replier_name`," +
	" IFNULL(r.`avatar`,'') AS `replier_avatar`" +
	" FROM `question` AS q" +
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
		db.Select(&q.Tags, "SELECT `tag_id` FROM `question_tag` WHERE `question_id`=?", id)
	}
	return &q, nil
}

func GetUserLastQuestion(uid int64) (*Question, error) {
	const qs = "SELECT * FROM `question` WHERE `asker`=? ORDER BY `asked_at` DESC LIMIT 1;"
	var q Question
	if e := db.Get(&q, qs, uid); e == sql.ErrNoRows {
		return nil, nil
	} else if e != nil {
		return nil, e
	}
	return &q, nil
}

type FindQuestionArg struct {
	Asker      int64  `json:"asker"`
	Replier    int64  `json:"replier"`
	Replied    string `json:"replied"`
	Urgent     string `json:"urgent"`
	Featured   string `json:"featured"`
	Private    string `json:"private"`
	Tag        int64  `json:"tag"`
	Deleted    bool   `json:"deleted"`
	Ascending  bool   `json:"ascending"`
	WantTags   bool   `json:"wantTags"`
	PageSize   uint32 `json:"pageSize"`
	PageNumber uint32 `json:"pageNumber"`
}

type FindQuestionResult struct {
	Total      uint32     `json:"total"`
	PageSize   uint32     `json:"pageSize"`
	PageNumber uint32     `json:"pageNumber"`
	Questions  []Question `json:"questions"`
}

func FindQuestion(fqa *FindQuestionArg) (*FindQuestionResult, error) {
	var args []interface{}
	var wheres []string
	var sb strings.Builder
	var result FindQuestionResult

	if fqa.PageSize == 0 {
		fqa.PageSize = 1
	}
	result.PageSize = fqa.PageSize

	tx, e := db.Beginx()
	if e != nil {
		return nil, e
	}
	defer tx.Rollback()

	if fqa.Asker > 0 {
		wheres = append(wheres, "q.`asker`=?")
		args = append(args, fqa.Asker)
	}

	if fqa.Replier > 0 {
		wheres = append(wheres, "q.`replier`=?")
		args = append(args, fqa.Replier)
	} else if fqa.Replied == "yes" {
		wheres = append(wheres, "q.`replier`>0")
	} else if fqa.Replied == "no" {
		wheres = append(wheres, "q.`replier`=0")
	}

	if fqa.Urgent == "yes" {
		wheres = append(wheres, "q.`urgent`<>0")
	} else if fqa.Urgent == "no" {
		wheres = append(wheres, "q.`urgent`=0")
	}

	if fqa.Featured == "yes" {
		wheres = append(wheres, "q.`featured`<>0")
	} else if fqa.Featured == "no" {
		wheres = append(wheres, "q.`featured`=0")
	}

	if fqa.Private == "yes" {
		wheres = append(wheres, "q.`private`<>0")
	} else if fqa.Private == "no" {
		wheres = append(wheres, "q.`private`=0")
	}

	if fqa.Deleted {
		wheres = append(wheres, "q.`deleted_at`>'2000-01-01'?")
	} else {
		wheres = append(wheres, "q.`deleted_at`<'2000-01-01'")
	}

	if fqa.Tag > 0 {
		wheres = append(wheres, "q.`id` IN (SELECT `question_id` FROM `question_tag` WHERE `tag_id`=?)")
		args = append(args, fqa.Tag)
	}

	sb.WriteString("SELECT COUNT(1) FROM `question` AS q")
	where := " WHERE " + strings.Join(wheres, " AND ")
	sb.WriteString(where)
	sb.WriteByte(';')

	if e := tx.Get(&result.Total, sb.String(), args...); e != nil {
		return nil, e
	} else if result.Total == 0 {
		return &result, nil
	}

	if fqa.PageSize*fqa.PageNumber >= result.Total {
		fqa.PageNumber = (result.Total+fqa.PageSize-1)/fqa.PageSize - 1
	}

	sb.Reset()
	sb.WriteString(sqlSelectQuestion)
	sb.WriteString(where)

	var orderby string
	if fqa.Replier > 0 || fqa.Replied == "yes" {
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

	if fqa.Featured == "first" {
		orderby = "q.`featured` DESC," + orderby
	}

	if fqa.Urgent == "first" {
		orderby = "q.`urgent` DESC," + orderby
	}

	sb.WriteString(" ORDER BY ")
	sb.WriteString(orderby)

	sb.WriteString(" LIMIT ?, ?")
	args = append(args, fqa.PageNumber*fqa.PageSize, fqa.PageSize)
	sb.WriteByte(';')

	if e := tx.Select(&result.Questions, sb.String(), args...); e != nil {
		return nil, e
	}

	result.PageNumber = fqa.PageNumber

	if fqa.WantTags && len(result.Questions) > 0 {
		ids := make([]string, 0, len(result.Questions))
		qm := map[int64]*Question{}
		for i := 0; i < len(result.Questions); i++ {
			q := &result.Questions[i]
			qm[q.ID] = q
			ids = append(ids, strconv.FormatInt(q.ID, 10))
		}

		var tags []QuestionTag
		in := strings.Join(ids, ",")
		tx.Select(&tags, "SELECT * FROM `question_tag` WHERE `question_id` in ("+in+");")

		for i := 0; i < len(tags); i++ {
			t := &tags[i]
			q := qm[t.QuestionID]
			q.Tags = append(q.Tags, t.TagID)
		}
	}

	return &result, nil
}

func RemoveQuestion(id int64) error {
	_, e := db.Exec("UPDATE `question` SET `deleted_at`=? WHERE `id`=?;", time.Now(), id)
	return e
}
