package handlers

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
	"time"
)

type event interface {}

type eventBus struct {
	sync.RWMutex
	subscribers []chan event
}

func (b *eventBus) subscribe() <-chan event {
	ch := make(chan event)
	b.Lock()
	b.subscribers = append(b.subscribers, ch)
	b.Unlock()
	return ch
}

func (b *eventBus) unsubscribe(ch <-chan event) bool {
	b.Lock()
	defer b.Unlock()
	for i, s := range b.subscribers {
		if ch == s {
			b.subscribers = append(b.subscribers[:i], b.subscribers[i+1:]...)
			return true
		}
	}
	return false
}

func (b *eventBus) publish(e event) {
	b.RLock()
	for _, s := range b.subscribers {
		go func(ch chan<- event) {
			ch <- e
		}(s)
	}
	b.RUnlock()
}

func (s *srv) eventsHandler(res http.ResponseWriter, req *http.Request) {
	conn, err := s.wsUpgrader.Upgrade(res, req, nil)
	if err != nil {
		log.Printf("upgrade to ws: %v\n", err)
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("close ws conn: %v\n", err)
		}
		log.Println("Disconnected WS", conn.RemoteAddr())
	}()
	log.Println("Connected WS", conn.RemoteAddr())

	eventCh := s.eventBus.subscribe()
	defer s.eventBus.unsubscribe(eventCh)
	doneCh := make(chan websocket.CloseError)
	errCh := make(chan error)

	conn.SetCloseHandler(func(code int, text string) error {
		doneCh <- websocket.CloseError{Code: code, Text: text}
		return nil
	})

	// Ignore messages from peer.
	go func() {
		for {
			if _, _, err := conn.NextReader(); err != nil {
				errCh <- fmt.Errorf("read from ws: %v", err)
				break
			}
		}
	}()

	for {
		select {
		case e := <-eventCh:
			if err := conn.WriteJSON(&e); err != nil {
				go func() {
					errCh <- fmt.Errorf("write json to ws: %v", err)
				}()
			}
		case err := <-doneCh:
			writeErr := conn.WriteControl(websocket.CloseMessage,
				websocket.FormatCloseMessage(err.Code, err.Text),
				time.Now().Add(timeout))
			if writeErr != nil {
				log.Printf("write control to ws: %v\n", writeErr)
			}
			return
		case err := <-errCh:
			log.Println(err)
			return
		}
	}
}
