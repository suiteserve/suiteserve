package persist

import "time"

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
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type SuiteEnvVar struct {
	Type  SuiteEnvType `json:"type"`
	Key   string       `json:"key"`
	Value string       `json:"value"`
}

type SuiteRun struct {
	Id           string             `json:"id"`
	FailureTypes []SuiteFailureType `json:"failure_types"`
	Tags         []string           `json:"tags"`
	EnvVars      []SuiteEnvVar      `json:"env"`
	PlannedCases int                `json:"planned_cases"`
	Status       SuiteStatus        `json:"status"`
	CreatedAt    time.Time          `json:"created_at"`
	FinishedAt   time.Time          `json:"finished_at,omitempty"`
}