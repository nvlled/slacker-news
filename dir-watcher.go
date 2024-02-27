package main

import (
	"errors"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

type DirWatcher struct {
	listeners map[*func()]struct{}
}

func NewDirwatcher() *DirWatcher {
	return &DirWatcher{
		listeners: map[*func()]struct{}{},
	}
}

func (dw *DirWatcher) Start() {
	// TODO: check dev flag instead
	if _, err := os.Stat("pages"); errors.Is(err, os.ErrNotExist) {
		println("dir watcher not started")
		return
	}

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

				if path.Ext(event.Name) == ".lua" && (event.Has(fsnotify.Write) || event.Has(fsnotify.Create)) {
					print("changed", event.Name)
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
		if info.IsDir() {
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
	for fn := range dw.listeners {
		(*fn)()
	}
}

func (dw *DirWatcher) AddLuaListener(fn *func()) {
	dw.listeners[fn] = struct{}{}
}
func (dw *DirWatcher) RemoveListener(fn *func()) {
	delete(dw.listeners, fn)
}
