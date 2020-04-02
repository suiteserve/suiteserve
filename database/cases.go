package database

import (
	"time"
)

type CaseLinkType string

const (
	CaseLinkTypeIssue CaseLinkType = "issue"
	CaseLinkTypeOther              = "other"
)

type CaseArgType string

const (
	CaseArgTypeString  CaseArgType = "string"
	CaseArgTypeNumber              = "number"
	CaseArgTypeBoolean             = "boolean"
)

type CaseStatus string

const (
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
	URL  string       `json:"url"`
}

type CaseArg struct {
	Type  CaseArgType `json:"type"`
	Key   string      `json:"key"`
	Value string      `json:"value"`
}

type CaseRun struct {
	Id          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Links       []CaseLink `json:"links"`
	Tags        []string   `json:"tags"`
	Args        []CaseArg  `json:"args"`
	Status      string     `json:"status"`
	StartedAt   time.Time  `json:"started_at,omitempty"`
	FinishedAt  time.Time  `json:"finished_at,omitempty"`
	RetryOf     string     `json:"retry_of,omitempty"`
}
