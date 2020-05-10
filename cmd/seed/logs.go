package main

import (
	"context"
	"github.com/tmazeika/testpass/database"
	"log"
)

func newLogMessage(caseLoc string) {
	_ = connGrp.Acquire(context.Background(), 1)
	defer connGrp.Release(1)
	defer waitGrp.Done()
	header := postJson(*baseUri+caseLoc+"/logs", database.NewLogMessage{
		Level:     database.LogLevelTypeInfo,
		Trace:     "",
		Message:   "Some nifty description would go here.",
		Timestamp: nowTimeMillis(),
	}, nil)
	loc := header.Get("location")
	log.Printf("Created log message: %s\n", loc)
}
