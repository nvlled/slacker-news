package main

import (
	"net/http"

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

func (glb *GoLuaBindings) GetTopStories(pageSize, pageNum int) ([]*hn.Item, error) {
	return hn.FetchTopStories(pageSize, pageNum)
}

func (glb *GoLuaBindings) GetThread() *Thread {
	return nil
}
