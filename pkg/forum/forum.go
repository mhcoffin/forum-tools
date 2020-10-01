package forum

import (
	"fmt"
	"github.com/mhcoffin/forum-tools/pkg/uniq"
	"math"
	"time"
)

const (
	admin        = "mhc"
	adminDisplay = "Mikey"
)

func (f Forum) CreateSection(ctx Context, subject string, description string, index int, uid string) ([]PostID, error) {
	post := &Post{
		Path:              []string{uniq.Uniq()},
		Parent:            "",
		Index:             index,
		Header:            subject,
		Body:              description,
		Author:            uid,
		ChildCount:        0,
		DescendentCount:   0,
		ViewCount:         0,
		Deleted:           nil,
		CreateTime:        time.Time{},
		BumpTime:          time.Time{},
		EditTime:          time.Time{},
	}
	path, err := f.addPost(ctx, post)
	if err != nil {
		return nil, fmt.Errorf("failed to create forum section: %w", err)
	}
	return path, nil
}

func (f Forum) GetSections(ctx Context) ([]*Post, error) {
	posts, _, err := f.getChildren(ctx, "", &IndexAsc{val: math.MinInt32}, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve sections: %w", err)
	}
	return posts, nil
}

func (f Forum) CreateThread(ctx Context, subject string, body string, author string, displayName string, sectionId PostID) ([]PostID, error) {
	post := &Post{
		Path:              []string{sectionId, uniq.Uniq()},
		Header:            subject,
		Body:              body,
		Author:            author,
		AuthorDisplayName: displayName,
		ChildCount:        0,
		DescendentCount:   0,
		ViewCount:         0,
		Deleted:           nil,
		CreateTime:        time.Time{},
		BumpTime:          time.Time{},
		EditTime:          time.Time{},
	}
	path, err := f.addPost(ctx, post)
	if err != nil {
		return nil, fmt.Errorf("failed to create thread: %w", err)
	}
	return path, nil
}

// GetThreads retrieves threads, most-recently-bumped thread first.
func (f Forum) GetThreads(ctx Context, section PostID, cursor Cursor, n int) ([]*Post, Cursor, error) {
	if cursor == nil {
		cursor = &BumpTimeDesc{}
	}
	posts, cursor, err := f.getChildren(ctx, section, &BumpTimeDesc{}, n)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve threads: %w", err)
	}
	return posts, cursor, nil
}

func (f Forum) CreateReply(ctx Context, parent []PostID, body string, author string, displayName string) ([]PostID, error) {
	path := append(parent, uniq.Uniq())
	post := &Post{
		Path:              path,
		Body:              body,
		Author:            author,
		AuthorDisplayName: displayName,
		ChildCount:        0,
		DescendentCount:   0,
		ViewCount:         0,
		Deleted:           nil,
		CreateTime:        time.Time{},
		BumpTime:          time.Time{},
		EditTime:          time.Time{},
	}
	path, err := f.addPost(ctx, post)
	if err != nil {
		return nil, fmt.Errorf("failed to create reply: %w", err)
	}
	return path, nil
}

func (f Forum) GetReplies(ctx Context, thread PostID, cursor Cursor, n int) ([]*Post, error) {
	if cursor == nil {
		cursor = &CreateTimeAsc{}
	}
	posts, cursor, err := f.getTree(ctx, thread, cursor, n)
	if err != nil {
		return nil, fmt.Errorf("failed to get replies: %w", err)
	}
	return posts, nil
}
