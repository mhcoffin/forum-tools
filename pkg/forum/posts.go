package forum

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"time"
)

type Context = context.Context

const (
	MaxDepth = 100
	Root     = "Posts"
)

type PostID = string
type DeleteInfo struct {
	When time.Time
	Who  string
	Why  string
}
type Post struct {
	Path              []PostID // Path to this post, from root down.
	Index             int      // For explicit ordering
	Parent            PostID   // ID of the parent of this post (same as next-to-last element of Path)
	Header            string   // Subject or summary of post
	Body              string   // Body of post (HTML)
	Author            string   // ID of author
	AuthorDisplayName string   // Display Name of author
	ChildCount        int      // Number of direct children
	DescendentCount   int      // Number of direct and indirect children
	ViewCount         int      // Number of times this post has been viewed
	Deleted           *DeleteInfo
	CreateTime        time.Time `firestore:",serverTimestamp"` // Time this post was created.
	BumpTime          time.Time `firestore:",serverTimestamp"` // Last time a descendant has been added or modified
	EditTime          time.Time `firestore:",serverTimestamp"` // Last time the header or body were edited
}

func (p *Post) ID() PostID {
	return p.Path[len(p.Path)-1]
}

type Order struct {
	Field     string
	Direction firestore.Direction
}

type Forum struct {
	fs *firestore.Client
}

// NewClient returns a new forum client
func NewClient(ctx Context, projectId string) (*Forum, error) {
	client, err := firestore.NewClient(ctx, projectId)
	if err != nil {
		return nil, fmt.Errorf("failed to create forum client: %w", err)
	}
	return &Forum{client}, nil
}

// addPost adds a post to the forum and updates the parents.
func (f Forum) addPost(ctx Context, post *Post) ([]PostID, error) {
	depth := len(post.Path)
	if depth > MaxDepth {
		post.Path[MaxDepth-1] = post.Path[depth-1]
		post.Path = post.Path[:MaxDepth]
	}
	switch depth {
	case 0:
		return nil, fmt.Errorf("empty path in addPost")
	case 1:
		post.Parent = ""
	default:
		post.Parent = post.Path[len(post.Path)-2]
	}
	wb := f.fs.Batch()
	for k := 0; k < depth-1; k++ {
		doc := f.fs.Collection(Root).Doc(post.Path[k])
		updates := []firestore.Update{
			{Path: "DescendentCount", Value: firestore.Increment(1)},
			{Path: "BumpTime", Value: firestore.ServerTimestamp},
		}
		if k == depth-2 {
			updates = append(updates, firestore.Update{Path: "ChildCount", Value: firestore.Increment(1)})
		}
		wb.Update(doc, updates)
	}
	wb.Create(f.fs.Collection(Root).Doc(post.Path[depth-1]), post)

	_, err := wb.Commit(ctx)
	if err != nil {
		return nil, err
	}
	return post.Path, nil
}

func (f Forum) getPost(ctx Context, postID string) (*Post, error) {
	path := f.fs.Collection(Root).Doc(postID)
	doc, err := path.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read post: %w", err)
	}
	result := &Post{}
	err = doc.DataTo(result)
	if err != nil {
		return nil, fmt.Errorf("failed to decode post: %w", err)
	}
	return result, nil
}

// getChildren returns direct children, paginated, newest first.
func (f Forum) getChildren(ctx Context, parent PostID, cursor Cursor, n int) ([]*Post, Cursor, error) {
	if cursor == nil {
		panic("nil cursor")
	}
	query := f.fs.
		Collection(Root).
		Where("Parent", "==", parent).
		Where("Deleted", "==", nil).
		OrderBy(cursor.field(), cursor.direction()).
		StartAfter(cursor.value()).
		Limit(n)
	return f.performQuery(ctx, query, cursor, n)
}

// getTree returns the parent and all descendents.
func (f Forum) getTree(ctx Context, parent PostID, cursor Cursor, n int) ([]*Post, Cursor, error) {
	if cursor == nil {
		panic("nil cursor")
	}
	query := f.fs.
		Collection(Root).
		Where("Path", "array-contains", parent).
		Where("Deleted", "==", nil).
		OrderBy(cursor.field(), cursor.direction()).
		StartAfter(cursor.value()).
		Limit(n)
	return f.performQuery(ctx, query, cursor, n)
}

// expunge deletes all posts. Mostly useful for testing
func (f Forum) expunge(ctx Context) {
	path := f.fs.Collection(Root)
	docs, err := path.Documents(ctx).GetAll()
	if err != nil {
		panic("failed to expunge posts")
	}
	count := 0
	for _, doc := range docs {
		_, err = f.fs.Collection(Root).Doc(doc.Ref.ID).Delete(ctx)
		if err != nil {
			count++
		}
	}
	if count > 0 {
		panic("failed to expunge some posts")
	}
}

func (f Forum) performQuery(ctx Context, query firestore.Query, cursor Cursor, n int) ([]*Post, Cursor, error) {
	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve children: %w", err)
	}
	result := make([]*Post, len(docs))
	for k, doc := range docs {
		post := &Post{}
		err = doc.DataTo(post)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to decode childrenL: %w", err)
		}
		result[k] = post
	}
	if len(result) == n {
		return result, cursor.Next(result[len(docs)-1]), nil
	} else {
		return result, nil, nil
	}
}

// deletePost marks a post deleted. It does not actually delete the post or any children.
func (f Forum) deletePost(ctx Context, postID PostID, who string, why string) error {
	fmt.Println(postID)
	path := f.fs.Collection(Root).Doc(postID)
	fmt.Println(path.Path)
	_, err := path.Update(ctx, []firestore.Update{
		{Path: "Deleted.Who", Value: who},
		{Path: "Deleted.Why", Value: why},
		{Path: "Deleted.When", Value: firestore.ServerTimestamp},
	})
	if err != nil {
		return fmt.Errorf("failed to delete post %s: %w", postID, err)
	}
	return nil
}

func (f Forum) expungePost(ctx Context, postId PostID) error {
	path := f.fs.Collection(Root).Doc(postId)
	_, err := path.Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to expunge doc %s: %w", postId, err)
	}
	return nil
}

func (f Forum) DeleteSection(ctx context.Context, sectionID string, uid string, reason string) error {
	return f.deletePost(ctx, sectionID, uid, reason)
}

// CreateDraftThread puts a draft thread in the database, returning the doc ID and an error.
func (f Forum) CreateDraftThread(ctx context.Context, sectionId string, subject string, body string, authorID string) (string, error) {
	return "", nil
	// TODO
}

func (f Forum) UpdateThread(ctx context.Context, sectionID string, threadID string, body string) error {
	return nil
	// TODO
}

func (f Forum) DeleteThread(ctx context.Context, sectionID string, threadID string, userID string) error {
	// TODO
	return nil
}

func (f Forum) InstallThread(ctx context.Context, sectionID string, author string, threadID string) error {
	return nil
	// TODO
}

func (f Forum) ListThreads(ctx context.Context, sectionID string) ([]*Post, error) {
	path := f.fs.Collection(Root).Where("Parent", "==", sectionID)
	docs, err := path.Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}
	result := make([]*Post, len(docs))
	for k, doc := range docs {
		p := &Post{}
		err = doc.DataTo(p)
		if err != nil {
			return nil, fmt.Errorf("failed to decode thread doc: %w", err)
		}
		result[k] = p
	}
	return result, nil
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
