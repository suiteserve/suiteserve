package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

type SuiteStatus string

const (
	SuiteStatusStarted      SuiteStatus = "started"
	SuiteStatusFinished     SuiteStatus = "finished"
	SuiteStatusDisconnected SuiteStatus = "disconnected"
)

type SuiteResult string

func (r *SuiteResult) UnmarshalJSON(b []byte) error {
	var res string
	if err := json.Unmarshal(b, &res); err != nil {
		return err
	}
	*r = SuiteResult(res)
	switch *r {
	case SuiteResultPassed:
	case SuiteResultFailed:
	default:
		return errBadFormat{fmt.Errorf("unknown suiteresult %q", r)}
	}
	return nil
}

const (
	SuiteResultPassed SuiteResult = "passed"
	SuiteResultFailed SuiteResult = "failed"
)

type Suite struct {
	Entity           `bson:",inline"`
	VersionedEntity  `bson:",inline"`
	SoftDeleteEntity `bson:",inline"`
	Name             *string      `json:"name,omitempty" bson:",omitempty"`
	Tags             []string     `json:"tags,omitempty" bson:",omitempty"`
	PlannedCases     *int64       `json:"plannedCases,omitempty" bson:"planned_cases,omitempty"`
	Status           *SuiteStatus `json:"status"`
	Result           *SuiteResult `json:"result,omitempty" bson:",omitempty"`
	DisconnectedAt   *Time        `json:"disconnectedAt,omitempty" bson:"disconnected_at,omitempty"`
	StartedAt        *Time        `json:"startedAt" bson:"started_at"`
	FinishedAt       *Time        `json:"finishedAt,omitempty" bson:"finished_at,omitempty"`
}

var suiteType = reflect.TypeOf(Suite{})

type SuitePage struct {
	More   bool    `json:"more"`
	Suites []Suite `json:"suites"`
}

func (r *Repo) InsertSuite(ctx context.Context, s Suite) (Id, error) {
	return r.insert(ctx, suitesColl, s)
}

func (r *Repo) Suite(ctx context.Context, id Id) (interface{}, error) {
	return r.findById(ctx, suitesColl, id, &Suite{})
}

func (r *Repo) SuitePage(ctx context.Context) (interface{}, error) {
	return r.suitePage(ctx, bson.D{
		{"deleted", false},
	})
}

func (r *Repo) SuitePageAfter(ctx context.Context, id Id) (interface{}, error) {
	var pivot Suite
	_, err := r.findByIdProj(ctx, suitesColl, id, bson.D{
		{"started_at", 1},
	}, &pivot)
	if err != nil {
		return nil, err
	}
	return r.suitePage(ctx, bson.D{
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
	})
}

func (r *Repo) suitePage(ctx context.Context,
	match bson.D) (res interface{}, err error) {
	const limit = 100
	return readOne(ctx, &SuitePage{Suites: []Suite{}},
		func() (*mongo.Cursor, error) {
			return r.db.Collection(suitesColl).Aggregate(ctx, mongo.Pipeline{
				{{"$match", match}},
				{{"$sort", bson.D{
					{"started_at", -1},
					{"_id", -1},
				}}},
				{{"$limit", limit + 1}},
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
							limit + 1,
						}},
					}},
					{"suites", bson.D{
						{"$slice", bson.A{"$suites", limit}},
					}},
				}}},
			})
		})
}

func (r *Repo) WatchSuites(ctx context.Context) *Watcher {
	return r.watch(ctx, suitesColl)
}

func (r *Repo) DeleteSuite(ctx context.Context, id Id, at Time) error {
	return r.deleteById(ctx, suitesColl, id, at)
}

func (r *Repo) FinishSuite(ctx context.Context, id Id, res SuiteResult,
	at Time) error {
	return r.updateById(ctx, suitesColl, id, bson.D{
		{"status", SuiteStatusFinished},
		{"result", res},
		{"finished_at", at},
	})
}

func (r *Repo) DisconnectSuite(ctx context.Context, id Id, at Time) error {
	return r.updateById(ctx, suitesColl, id, bson.D{
		{"status", SuiteStatusDisconnected},
		{"disconnected_at", at},
	})
}
