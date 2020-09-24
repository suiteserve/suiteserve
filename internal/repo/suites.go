package repo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const suitePageLimit = 30

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
	More   bool    `json:"more"`
	Suites []Suite `json:"suites"`
}

func (r *Repo) InsertSuite(ctx context.Context, s Suite) (Id, error) {
	return r.insert(ctx, "suites", s)
}

func (r *Repo) Suite(ctx context.Context, id Id) (interface{}, error) {
	var s Suite
	if err := r.findById(ctx, "suites", id, &s); err != nil {
		return nil, err
	}
	return s, nil
}

func (r *Repo) SuitePage(ctx context.Context) (interface{}, error) {
	c, err := r.db.Collection("suites").Aggregate(ctx, mongo.Pipeline{
		{{"$match", bson.D{
			{"deleted", false},
		}}},
		{{"$sort", bson.D{
			{"started_at", -1},
			{"_id", -1},
		}}},
		{{"$limit", suitePageLimit + 1}},
		{{"$group", bson.D{
			{"_id", nil},
			{"suites", bson.D{
				{"$push", "$$ROOT"},
			}},
		}}},
		{{"$set", bson.D{
			{"more", bson.D{
				{"$eq", bson.A{
					bson.D{{"$size", "$suites"}},
					suitePageLimit + 1,
				}},
			}},
			{"suites", bson.D{
				{"$slice", bson.A{"$suites", suitePageLimit}},
			}},
		}}},
	})
	if err != nil {
		return nil, err
	}
	s := []SuitePage{}
	if err := c.All(ctx, &s); err != nil {
		return nil, err
	}
	if len(s) == 0 {
		return SuitePage{Suites: []Suite{}}, nil
	}
	return s[0], nil
}

func (r *Repo) SuitePageAfter(ctx context.Context, id Id) (interface{}, error) {
	var pivot Suite
	err := r.db.Collection("suites").FindOne(ctx, bson.D{
		{"_id", id},
	}, options.FindOne().SetProjection(bson.D{
		{"started_at", 1},
	})).Decode(&pivot)
	if err == mongo.ErrNoDocuments {
		return nil, errNotFound{}
	} else if err != nil {
		return nil, err
	}
	c, err := r.db.Collection("suites").Aggregate(ctx, mongo.Pipeline{
		{{"$match", bson.D{
			{"deleted", false},
			{"started_at", bson.D{
				{"$lte", pivot.StartedAt},
			}},
			{"$or", bson.A{
				bson.D{{"started_at", bson.D{
					{"$lt", pivot.StartedAt},
				}}},
				bson.D{{"_id", bson.D{
					{"$lt", pivot.Id},
				}}},
			}},
		}}},
		{{"$sort", bson.D{
			{"started_at", -1},
			{"_id", -1},
		}}},
		{{"$limit", suitePageLimit + 1}},
		{{"$group", bson.D{
			{"_id", nil},
			{"suites", bson.D{
				{"$push", "$$ROOT"},
			}},
		}}},
		{{"$set", bson.D{
			{"more", bson.D{
				{"$eq", bson.A{
					bson.D{{"$size", "$suites"}},
					suitePageLimit + 1,
				}},
			}},
			{"suites", bson.D{
				{"$slice", bson.A{"$suites", suitePageLimit}},
			}},
		}}},
	})
	if err != nil {
		return nil, err
	}
	s := []SuitePage{}
	if err := c.All(ctx, &s); err != nil {
		return nil, err
	}
	if len(s) == 0 {
		return SuitePage{Suites: []Suite{}}, nil
	}
	return s[0], nil
}

func (r *Repo) WatchSuites(ctx context.Context) *Watcher {
	return r.watch(ctx, "suites")
}

func (r *Repo) DeleteSuite(ctx context.Context, id Id, at int64) error {
	return r.deleteById(ctx, "suites", id, at)
}

func (r *Repo) FinishSuite(ctx context.Context, id Id, res SuiteResult, at int64) error {
	return r.updateById(ctx, "suites", id, bson.D{
		{"status", SuiteStatusFinished},
		{"result", res},
		{"finished_at", at},
	})
}

func (r *Repo) DisconnectSuite(ctx context.Context, id Id, at int64) error {
	return r.updateById(ctx, "suites", id, bson.D{
		{"status", SuiteStatusDisconnected},
		{"disconnected_at", at},
	})
}
