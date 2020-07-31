package repo

import (
	"github.com/tidwall/buntdb"
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
	Name           string      `json:"name"`
	Tags           []string    `json:"tags"`
	PlannedCases   int64       `json:"planned_cases"`
	Status         SuiteStatus `json:"status"`
	Result         SuiteResult `json:"result"`
	DisconnectedAt int64       `json:"disconnected_at"`
	StartedAt      int64       `json:"started_at"`
	FinishedAt     int64       `json:"finished_at"`
}

type SuiteAgg struct {
	VersionedEntity
	TotalCount   int64 `json:"total_count"`
	StartedCount int64 `json:"started_count"`
}

type SuitePage struct {
	SuiteAgg
	HasMore bool    `json:"has_more"`
	Suites  []Suite `json:"suites"`
}

func (r *Repo) InsertSuite(s Suite) (id string, err error) {
	return r.insertFunc(SuiteColl, &s, func(tx *buntdb.Tx) error {
		var agg SuiteAgg
		err := r.update(tx, SuiteAggColl, "", &agg, func() {
			agg.Version++
			agg.TotalCount++
			if s.Status == SuiteStatusStarted {
				agg.StartedCount++
			}
		})
		if err != nil {
			return err
		}
		r.notifyHandlers([]Change{SuiteUpsert{Suite: s}, SuiteAggUpdate{agg}})
		return nil
	})
}

func (r *Repo) Suite(id string) (Suite, error) {
	var s Suite
	err := r.getById(SuiteColl, id, &s)
	return s, err
}
