package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/suiteserve/suiteserve/event"
	"github.com/suiteserve/suiteserve/repo"
	"log"
	"net/http"
)

type ChangefeedMsg struct {
	Seq int64  `json:"seq"`
	Cmd string `json:"cmd"`
}

type idsChangefeedMsg struct {
	Payload struct {
		Ids []string `json:"ids,omitempty"`
	} `json:"payload"`
}

type eventChangefeedMsg struct {
	ChangefeedMsg
	Payload interface{} `json:"payload"`
}

func newEventChangefeedMsg(seq int64) {
}

type changefeedRepo interface {
	attachmentFinder
	suiteFinder
	caseFinder
	logFinder

	Changes() *event.Bus
}

type changefeed struct {
	upgrader websocket.Upgrader
	repo     changefeedRepo
}

func newChangefeed(r changefeedRepo) *changefeed {
	return &changefeed{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// TODO: for testing
				return true
			},
		},
		repo: r,
	}
}

type changefeedSession struct {
	*changefeed
	conn        *websocket.Conn
	writer      chan interface{}
	done        chan interface{}
	suiteFilter func(op repo.ChangeOp, id string) bool
}

func (c *changefeed) newHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := c.upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("%s upgrade ws: %v\n", r.RemoteAddr, err)
			return
		}
		defer func() {
			if err := conn.Close(); err != nil {
				log.Printf("%s close ws: %v\n", conn.RemoteAddr(), err)
			}
		}()
		s := changefeedSession{
			changefeed: c,
			conn:       conn,
			writer:     make(chan interface{}),
			done:       make(chan interface{}),
			suiteFilter: func(_ repo.ChangeOp, _ string) bool {
				return true
			},
		}
		defer close(s.done)
		go s.readWs()
		go s.writeWs()
		s.readChanges()
	})
}

func (s *changefeedSession) readWs() {
	for {
		if err := s.readNextWs(); err != nil {
			var closeErr *websocket.CloseError
			if errors.As(err, &closeErr) && websocket.IsUnexpectedCloseError(closeErr) {
				return
			}
			log.Printf("%s read ws: %v\n", s.conn.RemoteAddr(), err)
			var tempErr interface {
				Temporary() bool
			}
			if errors.As(err, &tempErr) && tempErr.Temporary() {
				continue
			}
			return
		}
	}
}

func (s *changefeedSession) readNextWs() error {
	msgType, b, err := s.conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("read msg: %w", err)
	}
	if msgType != websocket.TextMessage {
		return errors.New("not a text msg")
	}
	var msg ChangefeedMsg
	if err := json.Unmarshal(b, &msg); err != nil {
		return fmt.Errorf("unmarshal json: %v", err)
	}
	switch msg.Cmd {
	case "sub_suites":
		return s.handleSubSuitesCmd(msg.Seq, b)
	default:
		return fmt.Errorf("unknown cmd %q", msg.Cmd)
	}
}

func (s *changefeedSession) handleSubSuitesCmd(seq int64, raw json.RawMessage) error {
	var msg idsChangefeedMsg
	if err := json.Unmarshal(raw, &msg); err != nil {
		return fmt.Errorf("unmarshal json: %v", err)
	}
	s.suiteFilter = func(op repo.ChangeOp, id string) bool {
		if op == repo.ChangeOpInsert {
			return true
		}
		for _, v := range msg.Payload.Ids {
			if v == id {
				return true
			}
		}
		return false
	}
	s.sendOk(seq)
	return nil
}

func (s *changefeedSession) writeWs() {
	for {
		select {
		case e := <-s.writer:
			if err := s.conn.WriteJSON(e); err != nil {
				log.Printf("%s write json: %v\n", s.conn.RemoteAddr(), err)
			}
		case <-s.done:
			return
		}
	}
}

func (s *changefeedSession) readChanges() {
	sub := s.repo.Changes().Subscribe()
	defer sub.Unsubscribe()
	for {
		select {
		case v := <-sub.Ch():
			s.onChange(v.(repo.Change))
		case <-s.done:
			return
		}
	}
}

func (s *changefeedSession) onChange(c repo.Change) {
	var ok bool
	switch c.Collection() {
	case repo.CollSuites:
		ok = s.suiteFilter(c.Operation(), c.DocId())
	case repo.CollSuiteAggs:
		ok = true
	}
	if ok {
		s.writer <- &ChangefeedMsg{
			Cmd:     "change",
		}
	}
}

func (s *changefeedSession) sendOk(seq int64) {
	s.writer <- &ChangefeedMsg{
		Seq: seq,
		Cmd: "ok",
	}
}
