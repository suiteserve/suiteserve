package repo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
)

type SuiteStatus string

const (
	SuiteStatusStarted      SuiteStatus = "started"
	SuiteStatusFinished     SuiteStatus = "finished"
	SuiteStatusDisconnected SuiteStatus = "disconnected"
)

type SuiteResult string

const (
	SuiteResultPassed SuiteResult = "passed"
	SuiteResultFailed SuiteResult = "failed"
)

type Suite struct {
	Entity           `bson:",inline"`
	VersionedEntity  `bson:",inline"`
	SoftDeleteEntity `bson:",inline"`
	Name             string      `json:"name,omitempty" bson:",omitempty"`
	Tags             []string    `json:"tags,omitempty" bson:",omitempty"`
	PlannedCases     int64       `json:"planned_cases,omitempty" bson:",omitempty"`
	Status           SuiteStatus `json:"status"`
	Result           SuiteResult `json:"result"`
	DisconnectedAt   int64       `json:"disconnected_at,omitempty" bson:"disconnected_at,omitempty"`
	StartedAt        int64       `json:"started_at" bson:"started_at"`
	FinishedAt       int64       `json:"finished_at,omitempty" bson:"finished_at,omitempty"`
}

func (r *Repo) InsertSuite(ctx context.Context, s Suite) (Id, error) {
	return r.insert(ctx, "suites", s)
}

func (r *Repo) Suite(ctx context.Context, id Id) (*Suite, error) {
	var s Suite
	if err := r.findById(ctx, "suites", id, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Repo) DeleteSuite(ctx context.Context, id Id, at int64) error {
	return r.deleteById(ctx, "suites", id, at)
}

func (r *Repo) FinishSuite(ctx context.Context, id Id, result SuiteResult,
	at int64) error {
	return r.updateById(ctx, "suites", id, bson.D{
		{"status", SuiteStatusFinished},
		{"result", result},
		{"finished_at", at},
	})
}

func (r *Repo) DisconnectSuite(ctx context.Context, id Id, at int64) error {
	return r.updateById(ctx, "suites", id, bson.D{
		{"status", SuiteStatusDisconnected},
		{"disconnected_at", at},
	})
}
