package repo

import (
	"errors"
)

type Coll string

const (
	// CollAttachments is the collection of attachments.
	CollAttachments Coll = "attachments"
	// CollCases is the collection of cases.
	CollCases Coll = "cases"
	// CollLogs is the collection of logs.
	CollLogs Coll = "logs"
	// CollSuites is the collection of suites.
	CollSuites Coll = "suites"
	// CollSuiteAggs is the collection of aggregations of the suite collection.
	CollSuiteAggs Coll = "suite_aggs"
)

type SavedEntity struct {
	Id string `json:"id" bson:"_id,omitempty"`
}

type VersionedEntity struct {
	Version int64 `json:"version"`
}

type SoftDeleteEntity struct {
	Deleted   bool  `json:"deleted"`
	DeletedAt int64 `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
}

var (
	ErrExpired          = errors.New("expired")
	ErrNotFound         = errors.New("not found")
	ErrNotReconnectable = errors.New("not reconnectable")
)
