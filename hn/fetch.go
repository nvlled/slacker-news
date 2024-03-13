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
	baseURL = "https://hacker-news.firebaseio.com/v0"
)

var (
	CacheDir        = ".hn-cache"
	writeCacheQueue chan *Item
)

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
	if err := os.MkdirAll(CacheDir, 0755); err != nil {
		panic(err)
	}

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

	item.FetchTime = time.Now().Unix()

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
					subID := subID
					wg.Add(1)
					go func() { queue <- Entry{subID, level + 1} }()
				}
			}()
		}
	}()

	op, err := FetchItem(opID)
	if err != nil {
		return nil, err
	}

	for _, subID := range op.Kids {
		subID := subID
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

func readCachedStories(filename string) ([]ItemID, error) {
	var ids []ItemID
	filename = path.Join(CacheDir, filename)

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

func writeCachedStories(filename string, ids []ItemID) error {
	filename = path.Join(CacheDir,filename)
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	return json.NewEncoder(f).Encode(ids)
}

func FetchStories(feed string, pageSize, pageNum int) ([]*Item, error, bool) {
	switch feed {
	case "top":
	case "new":
	case "best":
	case "ask":
	case "show":
	case "job":
	case "":
		feed = "top"
	default:
		return nil, errors.New("invalid feed type"), false
	}

	filename := feed + "stories.json"

        var ids []ItemID
        var err error

	//ids, err := readCachedStories(filename)
	//if err != nil {
	//	log.Printf("failed to read %v cache: %v", filename, err)
	//}

	if ids == nil {
		resp, err := http.Get(baseURL + "/" + filename)
		if err != nil {
			return nil, err, false
		}
		defer resp.Body.Close()

		if err = json.NewDecoder(resp.Body).Decode(&ids); err != nil {
			return nil, err, false
		}

		//if err = writeCachedStories(filename, ids); err != nil {
		//	log.Printf("failed to write %v cache: %v", filename, err)
		//}
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
