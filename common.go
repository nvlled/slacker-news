package main

import "time"

type User struct {
	ID       int64
	Username string
}

type FeedItem struct {
	Title        string
	URL          string
	Submitted    time.Time
	Username     string
	CommentCount int
	Points       int
}

type Post struct {
	ID        int64
	ThreadID  int64
	Username  string
	Points    int
	Submitted time.Time
}

type Thread struct {
	OP       *Post
	Comments []Post
}
