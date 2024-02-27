package main

import (
	"io/fs"
	"net/http"
)

func NewServer(
        config Config,
	fsys fs.FS,
	dirWatcher *DirWatcher,
) http.Handler {
	mux := http.NewServeMux()
	addRoutes(
                config,
		mux,
		fsys,
		dirWatcher,
	)

	var handler http.Handler = mux
	return handler
}
