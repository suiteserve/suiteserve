package repo

type LogLevelType string

const (
	LogLevelTypeUnknown LogLevelType = "unknown"
	LogLevelTypeTrace   LogLevelType = "trace"
	LogLevelTypeDebug   LogLevelType = "debug"
	LogLevelTypeInfo    LogLevelType = "info"
	LogLevelTypeWarn    LogLevelType = "warn"
	LogLevelTypeError   LogLevelType = "error"
)

type LogLine struct {
	Entity
	CaseId    string       `json:"case_id"`
	Idx       int64        `json:"idx"`
	Level     LogLevelType `json:"level"`
	Trace     string       `json:"trace,omitempty"`
	Message   string       `json:"message,omitempty"`
	Timestamp int64        `json:"timestamp"`
}

func (r *Repo) InsertLogLine(l LogLine) (id string, err error) {
	return r.insert(logColl, &l)
}

func (r *Repo) LogLine(id string) (*LogLine, error) {
	var l LogLine
	return &l, r.getById(logColl, id, &l)
}
