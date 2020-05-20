package repo

import (
	"encoding/json"
	"github.com/tidwall/buntdb"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type (
	CaseLinkType string
	CaseStatus   string
)

const (
	CaseLinkTypeIssue CaseLinkType = "issue"
	CaseLinkTypeOther CaseLinkType = "other"

	CaseStatusCreated  CaseStatus = "created"
	CaseStatusDisabled CaseStatus = "disabled"
	CaseStatusRunning  CaseStatus = "running"
	CaseStatusPassed   CaseStatus = "passed"
	CaseStatusFailed   CaseStatus = "failed"
	CaseStatusErrored  CaseStatus = "errored"
)

type CaseLink struct {
	Type CaseLinkType `json:"type"`
	Name string       `json:"name"`
	Url  string       `json:"url"`
}

type CaseArg struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

type Case struct {
	Id          string     `json:"id" bson:"_id,omitempty"`
	Suite       string     `json:"suite"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty" bson:",omitempty"`
	Tags        []string   `json:"tags,omitempty" bson:",omitempty"`
	Num         int64      `json:"num"`
	Links       []CaseLink `json:"links,omitempty" bson:",omitempty"`
	Args        []CaseArg  `json:"args,omitempty" bson:",omitempty"`
	Attachments []string   `json:"attachments,omitempty" bson:",omitempty"`
	Status      CaseStatus `json:"status"`
	Flaky       bool       `json:"flaky"`
	CreatedAt   int64      `json:"created_at" bson:"created_at"`
	StartedAt   int64      `json:"started_at,omitempty" bson:"started_at,omitempty"`
	FinishedAt  int64      `json:"finished_at,omitempty" bson:"finished_at,omitempty"`
}

type CaseRepoSaveStatusOptions struct {
	flaky      *bool
	startedAt  *int64
	finishedAt *int64
}

func (o *CaseRepoSaveStatusOptions) Flaky(flaky bool) {
	o.flaky = &flaky
}

func (o *CaseRepoSaveStatusOptions) StartedAt(startedAt int64) {
	o.startedAt = &startedAt
}

func (o *CaseRepoSaveStatusOptions) FinishedAt(finishedAt int64) {
	o.finishedAt = &finishedAt
}

type CaseRepo interface {
	Save(Case) (string, error)
	SaveAttachments(id string, attachments ...string) error
	SaveStatus(id string, status CaseStatus, opts *CaseRepoSaveStatusOptions) error
	Find(id string) (*Case, error)
	FindAllBySuite(suiteId string, num *int64) ([]Case, error)
}

type buntCaseRepo struct {
	*buntRepo
}

func (r *buntRepo) newCaseRepo() (*buntCaseRepo, error) {
	err := r.db.ReplaceIndex("cases_suite", "cases:*",
		buntdb.IndexJSON("suite"))
	if err != nil {
		return nil, err
	}
	return &buntCaseRepo{r}, nil
}

func (r *buntCaseRepo) Save(c Case) (string, error) {
	var id string
	err := r.db.Update(func(tx *buntdb.Tx) error {
		id = primitive.NewObjectID().Hex()
		c.Id = id
		b, err := json.Marshal(&c)
		if err != nil {
			return err
		}
		_, _, err = tx.Set("cases:"+id, string(b), nil)
		if err != nil {
			return err
		}
		r.changes <- Change{
			Op:      ChangeOpInsert,
			Coll:    ChangeCollCases,
			Payload: c,
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *buntCaseRepo) SaveAttachments(id string, attachments ...string) error {
	return r.db.Update(func(tx *buntdb.Tx) error {
		v, err := tx.Get("cases:" + id)
		if err != nil {
			return err
		}
		var c Case
		if err := json.Unmarshal([]byte(v), &c); err != nil {
			return err
		}
		c.Attachments = append(c.Attachments, attachments...)
		b, err := json.Marshal(&c)
		if err != nil {
			return err
		}
		_, _, err = tx.Set("cases:"+id, string(b), nil)
		if err != nil {
			return err
		}
		r.changes <- Change{
			Op:      ChangeOpUpdate,
			Coll:    ChangeCollCases,
			Payload: c,
		}
		return nil
	})
}

func (r *buntCaseRepo) SaveStatus(id string, status CaseStatus, opts *CaseRepoSaveStatusOptions) error {
	return r.db.Update(func(tx *buntdb.Tx) error {
		v, err := tx.Get("cases:" + id)
		if err != nil {
			return err
		}
		var c Case
		if err := json.Unmarshal([]byte(v), &c); err != nil {
			return err
		}
		c.Status = status
		if opts != nil {
			if opts.flaky != nil {
				c.Flaky = *opts.flaky
			}
			if opts.startedAt != nil {
				c.StartedAt = *opts.startedAt
			}
			if opts.finishedAt != nil {
				c.FinishedAt = *opts.finishedAt
			}
		}
		b, err := json.Marshal(&c)
		if err != nil {
			return err
		}
		_, _, err = tx.Set("cases:"+id, string(b), nil)
		r.changes <- Change{
			Op:      ChangeOpUpdate,
			Coll:    ChangeCollCases,
			Payload: c,
		}
		return err
	})
}

func (r *buntCaseRepo) Find(id string) (*Case, error) {
	var c Case
	err := r.db.View(func(tx *buntdb.Tx) error {
		v, err := tx.Get("cases:" + id)
		if err != nil {
			return err
		}
		return json.Unmarshal([]byte(v), &c)
	})
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *buntCaseRepo) FindAllBySuite(suiteId string, num *int64) ([]Case, error) {
	values := make([]string, 0)
	err := r.db.View(func(tx *buntdb.Tx) error {
		return tx.AscendEqual("cases_suite", `{"suite":"`+suiteId+`"}`, func(k, v string) bool {
			values = append(values, v)
			return true
		})
	})
	if err != nil {
		return nil, err
	}
	cases := make([]Case, 0, len(values))
	for _, v := range values {
		var c Case
		if err := json.Unmarshal([]byte(v), &c); err != nil {
			return nil, err
		}
		if num == nil || c.Num == *num {
			cases = append(cases, c)
		}
	}
	return cases, nil
}
