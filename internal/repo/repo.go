package repo

import (
	"encoding/json"
	"github.com/asdine/storm/v3"
	bolt "go.etcd.io/bbolt"
)

const (
	attachmentBkt = "attachments"
	attachmentSuiteOwnerIdxBkt = "attachments/suite_owner"
	attachmentCaseOwnerIdxBkt = "attachments/case_owner"
	caseBkt = "cases"
	logBkt = "logs"
	suiteBkt = "suites"
)

const (
	suiteAggKey = iota
)

type Entity struct {
	string `json:"id"`
}

type VersionedEntity struct {
	Version int64 `json:"version"`
}

type SoftDeleteEntity struct {
	Deleted   bool  `json:"deleted,omitempty"`
	DeletedAt int64 `json:"deleted_at,omitempty"`
}

type Repo struct {
	db *bolt.DB
	cb changeBroker
}

func Open(filename string) (*Repo, error) {
	db, err := bolt.Open(filename, 0600, nil)
	if err != nil {
		return nil, err
	}
	return &Repo{db: db}, nil
}

func (r *Repo) Close() error {
	return r.db.Close()
}

func updateAgg(tx storm.Node, key interface{}, v interface{},
	updateFn func()) error {
	const bucket = "agg"
	if err := tx.Get(bucket, key, v); err != nil && err != storm.ErrNotFound {
		return err
	}
	updateFn()
	return tx.Set(bucket, key, v)
}

func mustMarshalJson(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
