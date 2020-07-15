package rpc

import (
	"crypto/rand"
	"encoding/hex"
	"log"
)

type sessionId string

func genSessionId() sessionId {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		log.Panicf("read rand: %v", err)
	}
	return sessionId(hex.EncodeToString(b))
}

type sessionRepo map[sessionId]*session

func (r sessionRepo) newSession() (sessionId, *session) {
	var s session
	id := genSessionId()
	r[id] = &s
	return id, &s
}

type session struct {
}
