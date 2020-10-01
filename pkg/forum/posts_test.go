package forum

import (
	"context"
	"github.com/mhcoffin/forum-tools/pkg/testutil"
	"github.com/mhcoffin/forum-tools/pkg/uniq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

var ctx Context

func TestMain(m *testing.M) {
	ctx = context.Background()
	testutil.StartFirestoreEmulator(m)
}

func TestCreateClient(t *testing.T) {
	f, err := NewClient(ctx, "fugalist")
	assert.Nil(t, err)
	defer f.expunge(ctx)
}

func TestFBClient_AddPost(t *testing.T) {
	client, err := NewClient(ctx, "fugalist")
	require.Nil(t, err)
	defer client.expunge(ctx)
	root := AddRandomPost(t, client)
	after, err := client.getPost(ctx, root.ID())
	require.Nil(t, err)
	assert.WithinDuration(t, time.Now(), after.CreateTime, time.Second)
	assert.Equal(t, after.CreateTime, after.BumpTime)
	assert.Equal(t, after.CreateTime, after.EditTime)
	assert.Nil(t, after.Deleted)
}

func TestFBClient_AddReply(t *testing.T) {
	client, err := NewClient(ctx, "fugalist")
	require.Nil(t, err)
	defer client.expunge(ctx)
	root := AddRandomPost(t, client)
	reply := AddRandomPost(t, client, root.Path...)
	post, err := client.getPost(ctx, reply.ID())
	require.Nil(t, err)
	assert.Equal(t, root.ID(), post.Parent)

	rootAfterAdd, err := client.getPost(ctx, root.ID())
	require.Nil(t, err)
	assert.Equal(t, post.CreateTime, rootAfterAdd.BumpTime)
	assert.Equal(t, 1, rootAfterAdd.DescendentCount)
	assert.Equal(t, 1, rootAfterAdd.ChildCount)
}

func TestFBClient_AddPostReplies(t *testing.T) {
	client, err := NewClient(ctx, "fugalist")
	require.Nil(t, err)
	defer client.expunge(ctx)
	N := 9
	path := make([]*Post, N)
	path[0] = AddRandomPost(t, client)
	for k := 1; k < N; k++ {
		path[k] = AddRandomPost(t, client, path[k-1].Path...)
	}
	for k := 1; k < N; k++ {
		p, err := client.getPost(ctx, path[k].ID())
		require.Nil(t, err)
		assert.Equal(t, path[k-1].ID(), p.Parent)
	}
	lastPost, err := client.getPost(ctx, path[N-1].ID())
	require.Nil(t, err)
	for k := 0; k < N; k++ {
		p, err := client.getPost(ctx, path[k].ID())
		require.Nil(t, err)
		assert.Equal(t, lastPost.CreateTime, p.BumpTime)
		assert.Equal(t, N-k-1, p.DescendentCount)
		if k == N-1 {
			assert.Equal(t, 0, p.ChildCount)
		} else {
			assert.Equal(t, 1, p.ChildCount)
		}
	}
}

func TestFBClient_GetDirectChildren(t *testing.T) {
	client, err := NewClient(ctx, "fugalist")
	require.Nil(t, err)
	root := AddRandomPost(t, client)
	foo := AddRandomPost(t, client, root.Path...)
	child1 := AddRandomPost(t, client, foo.Path...)
	child2 := AddRandomPost(t, client, foo.Path...)
	child3 := AddRandomPost(t, client, foo.Path...)
	_ = AddRandomPost(t, client, child3.Path...)
	_ = AddRandomPost(t, client, child2.Path...)
	_ = AddRandomPost(t, client, child2.Path...)
	children, _, err := client.getChildren(ctx, foo.ID(), &CreateTimeAsc{}, 10)
	require.Nil(t, err)
	require.Len(t, children, 3)
	assert.Equal(t, child1.ID(), children[0].ID())
	assert.Equal(t, child2.ID(), children[1].ID())
	assert.Equal(t, child3.ID(), children[2].ID())
}

func TestFBClient_GetDirectChildrenPaginated(t *testing.T) {
	client, err := NewClient(ctx, "fugalist")
	require.Nil(t, err)
	root := AddRandomPost(t, client)
	N := 100
	expected := make([]*Post, N)
	for k := 0; k < N; k++ {
		child := AddRandomPost(t, client, root.Path...)
		expected[k] = child
	}

	batch1, cursor, err := client.getChildren(ctx, root.ID(), &CreateTimeAsc{}, 10)
	require.Nil(t, err)
	assert.Len(t, batch1, 10)
	assert.NotNil(t, cursor)
	for k := 0; k < 10; k++ {
		assert.Equal(t, expected[k].Body, batch1[k].Body)
	}

	batch2, cursor, err := client.getChildren(ctx, root.ID(), cursor, 20)
	require.Nil(t, err)
	assert.Len(t, batch2, 20)
	assert.NotNil(t, cursor)
	for k := 0; k < 20; k++ {
		assert.Equal(t, expected[10+k].Body, batch2[k].Body)
	}

	batch3, cursor, err := client.getChildren(ctx, root.ID(), cursor, 1)
	require.Nil(t, err)
	assert.Len(t, batch3, 1)
	assert.NotNil(t, cursor)
	for k := 0; k < 1; k++ {
		assert.Equal(t, expected[30+k].Body, batch3[k].Body)
	}

	batch4, cursor, err := client.getChildren(ctx, root.ID(), cursor, 100)
	require.Nil(t, err)
	assert.Len(t, batch4, 69)
	assert.Nil(t, cursor)
	for k := 0; k < 60; k++ {
		assert.Equal(t, expected[31+k].Body, batch4[k].Body)
	}
}

func TestOrderByUpdate(t *testing.T) {
	client, err := NewClient(ctx, "fugalist")
	require.Nil(t, err)
	defer client.expunge(ctx)
	root := AddRandomPost(t, client)
	thread1 := AddRandomPost(t, client, root.Path...)
	thread2 := AddRandomPost(t, client, root.Path...)
	threads, _, err := client.getChildren(ctx, root.ID(), &BumpTimeDesc{}, 100)
	require.Nil(t, err)
	assert.Equal(t, thread2.ID(), threads[0].ID())
	assert.Equal(t, thread1.ID(), threads[1].ID())

	// bump thread1
	reply := AddRandomPost(t, client, thread1.Path...)
	threads, _, err = client.getChildren(ctx, root.ID(), &BumpTimeDesc{}, 100)
	require.Nil(t, err)
	assert.Equal(t, thread1.ID(), threads[0].ID())

	// bump thread2
	_ = AddRandomPost(t, client, thread2.Path...)
	threads, _, err = client.getChildren(ctx, root.ID(), &BumpTimeDesc{}, 100)
	require.Nil(t, err)
	assert.Equal(t, thread2.ID(), threads[0].ID())

	// bump reply and hence thread1
	_ = AddRandomPost(t, client, reply.Path...)
	threads, _, err = client.getChildren(ctx, root.ID(), &BumpTimeDesc{}, 100)
	require.Nil(t, err)
	assert.Equal(t, thread1.ID(), threads[0].ID())
}

func TestFBClient_DeletePost(t *testing.T) {
	client, err := NewClient(ctx, "fugalist")
	require.Nil(t, err)
	root := AddRandomPost(t, client)
	thread := AddRandomPost(t, client, root.Path...)
	err = client.deletePost(ctx, thread.ID(), "mhc", "because")
	require.Nil(t, err)
	after, err := client.getPost(ctx, thread.ID())
	require.Nil(t, err)
	assert.NotNil(t, after.Deleted)
	assert.Equal(t, "mhc", after.Deleted.Who)
	assert.Equal(t, "because", after.Deleted.Why)
	assert.WithinDuration(t, time.Now(), after.Deleted.When, time.Second)

	children, _, err := client.getChildren(ctx, root.ID(), &CreateTimeAsc{}, 10)
	require.Nil(t, err)
	assert.Len(t, children, 0)
}

func TestFBClient_GetDescendents(t *testing.T) {
	client, err := NewClient(ctx, "fugalist")
	require.Nil(t, err)
	root := AddRandomPost(t, client)
	thread1 := AddRandomPost(t, client, root.Path...)
	thread2 := AddRandomPost(t, client, root.Path...)
	reply1 := AddRandomPost(t, client, thread1.Path...)
	reply2 := AddRandomPost(t, client, thread2.Path...)
	desc, _, err := client.getTree(ctx, root.ID(), &CreateTimeAsc{}, 10)
	require.Nil(t, err)
	assert.Len(t, desc, 5)
	assert.Equal(t, root.ID(), desc[0].ID())
	assert.Equal(t, thread1.ID(), desc[1].ID())
	assert.Equal(t, thread2.ID(), desc[2].ID())
	assert.Equal(t, reply1.ID(), desc[3].ID())
	assert.Equal(t, reply2.ID(), desc[4].ID())
}

func AddRandomPost(t *testing.T, client *Forum, path ...string) *Post {
	docId := uniq.Uniq()
	path = append(path, docId)
	post := &Post{
		Path:              path,
		Index:             100,
		Header:            uniq.Uniq(),
		Body:              uniq.Uniq(),
		CreateTime:        time.Time{},
		BumpTime:          time.Time{},
		EditTime:          time.Time{},
		Author:            uniq.Uniq(),
		AuthorDisplayName: uniq.Uniq(),
		ChildCount:        0,
		DescendentCount:   0,
		ViewCount:         0,
	}
	_, err := client.addPost(ctx, post)
	require.Nil(t, err)
	return post
}
