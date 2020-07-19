package repo

import (
	"fmt"
	"github.com/tidwall/buntdb"
	"github.com/tidwall/gjson"
)

type SuiteStatus string

const (
	SuiteStatusUnknown      SuiteStatus = "unknown"
	SuiteStatusStarted      SuiteStatus = "started"
	SuiteStatusFinished     SuiteStatus = "finished"
	SuiteStatusDisconnected SuiteStatus = "disconnected"
)

type SuiteResult string

const (
	SuiteResultUnknown SuiteResult = "unknown"
	SuiteResultPassed  SuiteResult = "passed"
	SuiteResultFailed  SuiteResult = "failed"
)

type Suite struct {
	Entity
	VersionedEntity
	SoftDeleteEntity
	Name           string      `json:"name,omitempty"`
	Tags           []string    `json:"tags,omitempty"`
	PlannedCases   int64       `json:"planned_cases,omitempty"`
	Status         SuiteStatus `json:"status"`
	Result         SuiteResult `json:"result"`
	DisconnectedAt int64       `json:"disconnected_at,omitempty"`
	StartedAt      int64       `json:"started_at"`
	FinishedAt     int64       `json:"finished_at,omitempty"`
}

type SuitePage struct {
	CountVersion int64    `json:"count_version"`
	TotalCount   int64    `json:"total_count"`
	RunningCount int64    `json:"running_count"`
	HasMore      bool     `json:"has_more"`
	Suites       []*Suite `json:"suites"`
}

func (r *Repo) InsertSuite(s Suite) (id string, err error) {
	return r.insert(SuiteColl, &s)
}

func (r *Repo) Suite(id string) (*Suite, error) {
	var s Suite
	return &s, r.getById(SuiteColl, id, &s)
}

func (r *Repo) SuitePage(fromId string, limit int) (*SuitePage, error) {
	var vals []string
	var page SuitePage
	err := r.db.View(func(tx *buntdb.Tx) error {
		var err error
		vals, page.HasMore, err = getSuites(tx, fromId, limit)
		if err != nil {
			return err
		}
		if page.CountVersion, err = getInt(tx, suiteKeyVersion); err != nil {
			return err
		}
		if page.TotalCount, err = getInt(tx, suiteKeyTotal); err != nil {
			return err
		}
		page.RunningCount, err = getInt(tx, suiteKeyRunning)
		return err
	})
	if err == buntdb.ErrNotFound {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}
	page.Suites = make([]*Suite, len(vals))
	unmarshalJsonVals(vals, func(i int) interface{} {
		return &page.Suites[i]
	})
	return &page, nil
}

func getSuites(tx *buntdb.Tx, fromId string, limit int) (vals []string, hasMore bool, err error) {
	var startedAt int64
	less := func(string) bool {
		return true
	}
	if fromId != "" {
		var err error
		startedAt, err = getJsonInt(tx, SuiteColl, fromId, "started_at")
		if err != nil {
			return nil, false, err
		}
		less = func(v string) bool {
			return gjson.Get(v, "started_at").Int() < startedAt
		}
	}
	itr := newPageItr(limit, &vals, &hasMore, less)
	err = descendSuites(tx, fromId, startedAt, itr)
	return
}

func descendSuites(tx *buntdb.Tx, fromId string, startedAt int64, itr itr) error {
	if fromId == "" {
		return tx.Descend(suiteIndexTimestamp, itr)
	}
	pivot := fmt.Sprintf(`{"started_at": %d}`, startedAt)
	return tx.DescendLessOrEqual(suiteIndexTimestamp, pivot, itr)
}
