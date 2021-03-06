package repo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"math/rand"
)

var (
	seedLlIdx int64 = 0
	seedRand        = rand.New(rand.NewSource(1597422555541))
)

func (r *Repo) Seed() error {
	ok, err := r.shouldSeed()
	if err != nil {
		return err
	}
	if !ok {
		log.Print("Not seeding non-empty database")
		return nil
	}
	log.Print("Seeding database...")
	for i := 0; i < 60; i++ {
		s, err := r.seedSuite()
		if err != nil {
			return err
		}
		for i := 0; i < genIdx(3); i++ {
			if _, err := r.seedAttachment(s.Id, nil); err != nil {
				return err
			}
		}
		for i := 0; i < int(*s.PlannedCases); i++ {
			c, err := r.seedCase(s.Id)
			if err != nil {
				return err
			}
			for i := 0; i < genIdx(3); i++ {
				if _, err := r.seedAttachment(nil, c.Id); err != nil {
					return err
				}
			}
			for i := 0; i < genIdx(60); i++ {
				if _, err := r.seedLogLine(c.Id); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (r *Repo) shouldSeed() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	colls, err := r.db.ListCollectionNames(ctx,
		bson.D{{"name", bson.D{
			{"$in", bson.A{
				"attachments",
				"suites",
				"cases",
				"logs",
			}},
		}}})
	if err != nil {
		return false, err
	}
	for _, coll := range colls {
		n, err := r.db.Collection(coll).EstimatedDocumentCount(ctx)
		if err != nil {
			return false, err
		}
		if n > 0 {
			return false, nil
		}
	}
	return true, nil
}

func (r *Repo) seedAttachment(suiteId, caseId *Id) (*Attachment, error) {
	a := genAttachment(suiteId, caseId)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	id, err := r.InsertAttachment(ctx, a)
	if err != nil {
		return nil, err
	}
	a.Id = &id
	return &a, nil
}

func genAttachment(suiteId, caseId *Id) Attachment {
	filenames := []string{
		"test.txt",
		"hello_world.png",
		"index.html",
	}
	contentTypes := []string{
		"text/plain; charset=utf-8",
		"image/png",
		"text/html",
	}
	size := int64(genIdx(1 << 16))
	timestamp, _, _ := genTimestamps()
	var a Attachment
	a.SuiteId = suiteId
	a.CaseId = caseId
	a.Filename = &filenames[genIdx(len(filenames))]
	a.ContentType = &contentTypes[genIdx(len(contentTypes))]
	a.Size = &size
	a.Timestamp = &timestamp
	return a
}

func (r *Repo) seedSuite() (*Suite, error) {
	s := genSuite()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	id, err := r.InsertSuite(ctx, s)
	if err != nil {
		return nil, err
	}
	s.Id = &id
	return &s, nil
}

func genSuite() Suite {
	startedAt, disconnectedAt, finishedAt := genTimestamps()
	projects := []string{
		"Massa Tincidunt Dui",
		"Auctor",
		"Sit Amet Luctus",
		"In Cursus",
		"Sit Amet Tellus Cras",
	}
	tags := [][]string{
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
	statuses := []SuiteStatus{
		SuiteStatusStarted,
		SuiteStatusFinished,
		SuiteStatusDisconnected,
	}
	results := []SuiteResult{
		SuiteResultPassed,
		SuiteResultFailed,
	}
	var s Suite
	s.Version = Int64(0)
	s.Project = &projects[genIdx(len(projects))]
	s.Tags = tags[genIdx(len(tags))]
	s.PlannedCases = Int64(int64(genIdx(20)))
	s.Status = &statuses[genIdx(len(statuses))]
	switch *s.Status {
	case SuiteStatusFinished:
		s.Result = &results[genIdx(len(results))]
		s.FinishedAt = &finishedAt
	case SuiteStatusDisconnected:
		s.DisconnectedAt = &disconnectedAt
	}
	s.StartedAt = &startedAt
	return s
}

func (r *Repo) seedCase(suiteId *Id) (*Case, error) {
	c := genCase(suiteId)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	id, err := r.InsertCase(ctx, c)
	if err != nil {
		return nil, err
	}
	c.Id = &id
	return &c, nil
}

func genCase(suiteId *Id) Case {
	createdAt, startedAt, finishedAt := genTimestamps()
	names := []string{
		"Lorem ipsum dolor sit",
		"Aliquam ut porttitor leo",
		"Ullamcorper dignissim cras tincidunt",
		"Morbi tincidunt ornare",
	}
	descriptions := []string{
		"Elementum tempus egestas sed sed.",
		"Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do " +
			"eiusmod tempor incididunt ut labore et dolore magna aliqua. " +
			"Vitae et leo duis ut diam quam nulla porttitor massa.",
		"",
		"Cursus in hac habitasse platea dictumst quisque.",
	}
	tags := [][]string{
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
	statuses := []CaseStatus{
		CaseStatusCreated,
		CaseStatusStarted,
		CaseStatusFinished,
	}
	results := []CaseResult{
		CaseResultPassed,
		CaseResultFailed,
		CaseResultSkipped,
		CaseResultAborted,
		CaseResultErrored,
	}
	var c Case
	c.Version = Int64(0)
	c.SuiteId = suiteId
	c.Name = &names[genIdx(len(names))]
	c.Description = &descriptions[genIdx(len(descriptions))]
	c.Tags = tags[genIdx(len(tags))]
	c.Idx = Int64(int64(genIdx(30)))
	c.Status = &statuses[genIdx(len(statuses))]
	switch *c.Status {
	case CaseStatusStarted:
		c.StartedAt = &startedAt
	case CaseStatusFinished:
		c.Result = &results[genIdx(len(results))]
		c.FinishedAt = &finishedAt
	}
	c.CreatedAt = &createdAt
	return c
}

func (r *Repo) seedLogLine(caseId *Id) (*LogLine, error) {
	ll := genLogLine(caseId)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	id, err := r.InsertLogLine(ctx, ll)
	if err != nil {
		return nil, err
	}
	ll.Id = &id
	return &ll, nil
}

func genLogLine(caseId *Id) LogLine {
	lines := []string{
		"Morbi blandit cursus risus at.",
		"Elit duis tristique sollicitudin nibh sit.\nRhoncus mattis rhoncus " +
			"urna neque viverra. Diam ut venenatis tellus in.",
		"scelerisque eleifend donec",
		"",
	}
	var ll LogLine
	ll.CaseId = caseId
	ll.Idx = &seedLlIdx
	seedLlIdx++
	ll.Error = genBool(0.01)
	ll.Line = &lines[genIdx(len(lines))]
	return ll
}

func genIdx(max int) int {
	return seedRand.Intn(max)
}

func genBool(chance float32) *bool {
	return Bool(seedRand.Float32() < chance)
}

func genTimestamps() (MsTime, MsTime, MsTime) {
	first := seedRand.Int63n(1504110555541) + 93312000000
	second := first + seedRand.Int63n(180000) + 500
	third := second + seedRand.Int63n(180000) + 500
	return NewMsTime(first), NewMsTime(second), NewMsTime(third)
}
