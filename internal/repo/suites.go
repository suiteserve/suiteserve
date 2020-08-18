package repo

import (
	"github.com/golang/protobuf/protoc-gen-go/generator"
	fieldmask_utils "github.com/mennanov/fieldmask-utils"
	"github.com/tidwall/buntdb"
)

type SuiteStatus string

const (
	SuiteStatusStarted      SuiteStatus = "started"
	SuiteStatusFinished     SuiteStatus = "finished"
	SuiteStatusDisconnected SuiteStatus = "disconnected"
)

type SuiteResult string

const (
	SuiteResultPassed SuiteResult = "passed"
	SuiteResultFailed SuiteResult = "failed"
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

func (r *Repo) InsertSuite(s Suite) (id string, err error) {
	return r.insertFunc(SuiteColl, &s, func(tx *buntdb.Tx) error {
		var agg SuiteAgg
		err := r.updateTx(tx, SuiteAggColl, "", &agg, func() error {
			agg.Version++
			agg.TotalCount++
			if s.Status == SuiteStatusStarted {
				agg.StartedCount++
			}
			return nil
		})
		if err != nil {
			return err
		}
		r.notifyHandlers([]Change{SuiteUpsert{Suite: s}, SuiteAggUpdate{agg}})
		return nil
	})
}

func (r *Repo) UpdateSuite(s1 Suite, mask Mask) error {
	var s0 Suite
	return r.update(SuiteColl, s1.Id, &s0, func() error {
		fm, err := fieldmask_utils.MaskFromPaths(mask, generator.CamelCase)
		if err != nil {
			return errBadMask{err}
		}
		if err := fieldmask_utils.StructToStruct(fm, &s1, &s0); err != nil {
			return errBadMask{err}
		}
		return nil
	})
}

func (r *Repo) Suite(id string) (Suite, error) {
	var s Suite
	err := r.byId(SuiteColl, id, &s)
	return s, err
}
