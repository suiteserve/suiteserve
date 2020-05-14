package main

import (
	"context"
	. "github.com/tmazeika/testpass/database"
	"log"
	"sync"
)

func newLogMessage(wg *sync.WaitGroup, caseLoc string) {
	defer wg.Done()
	_ = connGrp.Acquire(context.Background(), 1)
	header := postJson(*baseUri+caseLoc+"/logs", NewLogMessage{
		Seq: 1,
		Level: []LogLevelType{
			LogLevelTypeError,
			LogLevelTypeWarn,
			LogLevelTypeInfo,
			LogLevelTypeDebug,
			LogLevelTypeTrace,
		}[randUint(5)],
		Trace:     "",
		Message:   "Some nifty description would go here.",
		Timestamp: nowTimeMillis(),
	}, nil)
	connGrp.Release(1)
	loc := header.Get("location")
	log.Printf("Created log message: %s\n", loc)
}
