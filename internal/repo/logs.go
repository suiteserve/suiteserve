package repo

import (
	"context"
)

type LogLevelType string

const (
	LogLevelTypeTrace LogLevelType = "trace"
	LogLevelTypeDebug LogLevelType = "debug"
	LogLevelTypeInfo  LogLevelType = "info"
	LogLevelTypeWarn  LogLevelType = "warn"
	LogLevelTypeError LogLevelType = "error"
)

type LogLine struct {
	Entity    `bson:",inline"`
	CaseId    Id           `json:"case_id" bson:"case_id"`
	Idx       int64        `json:"idx"`
	Level     LogLevelType `json:"level"`
	Trace     string       `json:"trace,omitempty" bson:",omitempty"`
	Message   string       `json:"message,omitempty" bson:",omitempty"`
	Timestamp int64        `json:"timestamp"`
}

func (r *Repo) InsertLogLine(ctx context.Context, ll LogLine) (Id, error) {
	return r.insert(ctx, "logs", ll)
}

func (r *Repo) LogLine(ctx context.Context, id Id) (interface{}, error) {
	var ll LogLine
	if err := r.findById(ctx, "logs", id, &ll); err != nil {
		return nil, err
	}
	return ll, nil
}
