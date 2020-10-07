package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mhcoffin/forum-tools/pkg/forum"
)

/*
Command line tool for dealing with the forum.

forum section create
forum section list
forum section update
forum section delete

forum thread draft
forum thread create
forum thread list
forum thread update
forum thread delete

forum reply list
forum reply create
forum reply update
forum reply delete

args:
	-f sectionId
	-t topicID
	-r replyID
	-s subject
	-b body
	-u uid
*/

var (
	section       = flag.NewFlagSet("section", flag.ExitOnError)
	createSection = section.Bool("create", false, "create")
	listSection   = section.Bool("list", false, "read")
	updateSection = section.Bool("update", false, "update")
	deleteSection = section.Bool("delete", false, "delete")
	sectionTitle  = section.String("title", "", "title")
	sectionIndex  = section.Int("index", -1, "index")
	sectionUid    = section.String("uid", "", "user ID")
	sectionID     = section.String("id", "", "section ID")
	sectionReason = section.String("reason", "", "delete reason")

	thread            = flag.NewFlagSet("thread", flag.ExitOnError)
	threadCreate      = thread.Bool("create", false, "create a thread")
	threadList        = thread.Bool("list", false, "list threads")
	threadUpdate      = thread.Bool("update", false, "update thread")
	threadDelete      = thread.Bool("delete", false, "delete thread")
	threadSection     = thread.String("section", "", "section the thread belongs to")
	threadSubject     = thread.String("subject", "", "thread subject")
	threadBody        = thread.String("body", "", "thread body")
	threadUid         = thread.String("uid", "", "author or thread")
	threadDisplayName = thread.String("display", "", "display name of poster")
	threadReason      = thread.String("reason", "", "reason for delete")
	threadID          = thread.String("id", "", "ID of thread")

	reply            = flag.NewFlagSet("reply", flag.ExitOnError)
	replyCreate      = reply.Bool("create", false, "create a reply")
	replyList        = reply.Bool("list", false, "list replies")
	replyUpdate      = reply.Bool("update", false, "update reply")
	replyDelete      = reply.Bool("delete", false, "delete reply")
	replyExpunge     = reply.Bool("expunge", false, "expunge reply")
	replyHeader      = reply.String("subject", "", "Subject of thread")
	replyBody        = reply.String("body", "", "body of reply")
	replyUid         = reply.String("uid", "", "user ID of author")
	replyDisplayName = reply.String("display", "", "display name of poster")
	replyPath        = reply.String("path", "", "parent path")

	sectionId = flag.String("f", "", "section ID")
	threadId  = flag.String("t", "", "thread ID")
	replyId   = flag.String("r", "", "reply ID")
	subject   = flag.String("s", "", "subject")
	uid       = flag.String("u", "", "uid")
	body      = flag.String("b", "", "body")
)

var (
	ctx context.Context
	fm  *forum.Forum
)

func init() {
	ctx = context.Background()
	f, err := forum.NewClient(ctx, "fugalist")
	if err != nil {
		panic(fmt.Errorf("failed to create forum client: %w", err))
	}
	fm = f
}

func main() {
	flag.Parse()
	switch os.Args[1] {
	case "section":
		Section()
	case "thread":
		Thread()
	case "reply":
		Replies()
	default:
		log.Fatalf("No such subcommand: %s\n", flag.Arg(0))
	}
}

func Section() {
	err := section.Parse(os.Args[2:])
	if err != nil {
		log.Fatalf("failed to parse section flags: %s", err)
	}
	switch {
	case *createSection:
		CreateSection()
	case *listSection:
		ListSections()
	case *updateSection:
		UpdateSection()
	case *deleteSection:
		DeleteSection()
	default:
		log.Fatalf("no such subcommand: %s", flag.Arg(1))
	}
}

func DeleteSection() {
	if *sectionID == "" || *sectionUid == "" || *sectionReason == "" {
		log.Fatal("-id, -uid, and -reason required")
	}
	err := fm.DeleteSection(ctx, *sectionID, *sectionUid, *sectionReason)
	if err != nil {
		log.Fatal(err)
	}
}

func UpdateSection() {
	if *sectionId == "" {
		log.Fatal("-f required")
	}
	if *subject == "" && *body == "" && *sectionIndex == -1 {
		log.Fatal("-s, -b, or -i required")
	}
	panic("implement")
}

