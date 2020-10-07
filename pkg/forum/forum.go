package forum

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"github.com/mhcoffin/forum-tools/pkg/uniq"
	"math"
	"time"
)

const (
	admin        = "mhc"
	adminDisplay = "Mikey"
)

func (f Forum) CreateSection(ctx Context, subject string, description string, index int, author User) ([]PostID, error) {
	post := &Post{
		Path:            []string{uniq.Uniq()},
		Parent:          "",
		Index:           index,
		Head:            subject,
		Body:            description,
		Author:          author,
		Bump:            &Bump{Time: time.Time{}},
		ChildCount:      0,
		DescendentCount: 0,
		ViewCount:       0,
		Deleted:         nil,
		CreateTime:      time.Time{},
		EditTime:        time.Time{},
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

func (f Forum) CreateThread(ctx Context, subject string, body string, author User, sectionId PostID) ([]PostID, error) {
	post := &Post{
		Path:            []string{sectionId, uniq.Uniq()},
		Head:            subject,
		Body:            body,
		Author:          author,
		Bump:            &Bump{Time: time.Time{}},
		ChildCount:      0,
		DescendentCount: 0,
		ViewCount:       0,
		Deleted:         nil,
		CreateTime:      time.Time{},
		EditTime:        time.Time{},
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

func (f Forum) CreateReply(ctx Context, parent []PostID, subject string, body string, author User) ([]PostID, error) {
	path := append(parent, uniq.Uniq())
	post := &Post{
		Path:            path,
		Head:            "Re: " + subject,
		Body:            body,
		Author:          author,
		Bump:            &Bump{Time: time.Time{}},
		ChildCount:      0,
		DescendentCount: 0,
		ViewCount:       0,
		Deleted:         nil,
		CreateTime:      time.Time{},
		EditTime:        time.Time{},
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

func (f Forum) DeleteSection(ctx context.Context, sectionID string, user User, reason string) error {
	return f.deletePost(ctx, sectionID, user, reason)
}

func (f Forum) UpdateThread(ctx context.Context, threadID string, subject string, body string) error {
	path := f.fs.Collection(Root).Doc(threadID)
	_, err := path.Update(ctx, []firestore.Update{
		{Path: "Header", Value: subject},
		{Path: "Body", Value: body},
		{Path: "EditTime", Value: firestore.ServerTimestamp},
	})
	if err != nil {
		return fmt.Errorf("failed to update thread: %w", err)
	}
	return err
}

func (f Forum) DeleteThread(ctx context.Context, threadID string, user User, reason string) error {
	return f.deletePost(ctx, threadID, user, reason)
}

func (f Forum) ListThreads(ctx context.Context, sectionID string) ([]*Post, error) {
	posts, _, err := f.getChildren(ctx, sectionID, &BumpTimeDesc{}, 1000)
	if err != nil {
		return nil, err
	}
	return posts, nil
}

func (f Forum) ListReplies(ctx context.Context, threadID string) ([]*Post, error) {
	panic("not implemented")
}

func (f Forum) CreateDraftReply(ctx context.Context, s []string, body string, author string) (string, error) {
	panic("not implemented")
}

func (f Forum) DeleteReply(ctx context.Context, path []string, s string) error {
	panic("not implemented")
}

func (f Forum) UpdateReply(ctx context.Context, replyID string, body string) error {
	panic("not implemented")
}

func (f Forum) InstallReply(ctx context.Context, userID string, docID string) error {
	panic("not implemented")
}
