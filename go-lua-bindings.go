package main

import (
	"net/http"
	"net/url"
	"time"

	"github.com/nvlled/goom-temple/hn"
)

type GoLuaBindings struct {
	request *http.Request
}

func NewGoLuaBindings(r *http.Request) *GoLuaBindings {
	return &GoLuaBindings{request: r}
}

func (glb *GoLuaBindings) GetCurrentUser() *User {
	return &User{
		ID:       1,
		Username: "ronald",
	}
}

func (glb *GoLuaBindings) GetTopStories(pageSize, pageNum int) ([]*hn.Item, error, bool) {
	return hn.FetchTopStories(pageSize, pageNum)
}

func (glb *GoLuaBindings) GetThread(id hn.ItemID) ([]*hn.Item, error) {
	return hn.FetchThread(id)
}

func (glb *GoLuaBindings) GetItem(id hn.ItemID) (*hn.Item, error) {
	return hn.FetchItem(id)
}

func (glb *GoLuaBindings) FormatTime(unixTime int64) string {
	t := time.Unix(unixTime, 0)
	return t.Format(time.ANSIC)
}

func (glb *GoLuaBindings) ParseURL(rawURL string) (*url.URL, error) {
	return url.Parse(rawURL)
}
