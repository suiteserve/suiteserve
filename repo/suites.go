package repo

import (
	"github.com/tidwall/buntdb"
	"github.com/tidwall/gjson"
	"strconv"
	"strings"
)

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
	*SoftDeleteEntity `bson:",inline"`
	Name              string             `json:"name,omitempty" bson:",omitempty"`
	FailureTypes      []SuiteFailureType `json:"failure_types,omitempty" bson:"failure_types,omitempty"`
	Tags              []string           `json:"tags,omitempty" bson:",omitempty"`
	EnvVars           []SuiteEnvVar      `json:"env_vars,omitempty" bson:"env_vars,omitempty"`
	Attachments       []string           `json:"attachments,omitempty" bson:",omitempty"`
	PlannedCases      int64              `json:"planned_cases" bson:"planned_cases"`
	Status            SuiteStatus        `json:"status"`
	StartedAt         int64              `json:"started_at" bson:"started_at"`
	FinishedAt        int64              `json:"finished_at,omitempty" bson:"finished_at,omitempty"`
}

type SuitePage struct {
	RunningCount  int64   `json:"running_count" bson:"running_count"`
	FinishedCount int64   `json:"finished_count" bson:"finished_count"`
	NextId        *string `json:"next_id" bson:"next_id,omitempty"`
	Suites        []Suite `json:"suites" bson:",omitempty"`
}

type SuiteRepo interface {
	Save(Suite) (string, error)
	SaveAttachments(id string, attachments ...string) error
	SaveStatus(id string, status SuiteStatus, finishedAt *int64) error
	Page(fromId *string, n int64, includeDeleted bool) (*SuitePage, error)
	Find(id string) (*Suite, error)
	FuzzyFind(fuzzyIdOrName string, includeDeleted bool) ([]Suite, error)
	FindAll(includeDeleted bool) ([]Suite, error)
	Delete(id string) error
	DeleteAll() error
}

type buntSuiteRepo struct {
	*buntRepo
}

func (r *buntRepo) newSuiteRepo() (*buntSuiteRepo, error) {
	err := r.db.ReplaceIndex("suites_id", "suites:*",
		buntdb.IndexJSON("id"))
	if err != nil {
		return nil, err
	}
	err = r.db.ReplaceIndex("suites_status", "suites:*",
		buntdb.IndexJSON("status"))
	if err != nil {
		return nil, err
	}
	err = r.db.ReplaceIndex("suites_deleted", "suites:*",
		buntdb.IndexJSON("deleted"))
	if err != nil {
		return nil, err
	}
	return &buntSuiteRepo{r}, nil
}

func (r *buntSuiteRepo) Save(s Suite) (string, error) {
	return r.save(&s, SuiteCollection)
}

func (r *buntSuiteRepo) SaveAttachments(id string, attachments ...string) error {
	m := make(map[string]interface{})
	for _, a := range attachments {
		m["attachments.-1"] = a
	}
	return r.set(CaseCollection, id, m)
}

func (r *buntSuiteRepo) SaveStatus(id string, status SuiteStatus, finishedAt *int64) error {
	return r.set(SuiteCollection, id, map[string]interface{}{
		"status":      status,
		"finished_at": finishedAt,
	})
}

func (r *buntSuiteRepo) Page(fromId *string, n int64, includeDeleted bool) (*SuitePage, error) {
	var running int64
	var finished int64
	var nextId *string
	values := make([]string, 0)
	err := r.db.View(func(tx *buntdb.Tx) error {
		iterator := func(k, v string) bool {
			id := gjson.Get(v, "id").String()
			if int64(len(values)) == n {
				nextId = &id
				return false
			}
			if includeDeleted || !gjson.Get(v, "deleted").Bool() {
				values = append(values, v)
			}
			return true
		}
		var err error
		if fromId == nil {
			err = tx.Descend("suites_id", iterator)
		} else {
			escapedFromId := strconv.Quote(*fromId)
			err = tx.DescendLessOrEqual("suites_id", `{"id":`+escapedFromId+`}`, iterator)
		}
		if err != nil {
			return err
		}

		running, err = r.count(tx, "suites_status", "status", string(SuiteStatusRunning))
		if err != nil {
			return err
		}
		passed, err := r.count(tx, "suites_status", "status", string(SuiteStatusPassed))
		if err != nil {
			return err
		}
		failed, err := r.count(tx, "suites_status", "status", string(SuiteStatusFailed))
		if err != nil {
			return err
		}
		finished = passed + failed
		return nil
	})
	if err != nil {
		return nil, err
	}
	var suites []Suite
	if err := r.valuesToSlice(values, &suites); err != nil {
		return nil, err
	}
	return &SuitePage{
		RunningCount:  running,
		FinishedCount: finished,
		NextId:        nextId,
		Suites:        suites,
	}, nil
}

func (r *buntSuiteRepo) count(tx *buntdb.Tx, index, k, v string) (int64, error) {
	var n int64
	err := tx.AscendEqual(index, `{"`+k+`":"`+v+`"}`, func(k, v string) bool {
		n++
		return true
	})
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (r *buntSuiteRepo) Find(id string) (*Suite, error) {
	var s Suite
	if err := r.find(SuiteCollection, id, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *buntSuiteRepo) FuzzyFind(fuzzyIdOrName string, includeDeleted bool) ([]Suite, error) {
	values := make([]string, 0)
	err := r.db.View(func(tx *buntdb.Tx) error {
		return tx.Ascend("", func(k, v string) bool {
			id := gjson.Get(v, "id").String()
			name := gjson.Get(v, "name").String()
			deleted := gjson.Get(v, "deleted").Bool()
			idMatched := strings.Contains(id, fuzzyIdOrName)
			nameMatched := strings.Contains(name, fuzzyIdOrName)
			if (includeDeleted || !deleted) && (idMatched || nameMatched) {
				values = append(values, v)
			}
			return true
		})
	})
	if err != nil {
		return nil, err
	}
	var suites []Suite
	if err := r.valuesToSlice(values, &suites); err != nil {
		return nil, err
	}
	return suites, nil
}

func (r *buntSuiteRepo) FindAll(includeDeleted bool) ([]Suite, error) {
	var suites []Suite
	if err := r.findAll("suites_deleted", includeDeleted, &suites); err != nil {
		return nil, err
	}
	return suites, nil
}

func (r *buntSuiteRepo) Delete(id string) error {
	return r.delete(SuiteCollection, id)
}

func (r *buntSuiteRepo) DeleteAll() error {
	return r.deleteAll(SuiteCollection, "suites_deleted")
}
