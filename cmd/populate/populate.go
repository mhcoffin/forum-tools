package main

import (
	"context"
	"fmt"
	"github.com/mhcoffin/forum-tools/pkg/forum"
	"log"
	"math/rand"
	"strings"
	"time"
)

var (
	ctx context.Context
	fm  *forum.Forum
)

var lorem = strings.Split(`lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod 
tempor incididunt ut labore et dolore magna aliqua ut enim ad minim veniam, quis nostrud exercitation 
ullamco laboris nisi ut aliquip ex ea commodo consequat duis aute irure dolor in reprehenderit in 
voluptate velit esse cillum dolore eu fugiat nulla pariatur excepteur sint occaecat cupidatat non 
proident, sunt in culpa qui officia deserunt mollit anim id est laborum`, " ")

var subjects = make(map[string]string)

var mhc = forum.User{
	ID:     "VRf7soDS0BQ6praLnktgJfD5CVa2",
	Name:   "Michael Coffin",
	Joined: time.Now(),
}

var users = []forum.User{
	{
		ID:       "L3XOloruA4P02vJmTKdBkQiklub2",
		Name:     "Test1 Fugalist",
		PhotoURL: "",
		Joined:   time.Now(),
	},
	{
		ID:       "abcdefghijklmnopqrstuvwxyz12",
		Name:     "Test2 Fugalist",
		PhotoURL: "foobar",
		Joined:   time.Now(),
	},
}

func init() {
	ctx = context.Background()
	f, err := forum.NewClient(ctx, "fugalist")
	if err != nil {
		panic(fmt.Errorf("failed to create forum client: %w", err))
	}
	fm = f
}

type Path []string

func main() {
	// Create several sections
	_, err := fm.CreateSection(ctx, "Announcements", "Public announcements", 100, mhc)
	if err != nil {
		log.Fatalf("failed to create section: %s", err)
	}
	syn, err := fm.CreateSection(ctx, "Synchron Libraries", "VSL Synchron Libraries", 200, mhc)
	if err != nil {
		log.Fatalf("failed to create section: %s", err)
	}
	gen, err := fm.CreateSection(ctx, "General Discussion", "General discussion", 300, mhc)

	sections := []Path{syn, gen}

	threads := make([]Path, 0)
	for k := 0; k < 20; k++ {
		section := sections[rand.Intn(len(sections))]
		thread := createRandomThread(section)
		threads = append(threads, thread)
	}

	for k := 0; k < 100; k++ {
		path := threads[rand.Intn(len(threads))]
		p := createRandomReply(path)
		threads = append(threads, p)
	}
}

func createRandomThread(path Path) []string {
	subject := randomString(4)
	p, err := fm.CreateThread(ctx, subject, randomString(400), randomUser(), path[0])
	if err != nil {
		log.Fatalf("failed to create thread: %s", err)
	}
	subjects[p[1]] = subject
	return p
}

func createRandomReply(path Path) []string {
	p, err := fm.CreateReply(ctx, path, subjects[path[1]], randomString(20), randomUser())
	if err != nil {
		log.Fatalf("failed to create reply: %s", err)
	}
	return p
}

func randomString(words int) string {
	res := make([]string, words)
	for k := 0; k < words; k++ {
		res[k] = lorem[rand.Intn(len(lorem))]
	}
	return strings.Join(res, " ")
}

func randomUser() forum.User {
	return users[rand.Intn(len(users))]
}
