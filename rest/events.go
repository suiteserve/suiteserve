package rest

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func (s *srv) consumeRepoChanges() {
	for change := range s.repos.Changes() {
		b, err := json.Marshal(change)
		if err != nil {
			log.Printf("marshal json: %v\n", err)
		}
		s.events.Publish(string(b))
	}
}

func (s *srv) eventsHandler(w http.ResponseWriter, r *http.Request) {
	flusher := w.(http.Flusher)

	w.Header().Set("access-control-allow-origin", "*")
	w.Header().Set("cache-control", "no-cache")
	w.Header().Set("connection", "keep-alive")
	w.Header().Set("content-type", "text/event-stream")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	sub := s.events.Subscribe()
	defer sub.Unsubscribe()
	for {
		select {
		case e := <-sub.Ch():
			data := strings.ReplaceAll(e.(string), "\n", "\ndata:")
			if _, err := fmt.Fprintf(w, "data:%s\n\n", data); err != nil {
				log.Printf("write data: %v\n", err)
				return
			}
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
