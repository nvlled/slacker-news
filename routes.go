package main

import (
	"io/fs"
	"net/http"
)

func addRoutes(
	config Config,
	mux *http.ServeMux,
	fsys fs.FS,
	dirWatcher *DirWatcher,
) {
	mux.Handle("/", handleServeLuaPage(config, fsys))

	if config.DevMode {
		mux.Handle("/.autoreload", handleAutoReloadPage(dirWatcher))
	}
}
