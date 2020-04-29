package handlers

import (
	"log"
	"net/http"
)

func (s *srv) eventsHandler(res http.ResponseWriter, req *http.Request) {
	conn, err := s.wsUpgrader.Upgrade(res, req, nil)
	if err != nil {
		log.Printf("upgrade to ws: %v\n", err)
		return
	}

	conn.WriteJSON("hello, world!")
	conn.Close()
}