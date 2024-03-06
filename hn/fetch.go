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

	"golang.org/x/sync/errgroup"
)

const (
	baseURL  = "https://hacker-news.firebaseio.com/v0"
	CacheDir = ".hn-cache"
)

var writeCacheQueue chan *Item

func init() {
	//if err := os.RemoveAll(CacheDir); err != nil {
	//	log.Print(err)
	//}
	if err := os.MkdirAll(CacheDir, 0755); err != nil {
		panic(err)
	}

	//writeCacheQueue = make(chan *Item, 255)
	//go func() {
	//	for item := range writeCacheQueue {
	//		if err := writeCachedItem(item); err != nil {
	//			log.Print(err)
	//		}
	//	}
	//}()
}

func fetchCachedItem(id ItemID) (*Item, error) {
	filename := path.Join(CacheDir, fmt.Sprintf("item-%d.json", id))
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
	filename := path.Join(CacheDir, fmt.Sprintf("item-%d.json", item.ID))
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

	//sort.SliceStable(item.Kids, func(i, j int) bool {
	//	return item.Kids[i] < item.Kids[j]
	//})

	//writeCacheQueue <- &item
	go func() {
		if err := writeCachedItem(&item); err != nil {
			log.Print(err)
		}
	}()

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

func FetchCommentChain(opID ItemID) ([]*Item, error) {
	var result []*Item
	maxDepth := 1000

	currentID := opID
	for i := 0; i < maxDepth; i++ {
		if currentID <= 0 {
			break
		}
		item, err := FetchItem(currentID)
		if err != nil {
			return nil, err
		}
		if item == nil {
			break
		}
		currentID = item.Parent
		result = append(result, item)
	}

	slices.Reverse(result)

	return result, nil
}

func FetchThread(opID ItemID) ([]*Item, error) {
	var wg sync.WaitGroup
	var result []*Item
	var mu sync.Mutex

	type Entry struct {
		id    ItemID
		level uint
	}
	queue := make(chan Entry)

	var fetchError error
	go func() {
		for entry := range queue {
			id, level := entry.id, entry.level
			println("fetch", id)
			go func() {
				defer wg.Done()
				item, err := FetchItem(id)
				if err != nil {
					fetchError = err
					return
				}

				mu.Lock()
				item.Level = level
				result = append(result, item)
				mu.Unlock()

				for _, subID := range item.Kids {
					println("queue", subID)
					subID := subID
					wg.Add(1)
					go func() { queue <- Entry{subID, level + 1} }()
				}
			}()
		}
	}()

	println("fetch", opID)
	op, err := FetchItem(opID)
	if err != nil {
		return nil, err
	}

	for _, subID := range op.Kids {
		subID := subID
		println("queue", subID)
		wg.Add(1)
		go func() { queue <- Entry{subID, 1} }()
	}

	wg.Wait()
	close(queue)

	if fetchError != nil {
		return nil, err
	}

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Time < result[j].Time
	})

	result = slices.Insert(result, 0, op)

	return result, nil
}

func readCachedTopStories() ([]ItemID, error) {
	var ids []ItemID
	filename := path.Join(CacheDir, "topstories.json")

	f, err := os.Open(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	if err = json.NewDecoder(f).Decode(&ids); err != nil {
		return nil, err
	}

	return ids, nil
}

func writeCachedTopStories(ids []ItemID) error {
	filename := path.Join(CacheDir, "topstories.json")
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	return json.NewEncoder(f).Encode(ids)
}

func FetchTopStories(pageSize, pageNum int) ([]*Item, error, bool) {
	ids, err := readCachedTopStories()
	if err != nil {
		log.Printf("failed to read topstories cache: %v", err)
	}

	if ids == nil {
		resp, err := http.Get(baseURL + "/topstories.json")
		if err != nil {
			return nil, err, false
		}
		defer resp.Body.Close()

		if err = json.NewDecoder(resp.Body).Decode(&ids); err != nil {
			return nil, err, false
		}

		if err = writeCachedTopStories(ids); err != nil {
			log.Printf("failed to write topstories cache: %v", err)
		}
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
