package repo

import "github.com/tidwall/buntdb"

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
	*Entity     `bson:",inline"`
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
	err = r.db.ReplaceIndex("cases_suite_num", "cases:*",
		buntdb.IndexJSON("suite"), buntdb.IndexJSON("num"))
	if err != nil {
		return nil, err
	}
	err = r.db.ReplaceIndex("cases_deleted", "cases:*",
		buntdb.IndexJSON("deleted"))
	if err != nil {
		return nil, err
	}
	return &buntCaseRepo{r}, nil
}

func (r *buntCaseRepo) Save(c Case) (string, error) {
	return r.save(&c, CaseCollection)
}

func (r *buntCaseRepo) SaveAttachments(id string, attachments ...string) error {
	m := make(map[string]interface{})
	for _, a := range attachments {
		m["attachments.-1"] = a
	}
	return r.set(CaseCollection, id, m)
}

func (r *buntCaseRepo) SaveStatus(id string, status CaseStatus, opts *CaseRepoSaveStatusOptions) error {
	return r.set(CaseCollection, id, map[string]interface{}{
		"status":      status,
		"flaky":       opts.flaky,
		"started_at":  opts.startedAt,
		"finished_at": opts.finishedAt,
	})
}

func (r *buntCaseRepo) Find(id string) (*Case, error) {
	var c Case
	if err := r.find(CaseCollection, id, Case{}); err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *buntCaseRepo) FindAllBySuite(suiteId string, num *int64) ([]Case, error) {
	m := map[string]interface{}{
		"suite": suiteId,
	}
	index := "cases_suite"
	if num != nil {
		m["num"] = *num
		index = "cases_suite_num"
	}
	var cases []Case
	if err := r.findAllBy(index, m, &cases); err != nil {
		return nil, err
	}
	return cases, nil
}
