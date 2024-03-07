package main

import (
	"log"
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
	return nil
	//return &User{
	//	ID:       1,
	//	Username: "ronald",
	//}
}

func (glb *GoLuaBindings) GetTopStories(pageSize, pageNum int) ([]*hn.Item, string, bool) {
	items, err, hasMore := hn.FetchTopStories(pageSize, pageNum)
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
	t := time.Unix(unixTime, 0)
	return t.Format(time.ANSIC)
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
