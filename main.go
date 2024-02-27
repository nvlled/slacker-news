package main

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
)

//go:embed lua/*
//go:embed pages/*
var embeddedFiles embed.FS

func main() {
	var fsys fs.FS
	if _, err := os.Stat("lua"); errors.Is(err, os.ErrNotExist) {
		fsys = embeddedFiles
	} else {
		fsys = os.DirFS(".")
	}
	dirWatcher := NewDirwatcher()
	dirWatcher.Start()

	var config = Config{
		DevMode: os.Getenv("PROD") == "",
	}
        log.Printf("config: %+v\n", config)

	srv := NewServer(
		config,
		fsys,
		dirWatcher,
	)
	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: srv,
	}
	log.Printf("listening on %s\n", httpServer.Addr)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
	}
}
