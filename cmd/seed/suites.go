package main

import (
	"context"
	"github.com/tmazeika/testpass/database"
	"log"
	"sync"
)

func newSuite(wg *sync.WaitGroup, name string, cases int) {
	defer wg.Done()
	_ = connGrp.Acquire(context.Background(), 1)
	header := postJson(*baseUri+"/v1/suites", database.NewSuite{
		Name: name,
		FailureTypes: []database.SuiteFailureType{
			{Name: "IO", Description: "Input/Output exception"},
		},
		Tags: []string{"all"},
		EnvVars: []database.SuiteEnvVar{
			{Key: "BROWSER", Value: "chrome"},
		},
		PlannedCases: uint(cases),
		CreatedAt:    nowTimeMillis(),
	}, nil)
	connGrp.Release(1)
	loc := header.Get("location")
	log.Printf("Created suite: %s\n", loc)

	var childWg sync.WaitGroup
	childWg.Add(cases)
	for i := 0; i < cases; i++ {
		go newCase(&childWg, loc, i+1, randUint(30))
	}
	childWg.Wait()
}
