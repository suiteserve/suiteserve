package repo

import "encoding/json"

type CaseStatus string

const (
	CaseStatusCreated  CaseStatus = "created"
	CaseStatusStarted  CaseStatus = "started"
	CaseStatusFinished CaseStatus = "finished"
)

type CaseResult string

const (
	CaseResultPassed  CaseResult = "passed"
	CaseResultFailed  CaseResult = "failed"
	CaseResultSkipped CaseResult = "skipped"
	CaseResultAborted CaseResult = "aborted"
	CaseResultErrored CaseResult = "errored"
)

type Case struct {
	Entity          `storm:"inline"`
	VersionedEntity `storm:"inline"`
	SuiteId         string                         `json:"suite_id"`
	Name            string                     `json:"name,omitempty"`
	Description     string                     `json:"description,omitempty"`
	Tags            []string                   `json:"tags,omitempty"`
	Idx             int64                      `json:"idx"`
	Args            map[string]json.RawMessage `json:"args,omitempty"`
	Status          CaseStatus                 `json:"status,omitempty"`
	Result          CaseResult                 `json:"result,omitempty"`
	CreatedAt       int64                      `json:"created_at,omitempty"`
	StartedAt       int64                      `json:"started_at,omitempty"`
	FinishedAt      int64                      `json:"finished_at,omitempty"`
}

func (r *Repo) InsertCase(c Case) (string, error) {
	// err := r.db.Save(&c)
	// return c.Id, err
	// return c.Id, nil
	return "", nil
}

func (r *Repo) Case(id string) (c Case, err error) {
	// err = wrapNotFoundErr(r.db.One("Id", id, &c))
	return
}
