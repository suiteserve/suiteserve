package repo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	PlannedCases     int64       `json:"planned_cases,omitempty" bson:"planned_cases,omitempty"`
	Status           SuiteStatus `json:"status"`
	Result           SuiteResult `json:"result,omitempty" bson:",omitempty"`
	DisconnectedAt   int64       `json:"disconnected_at,omitempty" bson:"disconnected_at,omitempty"`
	StartedAt        int64       `json:"started_at" bson:"started_at"`
	FinishedAt       int64       `json:"finished_at,omitempty" bson:"finished_at,omitempty"`
}

type SuitePage struct {
	NextId        Id      `json:"next_id" bson:"next_id"`
	TotalCount    int64   `json:"total_count" bson:"total_count"`
	FinishedCount int64   `json:"finished_count" bson:"finished_count"`
	Suites        []Suite `json:"suites"`
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

func (r *Repo) SuitePage(ctx context.Context) (*SuitePage, error) {
	c, err := r.db.Collection("suites").Aggregate(ctx, mongo.Pipeline{
		{{}},
	}, options.Aggregate().SetHint("latest"))
	if err != nil {
		return nil, err
	}
}

func (r *Repo) WatchSuites(ctx context.Context) *Watcher {
	return r.watch(ctx, "suites")
}

func (r *Repo) DeleteSuite(ctx context.Context, id Id, at int64) error {
	// var old Suite
	// r.db.Collection("asdf").Cou
	// err := r.findAndUpdateById(ctx, "suites", id, bson.D{
	// 	{"deleted", true},
	// 	{"deleted_at", at},
	// }, bson.D{
	// 	{"deleted", 1},
	// 	{"status", 1},
	// }, &old)
	// if err != nil {
	// 	return err
	// }
	// if old.Deleted {
	// 	// was already deleted, so do nothing
	// 	return nil
	// }
	// // newly deleted
	// incStarted := 0
	// if old.Status == SuiteStatusStarted {
	// 	incStarted = -1
	// }
	// return r.updateSuiteAgg(ctx, -1, 0)
	return nil
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

// func (r *Repo) updateSuiteAgg(ctx context.Context, incTotal, incStarted int) error {
// 	_, err := r.db.Collection("suites_agg").UpdateOne(ctx, bson.D{}, bson.D{
// 		{"$inc", bson.D{
// 			{"total", incTotal},
// 			{"started", incStarted},
// 		}},
// 	}, options.Update().SetUpsert(true))
// 	return err
// }
