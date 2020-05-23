package repo

import "context"

type SuiteStatus string

const (
	SuiteStatusRunning      SuiteStatus = "running"
	SuiteStatusPassed       SuiteStatus = "passed"
	SuiteStatusFailed       SuiteStatus = "failed"
	SuiteStatusDisconnected SuiteStatus = "disconnected"
)

type SuiteFailureType struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty" bson:",omitempty"`
}

type SuiteEnvVar struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

type Suite struct {
	SoftDeleteEntity `bson:",inline"`
	Name             string             `json:"name,omitempty" bson:",omitempty"`
	FailureTypes     []SuiteFailureType `json:"failure_types,omitempty" bson:"failure_types,omitempty"`
	Tags             []string           `json:"tags,omitempty" bson:",omitempty"`
	EnvVars          []SuiteEnvVar      `json:"env_vars,omitempty" bson:"env_vars,omitempty"`
	Attachments      []string           `json:"attachments,omitempty" bson:",omitempty"`
	PlannedCases     int64              `json:"planned_cases" bson:"planned_cases"`
	Status           SuiteStatus        `json:"status"`
	StartedAt        int64              `json:"started_at" bson:"started_at"`
	FinishedAt       int64              `json:"finished_at,omitempty" bson:"finished_at,omitempty"`
}

type SuitePage struct {
	RunningCount  int64   `json:"running_count" bson:"running_count"`
	FinishedCount int64   `json:"finished_count" bson:"finished_count"`
	NextId        *string `json:"next_id" bson:"next_id,omitempty"`
	Suites        []Suite `json:"suites" bson:",omitempty"`
}

type SuiteRepo interface {
	Save(ctx context.Context, s Suite) (string, error)
	SaveAttachment(ctx context.Context, id string, attachmentId string) error
	SaveStatus(ctx context.Context, id string, status SuiteStatus, finishedAt *int64) error
	Page(ctx context.Context, fromId *string, n int64, includeDeleted bool) (*SuitePage, error)
	Find(ctx context.Context, id string) (*Suite, error)
	FuzzyFind(ctx context.Context, fuzzyIdOrName string, includeDeleted bool) ([]Suite, error)
	FindAll(ctx context.Context, includeDeleted bool) ([]Suite, error)
	Delete(ctx context.Context, id string, at int64) error
	DeleteAll(ctx context.Context, at int64) error
}
