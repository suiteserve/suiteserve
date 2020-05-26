package repo

import "context"

type LogLevelType string

const (
	LogLevelTypeTrace LogLevelType = "trace"
	LogLevelTypeDebug LogLevelType = "debug"
	LogLevelTypeInfo  LogLevelType = "info"
	LogLevelTypeWarn  LogLevelType = "warn"
	LogLevelTypeError LogLevelType = "error"
)

type UnsavedLogEntry struct {
	Case      string       `json:"case"`
	Index     int64        `json:"index"`
	Level     LogLevelType `json:"level"`
	Trace     string       `json:"trace,omitempty" bson:",omitempty"`
	Message   string       `json:"message,omitempty" bson:",omitempty"`
	Timestamp int64        `json:"timestamp"`
}

type LogEntry struct {
	SavedEntity     `bson:",inline"`
	UnsavedLogEntry `bson:",inline"`
}

type LogRepo interface {
	Save(ctx context.Context, e UnsavedLogEntry) (string, error)
	Find(ctx context.Context, id string) (*LogEntry, error)
	FindAllByCase(ctx context.Context, caseId string) ([]LogEntry, error)
}
