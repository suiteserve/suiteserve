package repo

type LogLevelType string

const (
	LogLevelTypeTrace LogLevelType = "trace"
	LogLevelTypeDebug LogLevelType = "debug"
	LogLevelTypeInfo  LogLevelType = "info"
	LogLevelTypeWarn  LogLevelType = "warn"
	LogLevelTypeError LogLevelType = "error"
)

type LogLine struct {
	Entity    `storm:"inline"`
	CaseId    string           `json:"case_id"`
	Idx       int64        `json:"idx"`
	Level     LogLevelType `json:"level,omitempty"`
	Trace     string       `json:"trace,omitempty"`
	Message   string       `json:"message,omitempty"`
	Timestamp int64        `json:"timestamp,omitempty"`
}

func (r *Repo) InsertLogLine(ll LogLine) (string, error) {
	// err := r.db.Save(&ll)
	// return ll.Id, err
	return "", nil
}

func (r *Repo) LogLine(id string) (ll LogLine, err error) {
	// err = wrapNotFoundErr(r.db.One("Id", id, &ll))
	return
}
