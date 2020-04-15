package database

import (
	"time"
)

type (
	CaseLinkType string
	CaseArgType  string
	CaseStatus   string
)

const (
	CaseLinkTypeIssue CaseLinkType = "issue"
	CaseLinkTypeOther              = "other"

	CaseArgTypeString  CaseArgType = "string"
	CaseArgTypeNumber              = "number"
	CaseArgTypeBoolean             = "boolean"

	CaseStatusDisabled CaseStatus = "disabled"
	CaseStatusCreated             = "created"
	CaseStatusRunning             = "running"
	CaseStatusPassed              = "passed"
	CaseStatusFailed              = "failed"
	CaseStatusErrored             = "errored"
)

type CaseLink struct {
	Type CaseLinkType `json:"type"`
	Name string       `json:"name"`
	Url  string       `json:"url"`
}

type CaseArg struct {
	Type  CaseArgType `json:"type"`
	Key   string      `json:"key"`
	Value string      `json:"value"`
}

type CaseRun struct {
	Id          interface{} `json:"id" bson:"_id,omitempty"`
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty" bson:",omitempty"`
	Attachments []string    `json:"attachments,omitempty" bson:",omitempty"`
	Links       []CaseLink  `json:"links,omitempty" bson:",omitempty"`
	Tags        []string    `json:"tags,omitempty" bson:",omitempty"`
	Args        []CaseArg   `json:"args,omitempty" bson:",omitempty"`
	Status      string      `json:"status"`
	StartedAt   int64       `json:"started_at,omitempty" bson:"started_at,omitempty"`
	FinishedAt  int64       `json:"finished_at,omitempty" bson:"finished_at,omitempty"`
	RetryOf     string      `json:"retry_of,omitempty" bson:"retry_of,omitempty"`
}

func (c *CaseRun) StartedAtTime() time.Time {
	return time.Unix(c.StartedAt, 0)
}

func (c *CaseRun) FinishedAtTime() time.Time {
	return time.Unix(c.FinishedAt, 0)
}
