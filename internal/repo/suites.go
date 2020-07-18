package repo

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
	StartedAt      int64       `json:"started_at,omitempty"`
	FinishedAt     int64       `json:"finished_at,omitempty"`
}

func (r *Repo) InsertSuite(s Suite) (id string, err error) {
	return r.insert(SuiteColl, &s)
}

func (r *Repo) Suite(id string) (*Suite, error) {
	var s Suite
	return &s, r.getById(SuiteColl, id, &s)
}
