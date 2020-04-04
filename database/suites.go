package database

import (
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	FailureTypes []SuiteFailureType `json:"failure_types" bson:"failureTypes,omitempty"`
	Tags         []string           `json:"tags" bson:"tags,omitempty"`
	EnvVars      []SuiteEnvVar      `json:"env_vars" bson:"envVars,omitempty"`
	PlannedCases int                `json:"planned_cases" bson:"plannedCases"`
	CreatedAt    time.Time          `json:"created_at" bson:"createdAt"`
}

type SuiteRun struct {
	NewSuiteRun `bson:",inline"`

	Id         string      `json:"id" bson:"_id,omitempty"`
	Status     SuiteStatus `json:"status" bson:"status"`
	FinishedAt time.Time   `json:"finished_at,omitempty" bson:"finishedAt,omitempty"`
}

const suiteRunsCollection = "suiteRuns"

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
