package hn

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
)

func TestLoopBug(t *testing.T) {
	done := make(chan bool)

	values := []string{"a", "b", "c"}
	for _, v := range values {
		go func() {
			fmt.Println(v)
			done <- true
		}()
	}

	// wait for all goroutines to complete before exiting
	for _ = range values {
		<-done
	}
}

func TestFetch(t *testing.T) {
	resp, err := http.Get("https://hacker-news.firebaseio.com/v0/item/8863.json?print=pretty")
	if err != nil {
		t.Error(err)
	}

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}

	var item Item
	if err := json.Unmarshal(bytes, &item); err != nil {
		t.Error(err)
	}
	bytes, err = json.MarshalIndent(item, "", "  ")
	if err != nil {
		t.Error(err)
	}
	println(string(bytes))
}

func TestFetchTopStories(t *testing.T) {
	items, err := FetchTopStories(10, 1)
	if err != nil {
		t.Error(err)
	}
	for _, item := range items {
		println(">", item.Title)
	}
}

func TestFetchReplies(t *testing.T) {
	item, err := FetchItem(39508046)
	if err != nil {
		t.Error(err)
	}
        println("OP", item.Title, item.Text)
        replies, err := FetchReplies(item)
	if err != nil {
		t.Error(err)
	}
        for _, sub := range replies {
            println(">", sub.Text)
            println("--------------------")
        }
}

func TestFetchThread(t *testing.T) {
        items, err := FetchThread(39511530)
	if err != nil {
		t.Error(err)
	}
        for _, sub := range items {
            println(">", sub.Title, sub.Text)
            println("------------------------------")
        }
        println("reply count", len(items))
}
