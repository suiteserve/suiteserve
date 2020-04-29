package database

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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
	Type CaseLinkType `json:"type" validate:"oneof=issue other"`
	Name string       `json:"name" validate:"required"`
	Url  string       `json:"url" validate:"url"`
}

type CaseArg struct {
	Key   string      `json:"key" validate:"required"`
	Value interface{} `json:"value,omitempty" bson:",omitempty"`
}

type NewCaseRun struct {
	Name        string     `json:"name" validate:"required"`
	Num         uint       `json:"num"`
	Description string     `json:"description,omitempty" bson:",omitempty"`
	Attachments []string   `json:"attachments,omitempty" bson:",omitempty" validate:"dive,required"`
	Links       []CaseLink `json:"links,omitempty" bson:",omitempty" validate:"dive"`
	Tags        []string   `json:"tags,omitempty" bson:",omitempty" validate:"unique,dive,required"`
	Args        []CaseArg  `json:"args,omitempty" bson:",omitempty" validate:"dive"`
	StartedAt   int64      `json:"started_at,omitempty" bson:"started_at,omitempty" validate:"gte=0"`
}

func (c *NewCaseRun) StartedAtTime() time.Time {
	return iToTime(c.StartedAt)
}

type UpdateCaseRun struct {
	Status     string `json:"status" validate:"oneof=disabled created running passed failed errored"`
	FinishedAt int64  `json:"finished_at,omitempty" bson:"finished_at,omitempty" validate:"gte=0"`
}

func (c *UpdateCaseRun) FinishedAtTime() time.Time {
	return iToTime(c.FinishedAt)
}

type CaseRun struct {
	Id            interface{} `json:"id" bson:"_id,omitempty"`
	Suite         string      `json:"suite"`
	NewCaseRun    `bson:",inline"`
	UpdateCaseRun `bson:",inline"`
}

func (d *WithContext) NewCaseRun(suiteId string, c NewCaseRun) (string, error) {
	if err := validate.Struct(&c); err != nil {
		log.Printf("validate case run: %v\n", err)
		return "", ErrInvalidModel
	}

	return d.insert(d.cases, CaseRun{
		Suite:      suiteId,
		NewCaseRun: c,
		UpdateCaseRun: UpdateCaseRun{
			Status:     CaseStatusCreated,
			FinishedAt: 0,
		},
	})
}

func (d *WithContext) UpdateCaseRun(caseId string, c UpdateCaseRun) error {
	caseOid, err := primitive.ObjectIDFromHex(caseId)
	if err != nil {
		return fmt.Errorf("%w: parse object id", ErrNotFound)
	}

	if err := validate.Struct(&c); err != nil {
		log.Printf("validate case run: %v\n", err)
		return ErrInvalidModel
	}

	ctx, cancel := d.newContext()
	defer cancel()
	res, err := d.cases.UpdateOne(ctx, bson.M{
		"_id":   caseOid,
	}, bson.M{
		"$set": &c,
	})
	if err != nil {
		return fmt.Errorf("update case run: %v", err)
	} else if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (d *WithContext) CaseRun(caseId string) (*CaseRun, error) {
	caseOid, err := primitive.ObjectIDFromHex(caseId)
	if err != nil {
		return nil, fmt.Errorf("%w: parse object id", ErrNotFound)
	}

	ctx, cancel := d.newContext()
	defer cancel()
	res := d.cases.FindOne(ctx, bson.M{
		"_id":   caseOid,
	})
	var caseRun CaseRun
	if err := res.Decode(&caseRun); err == mongo.ErrNoDocuments {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("find case run: %v", err)
	}
	return &caseRun, nil
}

func (d *WithContext) AllCaseRuns(suiteId string, caseNum *uint) ([]CaseRun, error) {
	ctx, cancel := d.newContext()
	defer cancel()
	filter := bson.M{
		"suite": suiteId,
	}
	if caseNum != nil {
		filter["num"] = *caseNum
	}
	cursor, err := d.cases.Find(ctx, filter, options.Find().SetSort(bson.D{
		{"started_at", 1},
		{"num", 1},
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

func (d *WithContext) DeleteAllCaseRuns(suiteId string) error {
	ctx, cancel := d.newContext()
	defer cancel()
	if _, err := d.cases.DeleteMany(ctx, bson.M{"suite": suiteId}); err != nil {
		return fmt.Errorf("delete all case runs for suite run: %v", err)
	}
	return nil
}
