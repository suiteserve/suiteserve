package database

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type (
	CaseLinkType string
	CaseStatus   string
)

const (
	CaseLinkTypeIssue CaseLinkType = "issue"
	CaseLinkTypeOther              = "other"

	CaseStatusDisabled CaseStatus = "disabled"
	CaseStatusCreated             = "created"
	CaseStatusRunning             = "running"
	CaseStatusPassed              = "passed"
	CaseStatusFailed              = "failed"
	CaseStatusErrored             = "errored"
)

type CaseLink struct {
	Type CaseLinkType `json:"type" validate:"required,oneof=issue other"`
	Name string       `json:"name" validate:"required"`
	Url  string       `json:"url" validate:"required"`
}

type CaseArg struct {
	Key   string      `json:"key" validate:"required"`
	Value interface{} `json:"value"`
}

type NewCaseRun struct {
	Name        string     `json:"name" validate:"required"`
	Num         uint       `json:"num" validate:"gte=0"`
	Description string     `json:"description,omitempty" bson:",omitempty"`
	Attachments []string   `json:"attachments,omitempty" bson:",omitempty" validate:"dive,required"`
	Links       []CaseLink `json:"links,omitempty" bson:",omitempty" validate:"dive"`
	Tags        []string   `json:"tags,omitempty" bson:",omitempty" validate:"dive,required"`
	Args        []CaseArg  `json:"args,omitempty" bson:",omitempty" validate:"dive"`
	StartedAt   int64      `json:"started_at,omitempty" bson:"started_at,omitempty" validate:"gte=0"`
}

func (c *NewCaseRun) StartedAtTime() time.Time {
	return time.Unix(c.StartedAt, 0)
}

type CaseRun struct {
	Id         interface{} `json:"id" bson:"_id,omitempty"`
	Suite      string      `json:"suite"`
	NewCaseRun `bson:",inline"`
	Status     string `json:"status"`
	FinishedAt int64  `json:"finished_at,omitempty" bson:"finished_at,omitempty"`
}

func (c *CaseRun) FinishedAtTime() time.Time {
	return time.Unix(c.FinishedAt, 0)
}

func (d *WithContext) NewCaseRun(suiteId string, c NewCaseRun) (string, error) {
	if err := validate.Struct(&c); err != nil {
		log.Printf("validate case run: %v\n", err)
		return "", ErrInvalidModel
	}

	return d.insert(d.cases, CaseRun{
		Suite:      suiteId,
		NewCaseRun: c,
		Status:     CaseStatusCreated,
	})
}

func (d *WithContext) CaseRuns(suiteId string, caseNum uint) ([]CaseRun, error) {
	ctx, cancel := d.newContext()
	defer cancel()
	cursor, err := d.cases.Find(ctx, bson.M{
		"suite": suiteId,
		"num":   caseNum,
	}, options.Find().SetSort(bson.D{
		{"started_at", 1},
		{"_id", 1},
	}))
	if err != nil {
		return nil, fmt.Errorf("find case runs: %v", err)
	}

	caseRuns := make([]CaseRun, 0)
	if err := cursor.All(ctx, &caseRuns); err != nil {
		return nil, fmt.Errorf("decode case runs: %v", err)
	}
	return caseRuns, nil
}

func (d *WithContext) AllCaseRuns(suiteId string) ([]CaseRun, error) {
	ctx, cancel := d.newContext()
	defer cancel()
	cursor, err := d.cases.Find(ctx, bson.M{
		"suite": suiteId,
	}, options.Find().SetSort(bson.D{
		{"num", 1},
		{"started_at", 1},
		{"_id", 1},
	}))
	if err != nil {
		return nil, fmt.Errorf("find all case runs for suite run: %v", err)
	}

	caseRuns := make([]CaseRun, 0)
	if err := cursor.All(ctx, &caseRuns); err != nil {
		return nil, fmt.Errorf("decode all case runs for suite run: %v", err)
	}
	return caseRuns, nil
}

func (d *WithContext) DeleteCaseRuns(suiteId string, caseNum uint) error {
	ctx, cancel := d.newContext()
	defer cancel()
	_, err := d.cases.DeleteMany(ctx, bson.M{
		"suite": suiteId,
		"num":   caseNum,
	})
	if err != nil {
		return fmt.Errorf("delete case runs: %v", err)
	}
	return nil
}

func (d *WithContext) DeleteAllCaseRuns(suiteId string) error {
	ctx, cancel := d.newContext()
	defer cancel()
	if _, err := d.cases.DeleteMany(ctx, bson.M{"suite": suiteId}); err != nil {
		return fmt.Errorf("delete all case runs for suite run: %v", err)
	}
	return nil
}
