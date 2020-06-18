package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/tmazeika/testpass/event"
	"log"
	"net/http"
)

type changes struct {
	upgrader    websocket.Upgrader
	changes     *event.Bus
	attachments attachmentFinder
	suites      suiteFinder
	cases       caseFinder
	logs        logFinder
}

func newChanges(repo Repo) *changes {
	return &changes{
		upgrader:    websocket.Upgrader{},
		changes:     repo.Changes(),
		attachments: repo,
		suites:      repo,
		cases:       repo,
		logs:        repo,
	}
}

type changesSession struct {
	conn *websocket.Conn
}

func (c *changes) newHandler() http.Handler {
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

		sess := changesSession{conn: conn}
		done := make(chan interface{})
		go c.handleMessages(&sess, done)
		c.handleChanges(&sess, done)
	})
}

func (c *changes) handleMessages(sess *changesSession, done chan<- interface{}) {
	defer close(done)
	for {
		if err := c.handleNextMessage(sess); err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				return
			}
			log.Printf("%s read ws: %v\n", sess.conn.RemoteAddr(), err)
			var tmpErr interface {
				Temporary() bool
			}
			if errors.As(err, &tmpErr) && tmpErr.Temporary() {
				continue
			}
			return
		}
	}
}

func (c *changes) handleNextMessage(sess *changesSession) error {
	msgType, b, err := sess.conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("read msg: %w", err)
	}
	if msgType != websocket.TextMessage {
		return errors.New("not a text msg")
	}
	var msg struct {
		Cmd string `json:"cmd"`
	}
	if err := json.Unmarshal(b, &msg); err != nil {
		return fmt.Errorf("unmarshal json: %v", err)
	}

	switch msg.Cmd {
	case "sub_suites":
		return c.handleSubSuitesCmd(sess, b)
	default:
		return fmt.Errorf("unknown cmd %q", msg.Cmd)
	}
}

func (c *changes) handleSubSuitesCmd(sess *changesSession, msg json.RawMessage) error {
	var subSuitesMsg struct {
		FromId int `json:"from_id"`
		ToId int `json:"to_id"`
	}
	if err := json.Unmarshal(msg, &subSuitesMsg); err != nil {
		return fmt.Errorf("unmarshal json: %v", err)
	}
	panic("TODO")
}

func (c *changes) handleChanges(sess *changesSession, done <-chan interface{}) {
	sub := c.changes.Subscribe()
	defer sub.Unsubscribe()
	for {
		select {
		case e := <-sub.Ch():
			if err := sess.conn.WriteJSON(&e); err != nil {
				log.Printf("%s write json: %v\n", sess.conn.RemoteAddr(), err)
			}
		case <-done:
			return
		}
	}
}
