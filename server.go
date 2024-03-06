package main

import (
	"io/fs"
	"net/http"
)

func NewServer(
        config Config,
	fsys fs.FS,
	dirWatcher *DirWatcher,
        cacheManager *CacheManager,
) http.Handler {
	mux := http.NewServeMux()
	addRoutes(
                config,
		mux,
		fsys,
		dirWatcher,
                cacheManager,
	)

	var handler http.Handler = mux
	return handler
}