func CreateSection() {
	if *sectionTitle == "" || *sectionIndex == -1 || *sectionUid == "" {
		log.Fatal("-title, -index, and -uid required")
	}
	id, err := fm.CreateSection(ctx, *sectionTitle, "", *sectionIndex, *sectionUid)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(id)
}

func ListSections() {
	posts, err := fm.GetSections(ctx)
	if err != nil {
		panic(err)
	}
	for _, post := range posts {
		fmt.Printf("%s %s\n", post.ID(), post.Header)
	}
}

func Replies() {
	err := reply.Parse(os.Args[2:])
	if err != nil {
		log.Fatalf("failed to parse reply flags: %s", err)
	}
	switch {
	case *replyCreate:
		CreateReply()
	case *replyList:
		ListReplies()
	case *replyUpdate:
		UpdateReply()
	case *replyDelete:
		DeleteReply()
	default:
		log.Fatalf("No such subcommand: %s", flag.Arg(1))
	}
}

func CreateReply() {
	if *replyUid == "" || *replyBody == "" || *replyDisplayName == "" || *replyPath == ""  || *replyHeader == ""{
		log.Fatal("-uid, -body, -display, -path required")
	}
	_, err := fm.CreateReply(ctx, strings.Split(*replyPath, "/"), *replyHeader, *replyBody, *replyUid, *replyDisplayName)
	if err != nil {
		log.Fatal(fmt.Errorf("create reply failed: %w", err))
	}
}

func UpdateReply() {
	if *sectionId == "" || *threadId == "" || *replyId == "" || *body == "" {
		log.Fatal("-f,-t, -r, and -b required")
	}
	err := fm.UpdateReply(ctx, *replyId, *body)
	if err != nil {
		log.Fatal(err)
	}
}

func DeleteReply() {
	if *sectionId == "" || *threadId == "" || *replyId == "" || *uid == "" {
		log.Fatal("-f,-t, -u, and -r required")
	}
	err := fm.DeleteReply(ctx, []string{}, *uid)
	if err != nil {
		log.Fatal(err)
	}
}

func ListReplies() {
	if *threadId == "" {
		log.Fatal("-f and -t required")
	}
	replies, err := fm.ListReplies(ctx, *threadId)
	if err != nil {
		log.Fatal(err)
	}
	s, err := json.MarshalIndent(replies, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(s))
}

func Thread() {
	err := thread.Parse(os.Args[2:])
	if err != nil {
		log.Fatalf("failed to thread section flags: %s", err)
	}
	switch {
	case *threadCreate:
		CreateThread()
	case *threadList:
		ListThreads()
	case *threadUpdate:
		UpdateThread()
	case *threadDelete:
		DeleteThread()
	default:
		log.Fatalf("No such subcommand: %s", flag.Arg(1))
	}
}

func CreateThread() {
	if *threadSection == "" || *threadSubject == "" || *threadBody == "" || *threadUid == "" || *threadDisplayName == "" {
		log.Fatal("-section, -subject, -body, and -uid required")
	}
	hash, err := fm.CreateThread(ctx, *threadSubject, *threadBody, *threadUid, *threadDisplayName, *threadSection)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to create draft: %w", err))
	}
	fmt.Println(hash)
}

func UpdateThread() {
	if *sectionId == "" || *threadId == "" || *body == "" {
		log.Fatal("-f and -b are required")
	}
	err := fm.UpdateThread(ctx, *sectionId, *threadId, *body)
	if err != nil {
		log.Fatal(err)
	}
}

func DeleteThread() {
	if *threadID == "" || *threadUid == "" || *threadReason == "" {
		log.Fatal("-id, -uid and -reason are required")
	}
	err := fm.DeleteThread(ctx, *threadID, *threadUid, *threadReason)
	if err != nil {
		log.Fatal(err)
	}
}

func ListThreads() {
	if *threadSection == "" {
		log.Fatal("-section required")
	}
	topics, err := fm.ListThreads(ctx, *threadSection)
	if err != nil {
		log.Fatal(err)
	}
	for _, topic := range topics {
		fmt.Printf("%s %s\n", strings.Join(topic.Path, "/"), topic.Header)
	}
}
