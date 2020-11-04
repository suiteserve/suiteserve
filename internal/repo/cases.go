package repo

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

type CaseStatus string

const (
	CaseStatusCreated  CaseStatus = "created"
	CaseStatusStarted  CaseStatus = "started"
	CaseStatusFinished CaseStatus = "finished"
)

type CaseResult string

const (
	CaseResultPassed  CaseResult = "passed"
	CaseResultFailed  CaseResult = "failed"
	CaseResultSkipped CaseResult = "skipped"
	CaseResultAborted CaseResult = "aborted"
	CaseResultErrored CaseResult = "errored"
)

type Case struct {
	Entity          `bson:",inline"`
	VersionedEntity `bson:",inline"`
	SuiteId         Id                         `json:"suiteId" bson:"suite_id"`
	Name            *string                    `json:"name,omitempty" bson:",omitempty"`
	Description     *string                    `json:"description,omitempty" bson:",omitempty"`
	Tags            []string                   `json:"tags,omitempty" bson:",omitempty"`
	Idx             *int64                     `json:"idx"`
	Args            map[string]json.RawMessage `json:"args,omitempty" bson:",omitempty"`
	Status          *CaseStatus                `json:"status"`
	Result          *CaseResult                `json:"result,omitempty" bson:",omitempty"`
	CreatedAt       *Time                      `json:"createdAt" bson:"created_at"`
	StartedAt       *Time                      `json:"startedAt,omitempty" bson:"started_at,omitempty"`
	FinishedAt      *Time                      `json:"finishedAt,omitempty" bson:"finished_at,omitempty"`
}

var caseType = reflect.TypeOf(Case{})

func (r *Repo) InsertCase(ctx context.Context, c Case) (Id, error) {
	return r.insert(ctx, casesColl, c)
}

func (r *Repo) Case(ctx context.Context, id Id) (interface{}, error) {
	return r.findById(ctx, casesColl, id, Case{})
}

func (r *Repo) SuiteCases(ctx context.Context,
	suiteId Id) (interface{}, error) {
	return readAll(ctx, []Case{}, func() (*mongo.Cursor, error) {
		return r.db.Collection(casesColl).Find(ctx, bson.D{
			{"suite_id", bsonId{suiteId}},
		})
	})
}

func (r *Repo) FinishCase(ctx context.Context, id Id, res CaseResult,
	at Time) error {
	return r.updateById(ctx, casesColl, id, bson.D{
		{"status", CaseStatusFinished},
		{"result", res},
		{"finished_at", at},
	})
}
