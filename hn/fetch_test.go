package hn

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)

type X struct {
}

func (x *X) Foo() (*Item, error) {
	return FetchItem(39533761)
}

func TestFetchAndParseHTML(t *testing.T) {
	os.Chdir("..")

	L := lua.NewState()
	err := L.DoString(`
            package.path = "lua/?.lua"
        `)
	if err != nil {
		t.Error(err)
	}

	L.SetGlobal("x", luar.New(L, new(X)))

	err = L.DoString(`
            require'html'
            return DIV {
                DIV{__noHTMLEscape=true, x:Foo().Text}
            }
        `)
	if err != nil {
		t.Error(err)
	}

	lv := L.Get(-1)
	if err := L.CallByParam(lua.P{
		Fn:      L.GetGlobal("tostring"),
		NRet:    1,
		Protect: true,
	}, lv); err != nil {
		panic(err)
	}
	ret := L.Get(-1)
	println(ret.String())
}

func TestSliceInsertWhileIter(t *testing.T) {
	xs := []int{1, 2, 3, 4}
	for _, x := range xs {
		if x == 2 {
			xs = append(xs, 21, 22, 23)
		}
		if x == 4 {
			xs = append(xs, 41, 42, 43)
		}
	}
	for _, x := range xs {
		println(">", x)
	}
}

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
	items, err, _ := FetchTopStories(10, 1)
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
	items, err := FetchThread(39511714)
	if err != nil {
		t.Error(err)
	}
	for _, sub := range items {
		println(">", sub.Title, sub.Text)
		println("------------------------------")
	}
	println("reply count", len(items))
}

func TestFetchCommentChain(t *testing.T) {
	items, err := FetchCommentChain(39597030)
	if err != nil {
		t.Error(err)
	}
	for _, sub := range items {
            println(">", sub.By, ":", sub.Text[0:min(50, len(sub.Text))])
		println("------------------------------")
	}
}
