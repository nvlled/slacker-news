package main

import (
	"fmt"
	"net/http"
)

func handleAutoReloadPage(dirWatcher *DirWatcher) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Connection", "keep-alive")
		w.WriteHeader(http.StatusOK)

		fn := func() {
			fmt.Fprintf(w, "event: fsevent\ndata: x\n\n")
			w.(http.Flusher).Flush()
		}
		dirWatcher.AddLuaListener(&fn)
		defer dirWatcher.RemoveListener(&fn)

		//running := true
		//go func() {
		//	for running {
		//		time.Sleep(5 * time.Second)
		//		fmt.Fprintf(w, "event:ping\n\n")
		//	}
		//}()

		<-r.Context().Done()
		//running = false
	})
}
