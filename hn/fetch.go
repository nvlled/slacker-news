package hn

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"golang.org/x/sync/errgroup"
)

const (
	baseURL = "https://hacker-news.firebaseio.com/v0"
)

func FetchItem(id ItemID) (*Item, error) {
	resp, err := http.Get(fmt.Sprintf(baseURL+"/item/%d.json", id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var item Item
	if err = json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, err
	}
	return &item, nil
}

func FetchReplies(parent *Item) ([]*Item, error) {
	g := new(errgroup.Group)
	result := make([]*Item, len(parent.Kids))

	for i, id := range parent.Kids {
		i := i
		id := id
		g.Go(func() error {
			item, err := FetchItem(ItemID(id))
			if err != nil {
				return err
			}
			result[i] = item
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return result, nil
}

func FetchThread(opID ItemID) ([]*Item, error) {
	//g := new(errgroup.Group)
	var wg sync.WaitGroup
	var result []*Item
	var mu sync.Mutex
	queue := make(chan ItemID)

	var lastError error

	go func() {
		for id := range queue {
			id := id
			println("fetch", id)
			go func() {
				defer wg.Done()
				item, err := FetchItem(id)
				if err != nil {
					lastError = err
					return
				}

				if id%2 == 0 {
					lastError = fmt.Errorf("blah")
					return
				}

				println("got", id, "replies: ", len(item.Kids))

				mu.Lock()
				result = append(result, item)
				mu.Unlock()

				for _, subID := range item.Kids {
					wg.Add(1)
					println("<-", subID)
					queue <- subID
				}
			}()
		}
	}()

	op, err := FetchItem(opID)
	if err != nil {
		return nil, err
	}

	result = append(result, op)
	for _, subID := range op.Kids {
		wg.Add(1)
		queue <- subID
	}

	wg.Wait()
	close(queue)

	if lastError != nil {
		return nil, lastError
	}

	return result, nil
}

func FetchTopStories(pageSize, pageNum int) ([]*Item, error) {
	resp, err := http.Get(baseURL + "/topstories.json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var ids []int64
	if err = json.NewDecoder(resp.Body).Decode(&ids); err != nil {
		return nil, err
	}

	i := pageSize * pageNum
	ids = ids[i : i+pageSize]

	g := new(errgroup.Group)
	result := make([]*Item, len(ids))
	for i, id := range ids {
		i := i
		id := id

		g.Go(func() error {
			item, err := FetchItem(ItemID(id))
			if err != nil {
				return err
			}
			result[i] = item
			return nil
		})
	}

	if err = g.Wait(); err != nil {
		return nil, err
	}

	return result, nil
}
