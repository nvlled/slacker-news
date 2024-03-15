package main

import (
	"io/fs"
	"log"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
)

type DirWatcher struct {
	mu        sync.Mutex
	listeners map[*func()]struct{}
}

func NewDirwatcher() *DirWatcher {
	return &DirWatcher{
		listeners: map[*func()]struct{}{},
	}
}

func (dw *DirWatcher) Start() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	// Start listening for events.
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if !event.Has(fsnotify.Write) {
					continue
				}

				ext := path.Ext(event.Name)
				if ext == ".lua" /*|| ext == ".go"*/ || strings.HasPrefix(event.Name, "pages/") {
					log.Println("event:", event, path.Ext(event.Name))
					dw.notify()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}

	}()

	filepath.Walk(".", func(filename string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && filename != ".git" && !strings.HasPrefix(filename, ".git/") {
			println("watching", filename)
			err = watcher.Add(filename)
			if err != nil {
				log.Fatal(err)
			}
		}
		return err
	})
}

func (dw *DirWatcher) notify() {
	listeners := dw.listeners
	for fn := range listeners {
		(*fn)()
	}
}

func (dw *DirWatcher) AddLuaListener(fn *func()) {
	dw.mu.Lock()
	dw.listeners[fn] = struct{}{}
	dw.mu.Unlock()
}
func (dw *DirWatcher) RemoveListener(fn *func()) {
	dw.mu.Lock()
	delete(dw.listeners, fn)
	dw.mu.Unlock()
}
