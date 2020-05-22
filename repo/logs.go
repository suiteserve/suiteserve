package repo

type LogLevelType string

const (
	LogLevelTypeTrace LogLevelType = "trace"
	LogLevelTypeDebug LogLevelType = "debug"
	LogLevelTypeInfo  LogLevelType = "info"
	LogLevelTypeWarn  LogLevelType = "warn"
	LogLevelTypeError LogLevelType = "error"
)

type LogEntry struct {
	*Entity   `bson:",inline"`
	Case      string       `json:"case"`
	Seq       int64        `json:"seq"`
	Level     LogLevelType `json:"level"`
	Trace     string       `json:"trace,omitempty" bson:",omitempty"`
	Message   string       `json:"message,omitempty" bson:",omitempty"`
	Timestamp int64        `json:"timestamp"`
}

type LogRepo interface {
	Save(LogEntry) (string, error)
	Find(id string) (*LogEntry, error)
	FindAllByCase(caseId string) ([]LogEntry, error)
}
