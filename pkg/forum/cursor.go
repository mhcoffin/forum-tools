package forum

import (
	"cloud.google.com/go/firestore"
	"time"
)

type Cursor interface {
	value() interface{}
	direction() firestore.Direction
	field() string
	Next(post *Post) Cursor

}

type CreateTimeAsc struct {
	tm time.Time
	fieldName string
}

func (tc *CreateTimeAsc) value() interface{} {
	return tc.tm
}

func (tc *CreateTimeAsc) field() string {
	return "CreateTime"
}

func (tc *CreateTimeAsc) direction() firestore.Direction {
	return firestore.Asc
}

func (tc *CreateTimeAsc) Next(post *Post) Cursor {
	return &CreateTimeAsc{
		tm: post.CreateTime,
		fieldName: "CreateTime",
	}
}

type BumpTimeDesc struct {
	tm time.Time
}

func (tc *BumpTimeDesc) value() interface{} {
	if tc.tm.IsZero() {
		return time.Now()
	}
	return tc.tm
}

func (tc *BumpTimeDesc) field() string {
	return "Bump.Time"
}

func (tc *BumpTimeDesc) Next(post *Post) Cursor {
	return &CreateTimeAsc{
		tm: post.Bump.Time,
	}
}

func (tc *BumpTimeDesc) direction() firestore.Direction {
	return firestore.Desc
}

type IndexAsc struct {
	val int
}

func (i *IndexAsc) value() interface{} {
	return i.val
}

func (i IndexAsc) direction() firestore.Direction {
	return firestore.Asc
}

func (i IndexAsc) field() string {
	return "Index"
}

func (i IndexAsc) Next(post *Post) Cursor {
	return &IndexAsc{
		val: post.Index,
	}
}
