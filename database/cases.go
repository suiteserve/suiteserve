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

type NewCase struct {
	Name        string     `json:"name" validate:"required"`
	Num         uint       `json:"num"`
	Description string     `json:"description,omitempty" bson:",omitempty"`
	Attachments []string   `json:"attachments,omitempty" bson:",omitempty" validate:"dive,required"`
	Links       []CaseLink `json:"links,omitempty" bson:",omitempty" validate:"dive"`
	Tags        []string   `json:"tags,omitempty" bson:",omitempty" validate:"unique,dive,required"`
	Args        []CaseArg  `json:"args,omitempty" bson:",omitempty" validate:"dive"`
	StartedAt   int64      `json:"started_at,omitempty" bson:"started_at,omitempty" validate:"gte=0"`
}

func (c *NewCase) StartedAtTime() time.Time {
	return iToTime(c.StartedAt)
}

type UpdateCase struct {
	Status     string `json:"status" validate:"oneof=disabled created running passed failed errored"`
	FinishedAt int64  `json:"finished_at,omitempty" bson:"finished_at,omitempty" validate:"gte=0"`
}

func (c *UpdateCase) FinishedAtTime() time.Time {
	return iToTime(c.FinishedAt)
}

type Case struct {
	Id         interface{} `json:"id" bson:"_id,omitempty"`
	Suite      string      `json:"suite"`
	NewCase    `bson:",inline"`
	UpdateCase `bson:",inline"`
}

func (d *WithContext) NewCase(suiteId string, c NewCase) (string, error) {
	if err := validate.Struct(&c); err != nil {
		log.Printf("validate case: %v\n", err)
		return "", ErrInvalidModel
	}

	return d.insert(d.cases, Case{
		Suite:   suiteId,
		NewCase: c,
		UpdateCase: UpdateCase{
			Status:     CaseStatusCreated,
			FinishedAt: 0,
		},
	})
}

func (d *WithContext) UpdateCase(caseId string, c UpdateCase) error {
	caseOid, err := primitive.ObjectIDFromHex(caseId)
	if err != nil {
		return fmt.Errorf("%w: parse object id", ErrNotFound)
	}

	if err := validate.Struct(&c); err != nil {
		log.Printf("validate case: %v\n", err)
		return ErrInvalidModel
	}

	ctx, cancel := d.newContext()
	defer cancel()
	res, err := d.cases.UpdateOne(ctx, bson.M{
		"_id": caseOid,
	}, bson.M{
		"$set": &c,
	})
	if err != nil {
		return fmt.Errorf("update case: %v", err)
	} else if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (d *WithContext) Case(id string) (*Case, error) {
	caseOid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("%w: parse object id", ErrNotFound)
	}

	ctx, cancel := d.newContext()
	defer cancel()
	res := d.cases.FindOne(ctx, bson.M{"_id": caseOid})
	var _case Case
	if err := res.Decode(&_case); err == mongo.ErrNoDocuments {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("find case: %v", err)
	}
	return &_case, nil
}

func (d *WithContext) AllCases(suiteId string, caseNum *uint) ([]Case, error) {
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
		return nil, fmt.Errorf("find many cases for suite: %v", err)
	}

	cases := make([]Case, 0)
	if err := cursor.All(ctx, &cases); err != nil {
		return nil, fmt.Errorf("decode many cases for suite: %v", err)
	}
	return cases, nil
}

func (d *WithContext) DeleteAllCases(suiteId string) error {
	ctx, cancel := d.newContext()
	defer cancel()
	if _, err := d.cases.DeleteMany(ctx, bson.M{"suite": suiteId}); err != nil {
		return fmt.Errorf("delete all cases for suite: %v", err)
	}
	return nil
}
