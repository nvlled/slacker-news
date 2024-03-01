package hn

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"slices"
	"sort"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	baseURL  = "https://hacker-news.firebaseio.com/v0"
	cacheDir = ".hn-cache"
)

var writeCacheQueue chan *Item

func init() {
	if err := os.RemoveAll(cacheDir); err != nil {
		log.Print(err)
	}
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		panic(err)
	}

	writeCacheQueue = make(chan *Item, 255)
	go func() {
		for item := range writeCacheQueue {
			if err := writeCachedItem(item); err != nil {
				log.Print(err)
			}
		}
	}()
}

func fetchCachedItem(id ItemID) (*Item, error) {
	filename := path.Join(cacheDir, fmt.Sprintf("item-%d.json", id))
	f, err := os.Open(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var item Item
	if err = json.NewDecoder(f).Decode(&item); err != nil {
		return nil, err
	}

	return &item, nil
}

func writeCachedItem(item *Item) error {
	filename := path.Join(cacheDir, fmt.Sprintf("item-%d.json", item.ID))
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	return json.NewEncoder(f).Encode(item)
}

func FetchItem(id ItemID) (*Item, error) {
	cachedItem, err := fetchCachedItem(id)
	if err != nil {
		log.Print(err)
	}
	if cachedItem != nil {
		return cachedItem, nil
	}

	resp, err := http.Get(fmt.Sprintf(baseURL+"/item/%d.json", id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var item Item
	if err = json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, err
	}

        writeCacheQueue <- &item

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
	g := new(errgroup.Group)
	var wg sync.WaitGroup
	var result []*Item
	var mu sync.Mutex
	queue := make(chan ItemID)

	g.SetLimit(20)

	sortMap := map[ItemID]int{}
	counter := 0

	go func() {
		for id := range queue {
			println("fetch", id)
			id := id
			g.Go(func() error {
				defer wg.Done()
				item, err := FetchItem(id)
				if err != nil {
					return err
				}

				mu.Lock()
				result = append(result, item)
				mu.Unlock()

				for _, subID := range item.Kids {
					subID := subID
					wg.Add(1)
					go func() {
						mu.Lock()
						counter++
						sortMap[subID] = counter
						mu.Unlock()
						queue <- subID
					}()
				}

				return nil
			})
		}
	}()

	sortMap[opID] = counter + 1

	op, err := FetchItem(opID)
	if err != nil {
		return nil, err
	}

	for _, subID := range op.Kids {
		wg.Add(1)
		mu.Lock()
		counter++
		sortMap[subID] = counter
		mu.Unlock()
		queue <- subID
	}

	time.Sleep(1 * time.Millisecond)
	if err := g.Wait(); err != nil {
		return nil, err
	}

	wg.Wait()
	close(queue)

	// well, actually I should just sort by chronological order...
	// just like how 4chan orders its posts
	//sort.SliceStable(result, func(i, j int) bool {
	//	c1, ok1 := sortMap[result[i].ID]
	//	c2, ok2 := sortMap[result[j].ID]
	//	if ok1 && ok2 {
	//		return c1 < c2
	//	}
	//	return false
	//})

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Time < result[j].Time
	})

	result = slices.Insert(result, 0, op)

	return result, nil
}

func FetchTopStories(pageSize, pageNum int) ([]*Item, error, bool) {
	resp, err := http.Get(baseURL + "/topstories.json")
	if err != nil {
		return nil, err, false
	}
	defer resp.Body.Close()

	var ids []int64
	if err = json.NewDecoder(resp.Body).Decode(&ids); err != nil {
		return nil, err, false
	}

	totalIDs := len(ids)

	i := pageSize * pageNum
	ids = ids[i:min(i+pageSize, len(ids))]

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
		return nil, err, false
	}

	hasMorePage := totalIDs-i > pageSize

	return result, nil, hasMorePage
}
