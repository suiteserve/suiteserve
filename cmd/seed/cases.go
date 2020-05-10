package main

import (
	"context"
	"github.com/tmazeika/testpass/database"
	"log"
	"strconv"
)

func newCase(suiteLoc string, num int, logMsgs int) {
	_ = connGrp.Acquire(context.Background(), 1)
	defer connGrp.Release(1)
	defer waitGrp.Done()
	header := postJson(*baseUri+suiteLoc+"/cases", database.NewCase{
		Name:        "Case " + strconv.Itoa(num),
		Num:         uint(num),
		Description: "Some test case I am!",
		Links: []database.CaseLink{
			{
				Type: database.CaseLinkTypeIssue,
				Name: "PROJ-1539",
				Url:  "https://example.com/issues/PROJ-1539",
			},
		},
		Tags: []string{
			"smoke",
			"footer",
		},
		Args: []database.CaseArg{
			{Key: "myarg", Value: "hello"},
		},
		StartedAt: nowTimeMillis(),
	}, nil)
	loc := header.Get("location")
	log.Printf("Created case: %s\n", loc)

	waitGrp.Add(logMsgs)
	for i := 0; i < logMsgs; i++ {
		go newLogMessage(loc)
	}
}
