package repo

import (
	"encoding/json"
	"github.com/tidwall/buntdb"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strings"
)

type SuiteStatus string

const (
	SuiteStatusRunning      SuiteStatus = "running"
	SuiteStatusPassed                   = "passed"
	SuiteStatusFailed                   = "failed"
	SuiteStatusDisconnected             = "disconnected"
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
	Id           string             `json:"id" bson:"_id,omitempty"`
	Name         string             `json:"name,omitempty" bson:",omitempty"`
	FailureTypes []SuiteFailureType `json:"failure_types,omitempty" bson:"failure_types,omitempty"`
	Tags         []string           `json:"tags,omitempty" bson:",omitempty"`
	EnvVars      []SuiteEnvVar      `json:"env_vars,omitempty" bson:"env_vars,omitempty"`
	Attachments  []string           `json:"attachments,omitempty" bson:",omitempty"`
	PlannedCases int64              `json:"planned_cases" bson:"planned_cases"`
	Status       SuiteStatus        `json:"status"`
	StartedAt    int64              `json:"started_at" bson:"started_at"`
	FinishedAt   int64              `json:"finished_at,omitempty" bson:"finished_at,omitempty"`
	Archived     bool               `json:"archived"`
	ArchivedAt   int64              `json:"archived_at,omitempty" bson:"archived_at,omitempty"`
}

type SuitePage struct {
	RunningCount  int64   `json:"running_count" bson:"running_count"`
	FinishedCount int64   `json:"finished_count" bson:"finished_count"`
	More          bool    `json:"more"`
	Suites        []Suite `json:"suites,omitempty" bson:",omitempty"`
}

type SuiteRepo interface {
	Save(Suite) (string, error)
	SaveAttachments(id string, attachments ...string) error
	SaveStatus(id string, status SuiteStatus, finishedAt *int64) error
	Page(afterId *string, n int64, includeArchived bool) (*SuitePage, error)
	Find(id string) (*Suite, error)
	FuzzyFind(fuzzyIdOrName string, includeArchived bool) ([]Suite, error)
	FindAll(includeArchived bool) ([]Suite, error)
	Archive(id string) error
	ArchiveAll() error
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
	err = r.db.ReplaceIndex("suites_name", "suites:*",
		buntdb.IndexJSON("name"))
	if err != nil {
		return nil, err
	}
	err = r.db.ReplaceIndex("suites_status", "suites:*",
		buntdb.IndexJSON("status"))
	if err != nil {
		return nil, err
	}
	err = r.db.ReplaceIndex("suites_archived", "suites:*",
		buntdb.IndexJSON("archived"))
	if err != nil {
		return nil, err
	}
	return &buntSuiteRepo{r}, nil
}

