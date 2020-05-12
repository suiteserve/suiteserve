package main

import (
	"context"
	. "github.com/tmazeika/testpass/database"
	"log"
	"sync"
)

func newSuite(wg *sync.WaitGroup, name string, cases int) {
	defer wg.Done()
	_ = connGrp.Acquire(context.Background(), 1)
	header := postJson(*baseUri+"/v1/suites", NewSuite{
		Name: name,
		FailureTypes: []SuiteFailureType{
			{Name: "IO", Description: "Input/Output exception"},
		},
		Tags: []string{"all"},
		EnvVars: []SuiteEnvVar{
			{Key: "BROWSER", Value: "chrome"},
		},
		PlannedCases: uint(cases),
		CreatedAt:    nowTimeMillis(),
	}, nil)
	loc := header.Get("location")

	status := []SuiteStatus{
		SuiteStatusRunning,
		SuiteStatusPassed,
		SuiteStatusFailed,
	}[randUint(3)]
	if status != SuiteStatusRunning {
		patchJson(*baseUri+loc, UpdateSuite{
			Status:     status,
			FinishedAt: nowTimeMillis() + int64(randUint(100*1000)),
		})
	}
	connGrp.Release(1)
	log.Printf("Created suite: %s\n", loc)

	var childWg sync.WaitGroup
	childWg.Add(cases)
	for i := 0; i < cases; i++ {
		go newCase(&childWg, loc, i+1, randUint(30))
	}
	childWg.Wait()
}
