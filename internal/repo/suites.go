package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
	"strconv"
	"strings"
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
	switch SuiteResult(res) {
	case SuiteResultPassed:
		fallthrough
	case SuiteResultFailed:
		*r = SuiteResult(res)
	default:
		return errBadFormat{fmt.Errorf("bad suiteresult %q", res)}
	}
	return nil
}

const (
	SuiteResultPassed SuiteResult = "passed"
	SuiteResultFailed SuiteResult = "failed"
)

type Suite struct {
	Entity          `bson:",inline"`
	VersionedEntity `bson:",inline"`
	Name            *string      `json:"name,omitempty" bson:",omitempty"`
	Tags            []string     `json:"tags,omitempty" bson:",omitempty"`
	PlannedCases    *int64       `json:"plannedCases,omitempty" bson:"planned_cases,omitempty"`
	Status          *SuiteStatus `json:"status,omitempty"`
	Result          *SuiteResult `json:"result,omitempty" bson:",omitempty"`
	DisconnectedAt  *MsTime      `json:"disconnectedAt,omitempty" bson:"disconnected_at,omitempty"`
	StartedAt       *MsTime      `json:"startedAt,omitempty" bson:"started_at"`
	FinishedAt      *MsTime      `json:"finishedAt,omitempty" bson:"finished_at,omitempty"`
}

var suiteType = reflect.TypeOf(Suite{})

type SuitePageCursor struct {
	Id        Id
	StartedAt MsTime
}

func NewSuitePageCursor(s string) (c SuitePageCursor, err error) {
	split := strings.SplitN(s, "_", 2)
	if len(split) < 2 {
		return c, errBadFormat{fmt.Errorf("bad suitepagecursor: %s", s)}
	}
	id, err := NewId(split[0])
	if err != nil {
		return c, errBadFormat{fmt.Errorf("bad suitepagecursor id: %v", err)}
	}
	ms, err := strconv.ParseInt(split[1], 10, 64)
	if err != nil {
		return c, errBadFormat{fmt.Errorf("bad suitepagecursor time: %v", err)}
	}
	return SuitePageCursor{id, NewMsTime(ms)}, nil
}

func (c SuitePageCursor) MarshalBSONValue() (bsontype.Type, []byte, error) {
	return bson.MarshalValue(c.String())
}

func (c *SuitePageCursor) UnmarshalBSONValue(_ bsontype.Type, b []byte) error {
	var s string
	if err := bson.Unmarshal(b, &s); err != nil {
		return err
	}
	cursor, err := NewSuitePageCursor(s)
	if err != nil {
		return err
	}
	*c = cursor
	return nil
}

func (c SuitePageCursor) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

func (c *SuitePageCursor) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	cursor, err := NewSuitePageCursor(s)
	if err != nil {
		return err
	}
	*c = cursor
	return nil
}

func (c SuitePageCursor) String() string {
	return fmt.Sprintf("%s_%s", c.Id, c.StartedAt)
}

type SuitePage struct {
	Next   *SuitePageCursor `json:"next,omitempty"`
	Suites []Suite          `json:"suites"`
}

func (r *Repo) InsertSuite(ctx context.Context, s Suite) (Id, error) {
	return r.insert(ctx, Suites, s)
}

func (r *Repo) Suite(ctx context.Context, id Id) (Suite, error) {
	var s Suite
	err := r.findById(ctx, Suites, id, &s)
	return s, err
}

func (r *Repo) SuitePage(ctx context.Context) (SuitePage, error) {
	return r.suitePage(ctx, bson.D{})
}

func (r *Repo) SuitePageAfter(ctx context.Context,
	cursor SuitePageCursor) (SuitePage, error) {
	return r.suitePage(ctx, bson.D{
		{"started_at", bson.D{
			{"$lte", cursor.StartedAt},
		}},
		{"$or", bson.A{
			bson.D{{"started_at", bson.D{
				{"$lt", cursor.StartedAt},
			}}},
			bson.D{{"_id", bson.D{
				{"$lt", cursor.Id},
			}}},
		}},
	})
}

func (r *Repo) suitePage(ctx context.Context, match bson.D) (SuitePage, error) {
	const limit = 100
	suitePage := SuitePage{
		Suites: []Suite{},
	}
	err := readOne(ctx, &suitePage, func() (*mongo.Cursor, error) {
		return r.db.Collection(suites).Aggregate(ctx, mongo.Pipeline{
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
				// {"next", bson.D{
				// 	{"$cond", bson.A{
				// 		bson.D{{"$eq", bson.A{
				// 			bson.D{{"$size", "$suites"}},
				// 			limit + 1,
				// 		}}},
				// 		bson.D{{"$concat", bson.A{
				// 			bson.D{{"$toString", bson.D{
				// 				{"$last", "$suites._id"},
				// 			}}},
				// 			"_",
				// 			bson.D{{"$toString", bson.D{
				// 				{"$toLong", bson.D{
				// 					{"$last", "$suites.started_at"},
				// 				}},
				// 			}}},
				// 		}}},
				// 		"$$REMOVE",
				// 	}},
				// }},
				{"suites", bson.D{
					{"$slice", bson.A{"$suites", limit}},
				}},
			}}},
		})
	})
	if err != nil {
		return SuitePage{}, err
	}
	return suitePage, nil
}

func (r *Repo) FinishSuite(ctx context.Context, id Id, res SuiteResult,
	at MsTime) error {
	return r.updateById(ctx, Suites, id, bson.D{
		{"status", SuiteStatusFinished},
		{"result", res},
		{"finished_at", at},
	})
}

func (r *Repo) DisconnectSuite(ctx context.Context, id Id, at MsTime) error {
	return r.updateById(ctx, Suites, id, bson.D{
		{"status", SuiteStatusDisconnected},
		{"disconnected_at", at},
	})
}