func (r *buntSuiteRepo) Save(s Suite) (string, error) {
	b, err := json.Marshal(&s)
	if err != nil {
		return "", err
	}
	var id string
	err = r.db.Update(func(tx *buntdb.Tx) error {
		id = primitive.NewObjectID().Hex()
		s.Id = id
		_, _, err = tx.Set("suites:"+id, string(b), nil)
		if err != nil {
			return err
		}
		r.changes <- Change{
			Op:      ChangeOpInsert,
			Coll:    ChangeCollSuites,
			Payload: s,
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *buntSuiteRepo) SaveAttachments(id string, attachments ...string) error {
	return r.db.Update(func(tx *buntdb.Tx) error {
		v, err := tx.Get("suites:" + id)
		if err != nil {
			return err
		}
		var s Suite
		if err := json.Unmarshal([]byte(v), &s); err != nil {
			return err
		}
		s.Attachments = append(s.Attachments, attachments...)
		b, err := json.Marshal(&s)
		if err != nil {
			return err
		}
		_, _, err = tx.Set("suites:"+id, string(b), nil)
		if err != nil {
			return err
		}
		r.changes <- Change{
			Op:      ChangeOpUpdate,
			Coll:    ChangeCollSuites,
			Payload: s,
		}
		return nil
	})
}

func (r *buntSuiteRepo) SaveStatus(id string, status SuiteStatus, finishedAt *int64) error {
	return r.db.Update(func(tx *buntdb.Tx) error {
		v, err := tx.Get("suites:" + id)
		if err != nil {
			return err
		}
		var s Suite
		if err := json.Unmarshal([]byte(v), &s); err != nil {
			return err
		}
		s.Status = status
		if finishedAt != nil {
			s.FinishedAt = *finishedAt
		}
		b, err := json.Marshal(&s)
		if err != nil {
			return err
		}
		_, _, err = tx.Set("suites:"+id, string(b), nil)
		r.changes <- Change{
			Op:      ChangeOpUpdate,
			Coll:    ChangeCollSuites,
			Payload: s,
		}
		return err
	})
}

func (r *buntSuiteRepo) Page(afterId *string, n int64, includeArchived bool) (*SuitePage, error) {
	var page SuitePage
	err := r.db.View(func(tx *buntdb.Tx) error {
		var running int64
		var finished int64
		suites := make([]Suite, 0)
		var err error
		if afterId != nil {
			if err := tx.DescendLessOrEqual("suites_id", *afterId, func(k, v string) bool {
				var s Suite
				if err = json.Unmarshal([]byte(v), &s); err != nil {
					return false
				}
				if s.Id != *afterId && (includeArchived || !s.Archived) {
					suites = append(suites, s)
				}
				// TODO: we may not want to limit ourselves to the max array length;
				//  this concerns memory efficiency as well as the max number of
				//  results allowed
				return int64(len(suites)) < n
			}); err != nil {
				return err
			}
		} else {
			if err := tx.Descend("suites_id", func(k, v string) bool {
				var s Suite
				if err = json.Unmarshal([]byte(v), &s); err != nil {
					return false
				}
				if includeArchived || !s.Archived {
					suites = append(suites, s)
				}
				// TODO: reduce duplication
				return int64(len(suites)) < n
			}); err != nil {
				return err
			}
		}
		if err != nil {
			return err
		}

		if err := tx.AscendEqual("suites_status", string(SuiteStatusRunning), func(k, v string) bool {
			running++
			return true
		}); err != nil {
			return err
		}
		if err := tx.AscendEqual("suites_status", SuiteStatusPassed, func(k, v string) bool {
			finished++
			return true
		}); err != nil {
			return err
		}
		if err := tx.AscendEqual("suites_status", SuiteStatusFailed, func(k, v string) bool {
			finished++
			return true
		}); err != nil {
			return err
		}
		var more bool
		if n > 0 && int64(len(suites)) == n {
			afterId := suites[n-1].Id
			// TODO: can weave in with the previous DescendLessOrEqual() call
			if err := tx.DescendLessOrEqual("suites_id", afterId, func(k, v string) bool {
				var s Suite
				if err = json.Unmarshal([]byte(v), &s); err != nil {
					return false
				}
				if s.Id != afterId {
					more = true
					return false
				}
				return true
			}); err != nil {
				return err
			}
		}
		page = SuitePage{
			RunningCount:  running,
			FinishedCount: finished,
			More:          more,
			Suites:        suites,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &page, nil
}

func (r *buntSuiteRepo) Find(id string) (*Suite, error) {
	var s Suite
	err := r.db.View(func(tx *buntdb.Tx) error {
		v, err := tx.Get("suites:" + id)
		if err != nil {
			return err
		}
		return json.Unmarshal([]byte(v), &s)
	})
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *buntSuiteRepo) FuzzyFind(fuzzyIdOrName string, includeArchived bool) ([]Suite, error) {
	var suites []Suite
	err := r.db.View(func(tx *buntdb.Tx) error {
		suites = make([]Suite, 0)
		var err error
		if err := tx.Ascend("", func(k, v string) bool {
			var s Suite
			if err = json.Unmarshal([]byte(v), &s); err != nil {
				return false
			}
			if (includeArchived || !s.Archived) &&
				(strings.Contains(s.Id, fuzzyIdOrName) || strings.Contains(s.Name, fuzzyIdOrName)) {
				suites = append(suites, s)
			}
			return true
		}); err != nil {
			return err
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return suites, nil
}

func (r *buntSuiteRepo) FindAll(includeArchived bool) ([]Suite, error) {
	values := make([]string, 0)
	err := r.db.View(func(tx *buntdb.Tx) error {
		iterator := func(k, v string) bool {
			values = append(values, v)
			return true
		}
		if includeArchived {
			return tx.Ascend("", iterator)
		}
		return tx.AscendEqual("suites_archived", "false", iterator)
	})
	if err != nil {
		return nil, err
	}
	suites := make([]Suite, len(values))
	for i, v := range values {
		var s Suite
		if err := json.Unmarshal([]byte(v), &s); err != nil {
			return nil, err
		}
		suites[i] = s
	}
	return suites, nil
}

func (r *buntSuiteRepo) Archive(id string) error {
	return r.db.Update(func(tx *buntdb.Tx) error {
		v, err := tx.Get("suites:" + id)
		if err != nil {
			return err
		}
		var s Suite
		if err := json.Unmarshal([]byte(v), &s); err != nil {
			return err
		}
		s.Archived = true
		s.ArchivedAt = nowTimeMillis()
		b, err := json.Marshal(&s)
		if err != nil {
			return err
		}
		_, _, err = tx.Set("suites:"+id, string(b), nil)
		if err != nil {
			return err
		}
		r.changes <- Change{
			Op:      ChangeOpUpdate,
			Coll:    ChangeCollSuites,
			Payload: s,
		}
		return nil
	})
}

func (r *buntSuiteRepo) ArchiveAll() error {
	return r.db.Update(func(tx *buntdb.Tx) error {
		values := make([]string, 0)
		err := tx.AscendEqual("suites_archived", "false", func(k, v string) bool {
			values = append(values, v)
			return true
		})
		if err != nil {
			return err
		}
		deletedAt := nowTimeMillis()
		for _, v := range values {
			var s Suite
			if err := json.Unmarshal([]byte(v), &s); err != nil {
				return err
			}
			s.Archived = true
			s.ArchivedAt = deletedAt
			b, err := json.Marshal(&s)
			if err != nil {
				return err
			}
			if _, _, err := tx.Set("suites:"+s.Id, string(b), nil); err != nil {
				return err
			}
			r.changes <- Change{
				Op:      ChangeOpUpdate,
				Coll:    ChangeCollSuites,
				Payload: s,
			}
		}
		return nil
	})
}
