package repo

import "context"

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

type UnsavedCase struct {
	Suite       string     `json:"suite"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty" bson:",omitempty"`
	Tags        []string   `json:"tags,omitempty" bson:",omitempty"`
	Num         int64      `json:"num"`
	Links       []CaseLink `json:"links,omitempty" bson:",omitempty"`
	Args        []CaseArg  `json:"args,omitempty" bson:",omitempty"`
	Attachments []string   `json:"attachments,omitempty" bson:",omitempty"`
	Status      CaseStatus `json:"status"`
	CreatedAt   int64      `json:"created_at" bson:"created_at"`
	StartedAt   int64      `json:"started_at,omitempty" bson:"started_at,omitempty"`
	FinishedAt  int64      `json:"finished_at,omitempty" bson:"finished_at,omitempty"`
}

func (c UnsavedCase) Finished() bool {
	return c.Status != CaseStatusCreated && c.Status != CaseStatusRunning
}

type Case struct {
	SavedEntity     `bson:",inline"`
	VersionedEntity `bson:",inline"`
	UnsavedCase     `bson:",inline"`
}

type CaseRepoSaveStatusOptions struct {
	startedAt  *int64
	finishedAt *int64
}

func NewCaseRepoSaveStatusOptions() *CaseRepoSaveStatusOptions {
	return &CaseRepoSaveStatusOptions{}
}

func (o *CaseRepoSaveStatusOptions) StartedAt(startedAt int64) *CaseRepoSaveStatusOptions {
	o.startedAt = &startedAt
	return o
}

func (o *CaseRepoSaveStatusOptions) FinishedAt(finishedAt int64) *CaseRepoSaveStatusOptions {
	o.finishedAt = &finishedAt
	return o
}

type CaseRepo interface {
	Save(ctx context.Context, c UnsavedCase) (string, error)
	SaveAttachment(ctx context.Context, id string, attachmentId string) error
	SaveStatus(ctx context.Context, id string, status CaseStatus, opts *CaseRepoSaveStatusOptions) error
	Find(ctx context.Context, id string) (*Case, error)
	FindAllBySuite(ctx context.Context, suiteId string, num *int64) ([]Case, error)
}
