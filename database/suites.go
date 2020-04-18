package database

import (
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type SuiteStatus string

const (
	SuiteStatusCreated  SuiteStatus = "created"
	SuiteStatusRunning              = "running"
	SuiteStatusFinished             = "finished"
)

type SuiteFailureType struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty" bson:",omitempty"`
}

type SuiteEnvVar struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

type NewSuiteRun struct {
	FailureTypes []SuiteFailureType `json:"failure_types,omitempty" bson:"failure_types,omitempty"`
	Tags         []string           `json:"tags,omitempty" bson:",omitempty"`
	EnvVars      []SuiteEnvVar      `json:"env_vars,omitempty" bson:"env_vars,omitempty"`
	Attachments  []string           `json:"attachments,omitempty" bson:",omitempty"`
	PlannedCases int                `json:"planned_cases" bson:"planned_cases"`
	CreatedAt    int64              `json:"created_at" bson:"created_at"`
}

func (s *NewSuiteRun) CreatedAtTime() time.Time {
	return time.Unix(s.CreatedAt, 0)
}

type SuiteRun struct {
	Id           interface{} `json:"id" bson:"_id,omitempty"`
	NewSuiteRun `bson:",inline"`
	Status       SuiteStatus `json:"status"`
	FinishedAt   int64       `json:"finished_at,omitempty" bson:"finished_at,omitempty"`
}

func (s *SuiteRun) FinishedAtTime() time.Time {
	return time.Unix(s.FinishedAt, 0)
}

func (d *Database) NewSuiteRun(b []byte) (string, error) {
	var newSuiteRun NewSuiteRun
	if err := json.Unmarshal(b, &newSuiteRun); err != nil {
		return "", ErrBadJson
	}

	suiteRun := SuiteRun{
		NewSuiteRun: newSuiteRun,
		Status:      SuiteStatusCreated,
	}

	res, err := d.suites.InsertOne(newCtx(), suiteRun)
	if err != nil {
		return "", fmt.Errorf("insert suite run: %v", err)
	}
	return res.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (d *Database) SuiteRun(id string) (*SuiteRun, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("parse object id: %w", ErrNotFound)
	}

	res := d.suites.FindOne(newCtx(), bson.M{"_id": oid})

	var suiteRun SuiteRun
	if err := res.Decode(&suiteRun); err == mongo.ErrNoDocuments {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("decode suite run: %v", err)
	}
	return &suiteRun, nil
}

func (d *Database) AllSuiteRuns(since time.Time) ([]SuiteRun, error) {
	ctx := newCtx()
	cursor, err := d.suites.Find(ctx, bson.M{
		"created_at": bson.M{"$gte": since.Unix()},
	}, options.Find().SetSort(bson.D{
		{"created_at", -1},
		{"_id", -1},
	}))
	if err != nil {
		return nil, fmt.Errorf("find suite runs: %v", err)
	}

	suiteRuns := make([]SuiteRun, 0)
	if err := cursor.All(ctx, &suiteRuns); err != nil {
		return nil, fmt.Errorf("decode suite run: %v", err)
	}
	return suiteRuns, nil
}

func (d *Database) DeleteSuiteRun(id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("parse object id: %w", ErrNotFound)
	}

	res, err := d.suites.DeleteOne(newCtx(), bson.M{"_id": oid})
	if err != nil {
		return fmt.Errorf("delete suite run: %v", err)
	} else if res.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (d *Database) DeleteAllSuiteRuns() error {
	if _, err := d.suites.DeleteMany(newCtx(), bson.M{}); err != nil {
		return fmt.Errorf("delete suite runs: %v", err)
	}
	return nil
}
