package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
)

var isDevMode = true

//go:embed lua/*
//go:embed pages/*
//go:embed includes/*
var embeddedFiles embed.FS

func main() {
	var fsys fs.ReadFileFS
	var bindAddr string
	var config = Config{DevMode: isDevMode}
	dirWatcher := NewDirwatcher()
	cacheManager := NewCacheManager()
	bindAddr = ":8080"

	if !isDevMode {
		fsys = embeddedFiles

		binpath, err := os.Executable()
		if err != nil {
			panic(err)
		}
		os.Chdir(path.Dir(binpath))
	} else {
		d := os.DirFS(".")

		dirWatcher.Start()

		if rfs, ok := d.(fs.ReadFileFS); !ok {
			panic("cannot cast os.DirFS to fs.ReadFileFS")
		} else {
			fsys = rfs
		}

	}

	cacheManager.Init(fsys)
	cacheManager.Start()

	cwd, _ := os.Getwd()
	log.Printf("pwd: %v\n", cwd)
	log.Printf("config: %+v\n", config)

	srv := NewServer(
		config,
		fsys,
		dirWatcher,
		cacheManager,
	)
	httpServer := &http.Server{
		Addr:    bindAddr,
		Handler: srv,
	}

	log.Printf("listening on %s\n", httpServer.Addr)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
	}
}
