package database

import (
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type SuiteStatus string

const (
	SuiteStatusCreated  SuiteStatus = "created"
	SuiteStatusRunning              = "running"
	SuiteStatusFinished             = "finished"
)

type SuiteEnvType string

const (
	SuiteEnvTypeString  SuiteEnvType = "string"
	SuiteEnvTypeNumber               = "number"
	SuiteEnvTypeBoolean              = "boolean"
)

type SuiteFailureType struct {
	Name        string `json:"name" bson:"name"`
	Description string `json:"description,omitempty" bson:"description,omitempty"`
}

type SuiteEnvVar struct {
	Type  SuiteEnvType `json:"type" bson:"type"`
	Key   string       `json:"key" bson:"key"`
	Value string       `json:"value" bson:"value"`
}

type NewSuiteRun struct {
	FailureTypes []SuiteFailureType `json:"failure_types,omitempty" bson:"failure_types,omitempty"`
	Tags         []string           `json:"tags,omitempty" bson:"tags,omitempty"`
	EnvVars      []SuiteEnvVar      `json:"env_vars,omitempty" bson:"env_vars,omitempty"`
	PlannedCases int                `json:"planned_cases" bson:"planned_cases"`
	CreatedAt    int64              `json:"created_at" bson:"created_at"`
}

func (s *NewSuiteRun) CreatedAtTime() time.Time {
	return time.Unix(s.CreatedAt, 0)
}

type SuiteRun struct {
	Id          string `json:"id" bson:"_id,omitempty"`
	NewSuiteRun `bson:",inline"`
	Status      SuiteStatus `json:"status" bson:"status"`
	FinishedAt  int64       `json:"finished_at,omitempty" bson:"finished_at,omitempty"`
}

func (s *SuiteRun) FinishedAtTime() time.Time {
	return time.Unix(s.FinishedAt, 0)
}

const suiteRunsCollection = "suite_runs"

func (d *Database) NewSuiteRun(b []byte) (string, error) {
	var newSuiteRun NewSuiteRun
	if err := json.Unmarshal(b, &newSuiteRun); err != nil {
		return "", ErrBadJson
	}

	suiteRun := SuiteRun{
		NewSuiteRun: newSuiteRun,
		Status:      SuiteStatusCreated,
	}

	res, err := d.mgoDb.Collection(suiteRunsCollection).InsertOne(newCtx(), suiteRun)
	if err != nil {
		return "", fmt.Errorf("failed to insert new suite run: %v", err)
	}
	return res.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (d *Database) SuiteRun(id string) (*SuiteRun, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrNotFound
	}

	res := d.mgoDb.Collection(suiteRunsCollection).FindOne(newCtx(), bson.M{"_id": oid})
	if err := res.Err(); err == mongo.ErrNoDocuments {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("failed to find suite run: %v", err)
	}

	var suiteRun SuiteRun
	if err := res.Decode(&suiteRun); err != nil {
		return nil, fmt.Errorf("failed to decode suite run result: %v", err)
	}
	return &suiteRun, nil
}
