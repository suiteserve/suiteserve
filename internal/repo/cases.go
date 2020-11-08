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
	SuiteId         *Id                        `json:"suiteId" bson:"suite_id"`
	Name            *string                    `json:"name,omitempty" bson:",omitempty"`
	Description     *string                    `json:"description,omitempty" bson:",omitempty"`
	Tags            []string                   `json:"tags,omitempty" bson:",omitempty"`
	Idx             *int64                     `json:"idx"`
	Args            map[string]json.RawMessage `json:"args,omitempty" bson:",omitempty"`
	Status          *CaseStatus                `json:"status"`
	Result          *CaseResult                `json:"result,omitempty" bson:",omitempty"`
	CreatedAt       *MsTime                    `json:"createdAt" bson:"created_at"`
	StartedAt       *MsTime                    `json:"startedAt,omitempty" bson:"started_at,omitempty"`
	FinishedAt      *MsTime                    `json:"finishedAt,omitempty" bson:"finished_at,omitempty"`
}

var caseType = reflect.TypeOf(Case{})

func (r *Repo) InsertCase(ctx context.Context, c Case) (Id, error) {
	return r.insert(ctx, Cases, c)
}

func (r *Repo) Case(ctx context.Context, id Id) (Case, error) {
	var c Case
	err := r.findById(ctx, Cases, id, &c)
	return c, err
}

func (r *Repo) SuiteCases(ctx context.Context,
	suiteId Id) ([]Case, error) {
	cs := []Case{}
	return cs, readAll(ctx, &cs, func() (*mongo.Cursor, error) {
		return r.db.Collection(cases).Find(ctx, bson.D{
			{"suite_id", suiteId},
		})
	})
}

func (r *Repo) FinishCase(ctx context.Context, id Id, res CaseResult,
	at MsTime) error {
	return r.updateById(ctx, Cases, id, bson.D{
		{"status", CaseStatusFinished},
		{"result", res},
		{"finished_at", at},
	})
}
