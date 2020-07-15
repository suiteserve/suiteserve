package repo

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

type UnsavedSuite struct {
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
	DisconnectedAt   int64              `json:"disconnected_at,omitempty" bson:"disconnected_at,omitempty"`
}

type Suite struct {
	SavedEntity     `bson:",inline"`
	VersionedEntity `bson:",inline"`
	UnsavedSuite    `bson:",inline"`
}

type SuiteAggs struct {
	VersionedEntity `bson:",inline"`
	Running         int64 `json:"running"`
	Finished        int64 `json:"finished"`
}

type SuitePage struct {
	Aggs   SuiteAggs `json:"aggs" bson:",inline"`
	NextId *string   `json:"next_id" bson:"next_id,omitempty"`
	Suites []Suite   `json:"suites,omitempty" bson:",omitempty"`
}
