package sse

import "net/http"

func NewMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("cache-control", "no-cache, no-transform")
		w.Header().Set("connection", "keep-alive")
		w.Header().Set("content-type", "text/event-stream")
		h.ServeHTTP(w, r)
	})
}
