package database

import (
	"log"
	"time"
)

type LogLevelType string

const (
	LogLevelTypeTrace LogLevelType = "trace"
	LogLevelTypeDebug              = "debug"
	LogLevelTypeInfo               = "info"
	LogLevelTypeWarn               = "warn"
	LogLevelTypeError              = "error"
)

type NewLogMessage struct {
	Level       LogLevelType `json:"level" validate:"oneof=trace debug info warn error"`
	Trace       string       `json:"trace,omitempty" bson:",omitempty"`
	Message     string       `json:"message,omitempty" bson:",omitempty"`
	Attachments []string     `json:"attachments,omitempty" bson:",omitempty" validate:"dive,required"`
	Timestamp   int64        `json:"timestamp" validate:"gte=0"`
}

func (c *NewLogMessage) TimestampTime() time.Time {
	return time.Unix(c.Timestamp, 0)
}

type LogMessage struct {
	Id            interface{} `json:"id" bson:"_id,omitempty"`
	Case          string      `json:"case"`
	NewLogMessage `bson:",inline"`
}

func (d *WithContext) NewLogMessage(caseId string, l NewLogMessage) (string, error) {
	if err := validate.Struct(&l); err != nil {
		log.Printf("validate log message: %v\n", err)
		return "", ErrInvalidModel
	}

	return d.insert(d.logs, LogMessage{
		Case:          caseId,
		NewLogMessage: l,
	})
}
