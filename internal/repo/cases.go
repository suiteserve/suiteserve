package repo

import (
	"context"
	"encoding/json"
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
	SuiteId         Id                         `json:"suite_id" bson:"suite_id"`
	Name            string                     `json:"name,omitempty" bson:",omitempty"`
	Description     string                     `json:"description,omitempty" bson:",omitempty"`
	Tags            []string                   `json:"tags,omitempty" bson:",omitempty"`
	Idx             int64                      `json:"idx"`
	Args            map[string]json.RawMessage `json:"args,omitempty" bson:",omitempty"`
	Status          CaseStatus                 `json:"status"`
	Result          CaseResult                 `json:"result,omitempty" bson:",omitempty"`
	CreatedAt       int64                      `json:"created_at" bson:"created_at"`
	StartedAt       int64                      `json:"started_at,omitempty" bson:"started_at,omitempty"`
	FinishedAt      int64                      `json:"finished_at,omitempty" bson:"finished_at,omitempty"`
}

func (r *Repo) InsertCase(ctx context.Context, c Case) (Id, error) {
	return r.insert(ctx, "cases", c)
}

func (r *Repo) Case(ctx context.Context, id Id) (interface{}, error) {
	var c Case
	if err := r.findById(ctx, "cases", id, &c); err != nil {
		return nil, err
	}
	return c, nil
}
