package forum

import (
	"github.com/mhcoffin/forum-tools/pkg/uniq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

var mhc = User{
	ID: "mhcoffin",
	Name: "Mike Coffin",
	PhotoURL: "https://foo.bar.photo.jpeg",
	Joined: time.Now(),
}

var ella = User{
	ID: "jane",
	Name: "Ella Fitzgerald",
	PhotoURL: "https://fitz.photo.jpeg",
	Joined: time.Now(),
}

func TestClient_CreateSection(t *testing.T) {
	f, err := NewClient(ctx, "fugalist")
	require.Nil(t, err)
	defer f.expunge(ctx)

	path, err := f.CreateSection(ctx, "Announcements", "Important stuff", 0, mhc)
	require.Nil(t, err)
	s, err := f.GetSections(ctx)
	require.Nil(t, err)
	assert.Len(t, s, 1)
	assert.Equal(t, path[0], s[0].ID())
}

func TestForum_GetSections(t *testing.T) {
	f, err := NewClient(ctx, "fugalist")
	require.Nil(t, err)
	defer f.expunge(ctx)
	_, err = f.CreateSection(ctx, "Announcements", "Important stuff", 100, mhc)
	require.Nil(t, err)
	_, err = f.CreateSection(ctx, "Discussion", "Random stuff", 200, mhc)
	require.Nil(t, err)
	s, err := f.GetSections(ctx)
	require.Nil(t, err)
	require.Len(t, s, 2)
	assert.Equal(t, "Announcements", s[0].Head)
	assert.Equal(t, "Discussion", s[1].Head)
}

func TestForum_CreateThread(t *testing.T) {
	f, err := NewClient(ctx, "fugalist")
	require.Nil(t, err)
	defer f.expunge(ctx)
	ann, err := f.CreateSection(ctx, "Announcements", "Important stuff", 100, mhc)
	require.Nil(t, err)
	_, err = f.CreateThread(ctx, "Hi", "First Post!", mhc, ann[0])
	assert.Nil(t, err)
}

func TestForum_CreateThreads(t *testing.T) {
	f, err := NewClient(ctx, "fugalist")
	require.Nil(t, err)
	defer f.expunge(ctx)
	ann, err := f.CreateSection(ctx, "Announcements", "Important stuff", 100, mhc)
	require.Nil(t, err)
	require.Len(t, ann, 1)
	for k := 0; k < 100; k++ {
		createRandomThread(t, ctx, f, ann[0])
	}
	threads, cursor, err := f.GetThreads(ctx, ann[0], &CreateTimeAsc{}, 120)
	require.Nil(t, err)
	assert.Nil(t, cursor)
	assert.Len(t, threads, 100)
}

func TestForum_CreateReply(t *testing.T) {
	f, err := NewClient(ctx, "fugalist")
	require.Nil(t, err)
	defer f.expunge(ctx)
	ann, err := f.CreateSection(ctx, "Announcements", "Important stuff", 100, mhc)
	require.Nil(t, err)
	hello, err := f.CreateThread(ctx, "Hello", "First post", mhc, ann[0])
	require.Nil(t, err)
	_, err = f.CreateReply(ctx, hello, "reply body", "bob", ella)
	assert.Nil(t, err)
}

func createRandomThread(t *testing.T, ctx Context, forum *Forum, section PostID) {
	_, err := forum.CreateThread(ctx, uniq.Uniq(), uniq.Uniq(), ella, section)
	require.Nil(t, err)
}
