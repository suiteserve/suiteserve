package main

import (
	"context"
	"github.com/tmazeika/testpass/database"
	"log"
)

func newSuite(name string, cases int) {
	_ = connGrp.Acquire(context.Background(), 1)
	defer connGrp.Release(1)
	defer waitGrp.Done()
	header := postJson(*baseUri+"/suites", database.NewSuite{
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
	loc := header.Get("location")
	log.Printf("Created suite: %s\n", loc)

	waitGrp.Add(cases)
	for i := 0; i < cases; i++ {
		go newCase(loc, i+1, randUint(30))
	}
}
