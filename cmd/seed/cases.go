package main

import (
	"context"
	. "github.com/tmazeika/testpass/database"
	"log"
	"strconv"
	"sync"
)

func newCase(wg *sync.WaitGroup, suiteLoc string, num int, logMsgs int) {
	defer wg.Done()
	_ = connGrp.Acquire(context.Background(), 1)
	header := postJson(*baseUri+suiteLoc+"/cases", NewCase{
		Name:        "Case " + strconv.Itoa(num),
		Num:         uint(num),
		Description: "Some test case I am!",
		Links: []CaseLink{
			{
				Type: CaseLinkTypeIssue,
				Name: "PROJ-1539",
				Url:  "https://example.com/issues/PROJ-1539",
			},
		},
		Tags: []string{
			"smoke",
			"footer",
		},
		Args: []CaseArg{
			{Key: "myarg", Value: "hello"},
		},
		StartedAt: nowTimeMillis(),
	}, nil)
	loc := header.Get("location")

	status := []CaseStatus{
		CaseStatusCreated,
		CaseStatusDisabled,
		CaseStatusPassed,
		CaseStatusFlaky,
		CaseStatusFailed,
		CaseStatusErrored,
	}[randUint(6)]
	if status != CaseStatusCreated {
		patchJson(*baseUri+loc, UpdateCase{
			Status:     status,
			FinishedAt: nowTimeMillis() + int64(randUint(100*1000)),
		})
	}
	connGrp.Release(1)
	log.Printf("Created case: %s\n", loc)

	var childWg sync.WaitGroup
	childWg.Add(logMsgs)
	for i := 0; i < logMsgs; i++ {
		go newLogMessage(&childWg, loc)
	}
	childWg.Wait()
}
