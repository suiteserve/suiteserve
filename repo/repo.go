package repo

import (
	"errors"
)

type Coll string

const (
	CollAttachments Coll = "attachments"
	CollCases       Coll = "cases"
	CollLogs        Coll = "logs"
	CollSuites      Coll = "suites"
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
