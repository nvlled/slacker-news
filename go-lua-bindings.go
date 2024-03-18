package main

import (
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/nvlled/goom-temple/hn"
)

var gmtLocation *time.Location

func init() {
	loc, err := time.LoadLocation("GMT")
	if err != nil {
		panic(err)
	}
	gmtLocation = loc
}

type GoLuaBindings struct {
	response http.ResponseWriter
	request  *http.Request
}

func NewGoLuaBindings(w http.ResponseWriter, r *http.Request) *GoLuaBindings {
	return &GoLuaBindings{response: w, request: r}
}

func (glb *GoLuaBindings) GetCurrentUser() *User {
	return nil
	//return &User{
	//	ID:       1,
	//	Username: "ronald",
	//}
}

func (glb *GoLuaBindings) GetStories(feed string, pageSize, pageNum int) ([]*hn.Item, string, bool) {
	items, err, hasMore := hn.FetchStories(feed, pageSize, pageNum)
	if err != nil {
		log.Print(err)
		return nil, err.Error(), false
	}
	return items, "", hasMore
}

func (glb *GoLuaBindings) GetThread(id hn.ItemID) ([]*hn.Item, string) {
	return logError(hn.FetchThread(id))
}

func (glb *GoLuaBindings) GetItem(id hn.ItemID) (*hn.Item, string) {
	return logError(hn.FetchItem(id))
}

func (glb *GoLuaBindings) GetCommentChain(id hn.ItemID) ([]*hn.Item, string) {
	return logError(hn.FetchCommentChain(id))
}

func (glb *GoLuaBindings) FormatTime(unixTime int64) string {
	t := time.Unix(unixTime, 0).In(time.UTC)
	return t.Format(time.RFC822)
}

func (glb *GoLuaBindings) ParseURL(rawURL string) (*url.URL, string) {
	return logError(url.Parse(rawURL))
}

func logError[T any](x T, err error) (T, string) {
	var defaultVal T
	if err != nil {
		log.Print(err)
		return defaultVal, err.Error()
	}
	return x, ""
}

func (glb *GoLuaBindings) Redirect(url string) {
	glb.response.Header().Add("Location", url)
	glb.response.WriteHeader(http.StatusFound)
}

func (glb *GoLuaBindings) UnixMilli() int64 {
	return time.Now().UnixMilli()
}
