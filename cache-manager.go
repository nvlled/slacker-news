package main

import (
	"errors"
	"io/fs"
	"log"
	"os"
	"time"

	"github.com/nvlled/goom-temple/hn"
)

const (
	cacheUpdateFilename = ".last-cache-update"
)

type CacheManager struct {
	LastUpdate int64
}

func NewCacheManager() *CacheManager {
	return &CacheManager{}
}

func (c *CacheManager) GetNextUpdate() int64 {
	t := time.Unix(c.LastUpdate, 0)
	mins := t.Minute()
	t = t.Add(time.Duration(-mins) * time.Minute)
	if mins < 30 {
		mins = 30
	} else {
		mins = 60
	}
	return t.Add(time.Duration(mins) * time.Minute).Unix()
}

func (c *CacheManager) Init(fsys fs.FS) {
	stat, err := fsStat(fsys, cacheUpdateFilename)
	if err != nil || stat == nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Printf("failed to read file: %v", err)
		} else {
			if err = os.WriteFile(cacheUpdateFilename, nil, 0644); err != nil {
				log.Printf("failed to create file: %v", err)
			}
			c.LastUpdate = time.Now().Unix()
		}
	} else {
		c.LastUpdate = stat.ModTime().Unix()
	}
}

func (c *CacheManager) startLoop() {
	for {
		now := time.Now().Unix()
		if now >= c.GetNextUpdate() {
			log.Println("clearing cache")
			if err := os.RemoveAll(hn.CacheDir); err != nil {
				log.Printf("failed to clear cache: %v", err)
			} else {
				if err := os.MkdirAll(hn.CacheDir, 0755); err != nil {
					panic(err)
				}
				if err = os.WriteFile(cacheUpdateFilename, nil, 0644); err != nil {
					log.Printf("failed to create file: %v", err)
				}
				log.Println("cache cleared")
				c.LastUpdate = now
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func (c *CacheManager) Start() {
	go c.startLoop()
}
