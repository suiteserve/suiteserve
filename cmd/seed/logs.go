package main

import (
	"context"
	"github.com/tmazeika/testpass/database"
	"log"
	"sync"
)

func newLogMessage(wg *sync.WaitGroup, caseLoc string) {
	defer wg.Done()
	_ = connGrp.Acquire(context.Background(), 1)
	header := postJson(*baseUri+caseLoc+"/logs", database.NewLogMessage{
		Level:     database.LogLevelTypeInfo,
		Trace:     "",
		Message:   "Some nifty description would go here.",
		Timestamp: nowTimeMillis(),
	}, nil)
	connGrp.Release(1)
	loc := header.Get("location")
	log.Printf("Created log message: %s\n", loc)
}
