package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"math/rand"
)

var seedRand = rand.New(rand.NewSource(1597422555541))

func (r *Repo) Seed() error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	colls, err := r.db.ListCollectionNames(ctx,
		bson.D{{"name", bson.D{{"$ne", "schema_migrations"}}}})
	if err != nil {
		return err
	}
	for _, coll := range colls {
		n, err := r.db.Collection(coll).EstimatedDocumentCount(ctx)
		if err != nil {
			return err
		}
		if n > 0 {
			log.Print("Not seeding non-empty database")
			return nil
		}
	}
	log.Print("Seeding database...")
	for i := 0; i < 60; i++ {
		s := genSuite()
		id, err := r.InsertSuite(s)
		if err != nil {
			return err
		}
		fmt.Printf("> insert suite %q\n", id)
		for j := 0; j < int(s.PlannedCases); j++ {
			id, err := r.InsertCase(genCase(id))
			if err != nil {
				return err
			}
			fmt.Printf("  > insert case %q\n", id)
			for k := 0; k < genIdx(60); k++ {
				id, err := r.InsertLogLine(genLogLine(id))
				if err != nil {
					return err
				}
				fmt.Printf("    > insert log line %q\n", id)
			}
		}
	}
	return nil
}

func genSuite() Suite {
	startedAt, disconnectedAt, finishedAt, deletedAt := genTimestamps()
	nameArr := []string{
		"Massa Tincidunt Dui",
		"Auctor",
		"Sit Amet Luctus",
		"In Cursus",
		"Sit Amet Tellus Cras",
	}
	tagsArr := [][]string{
		{"erat", "sed"},
		{"id"},
		{"risus"},
		{"porta nibh"},
		{"erat", "velit", "scelerisque"},
		{"pellentesque", "sit"},
		nil,
		nil,
		nil,
	}
	statusArr := []SuiteStatus{
		SuiteStatusStarted,
		SuiteStatusFinished,
		SuiteStatusDisconnected,
	}
	resultArr := []SuiteResult{
		SuiteResultPassed,
		SuiteResultFailed,
	}
	var s Suite
	s.Name = nameArr[genIdx(len(nameArr))]
	s.Tags = tagsArr[genIdx(len(tagsArr))]
	s.PlannedCases = int64(genIdx(20))
	s.Status = statusArr[genIdx(len(statusArr))]
	switch s.Status {
	case SuiteStatusFinished:
		s.Result = resultArr[genIdx(len(resultArr))]
		s.FinishedAt = finishedAt
	case SuiteStatusDisconnected:
		s.DisconnectedAt = disconnectedAt
	}
	s.StartedAt = startedAt
	if genBool(0.1) {
		s.Deleted = true
		s.DeletedAt = deletedAt
	}
	return s
}

func genCase(suiteId string) Case {
	createdAt, startedAt, finishedAt, _ := genTimestamps()
	nameArr := []string{
		"Lorem ipsum dolor sit",
		"Aliquam ut porttitor leo",
		"Ullamcorper dignissim cras tincidunt",
		"Morbi tincidunt ornare",
	}
	descriptionArr := []string{
		"Elementum tempus egestas sed sed.",
		"Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do " +
			"eiusmod tempor incididunt ut labore et dolore magna aliqua. " +
			"Vitae et leo duis ut diam quam nulla porttitor massa.",
		"",
		"Cursus in hac habitasse platea dictumst quisque.",
	}
	tagsArr := [][]string{
		{"erat", "sed"},
		{"id"},
		{"risus"},
		{"porta nibh"},
		{"erat", "velit", "scelerisque"},
		{"pellentesque", "sit"},
		nil,
		nil,
		nil,
	}
	argsArr := []map[string]json.RawMessage{
		nil,
		{
			"x": json.RawMessage("10"),
			"y": json.RawMessage("6"),
		},
		{
			"name": json.RawMessage(`"hello"`),
			"obj":  json.RawMessage(`{"first":true,"second":null}`),
		},
		{
			"arg": json.RawMessage(`null`),
		},
	}
	statusArr := []CaseStatus{
		CaseStatusCreated,
		CaseStatusStarted,
		CaseStatusFinished,
	}
	resultArr := []CaseResult{
		CaseResultPassed,
		CaseResultFailed,
		CaseResultSkipped,
		CaseResultAborted,
		CaseResultErrored,
	}
	var c Case
	c.SuiteId = suiteId
	c.Name = nameArr[genIdx(len(nameArr))]
	c.Description = descriptionArr[genIdx(len(descriptionArr))]
	c.Tags = tagsArr[genIdx(len(tagsArr))]
	c.Idx = int64(genIdx(30))
	c.Args = argsArr[genIdx(len(argsArr))]
	c.Status = statusArr[genIdx(len(statusArr))]
	switch c.Status {
	case CaseStatusStarted:
		c.StartedAt = startedAt
	case CaseStatusFinished:
		c.Result = resultArr[genIdx(len(resultArr))]
		c.FinishedAt = finishedAt
	}
	c.CreatedAt = createdAt
	return c
}

var idxInc int64 = 0

func genLogLine(caseId string) LogLine {
	levelArr := []LogLevelType{
		LogLevelTypeTrace,
		LogLevelTypeDebug,
		LogLevelTypeInfo,
		LogLevelTypeWarn,
		LogLevelTypeError,
	}
	traceArr := []string{
		`panic: Hello, world!

goroutine 1 [running]:
main.f3(...)
        /Users/username/scratch_1.go:16
main.f2(...)
        /Users/username/scratch_1.go:12
main.f1(...)
        /Users/username/scratch_1.go:8
main.main()
        /Users/username/scratch_1.go:4 +0x3b`,
		`Exception in thread "main" java.lang.IllegalArgumentException: Hello, world!
	at Main.f3(Main.java:16)
	at Main.f2(Main.java:12)
	at Main.f1(Main.java:8)
	at Main.main(Main.java:4)`,
		`Traceback (most recent call last):
  File "<stdin>", line 1, in <module>
  File "<stdin>", line 2, in f1
  File "<stdin>", line 2, in f2
  File "<stdin>", line 2, in f3
ZeroDivisionError: division by zero`,
	}
	messageArr := []string{
		"Morbi blandit cursus risus at.",
		"Elit duis tristique sollicitudin nibh sit.\nRhoncus mattis rhoncus " +
			"urna neque viverra. Diam ut venenatis tellus in.",
		"scelerisque eleifend donec",
		"",
	}
	var ll LogLine
	ll.CaseId = caseId
	ll.Idx = idxInc
	idxInc++
	ll.Level = levelArr[genIdx(len(levelArr))]
	if ll.Level == LogLevelTypeError {
		ll.Trace = traceArr[genIdx(len(traceArr))]
	}
	ll.Message = messageArr[genIdx(len(messageArr))]
	ll.Timestamp, _, _, _ = genTimestamps()
	return ll
}

func genIdx(max int) int {
	return seedRand.Intn(max)
}

func genBool(chance float32) bool {
	return seedRand.Float32() < chance
}

func genTimestamps() (int64, int64, int64, int64) {
	first := seedRand.Int63n(1504110555541) + 93312000000
	second := first + seedRand.Int63n(180000) + 500
	third := second + seedRand.Int63n(180000) + 500
	fourth := third + seedRand.Int63n(180000) + 500
	return first, second, third, fourth
}
