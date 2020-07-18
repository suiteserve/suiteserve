package repo

import "encoding/json"

type CaseStatus string

const (
	CaseStatusUnknown  CaseStatus = "unknown"
	CaseStatusCreated  CaseStatus = "created"
	CaseStatusStarted  CaseStatus = "started"
	CaseStatusFinished CaseStatus = "finished"
)

type CaseResult string

const (
	CaseResultUnknown CaseResult = "unknown"
	CaseResultPassed  CaseResult = "passed"
	CaseResultFailed  CaseResult = "failed"
	CaseResultSkipped CaseResult = "skipped"
	CaseResultAborted CaseResult = "aborted"
	CaseResultErrored CaseResult = "errored"
)

type Case struct {
	Entity
	VersionedEntity
	SuiteId     string                     `json:"suite_id"`
	Name        string                     `json:"name,omitempty"`
	Description string                     `json:"description,omitempty"`
	Tags        []string                   `json:"tags,omitempty"`
	Idx         int64                      `json:"idx"`
	Args        map[string]json.RawMessage `json:"args,omitempty"`
	Status      CaseStatus                 `json:"status"`
	Result      CaseResult                 `json:"result"`
	CreatedAt   int64                      `json:"created_at"`
	StartedAt   int64                      `json:"started_at,omitempty"`
	FinishedAt  int64                      `json:"finished_at,omitempty"`
}

func (r *Repo) InsertCase(c Case) (id string, err error) {
	return r.insert(CaseColl, &c)
}

func (r *Repo) Case(id string) (*Case, error) {
	var c Case
	return &c, r.getById(CaseColl, id, &c)
}
