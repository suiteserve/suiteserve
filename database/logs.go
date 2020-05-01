package database

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	return iToTime(c.Timestamp)
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

func (d *WithContext) LogMessage(logId string) (*LogMessage, error) {
	logOid, err := primitive.ObjectIDFromHex(logId)
	if err != nil {
		return nil, fmt.Errorf("%w: parse object id", ErrNotFound)
	}

	ctx, cancel := d.newContext()
	defer cancel()
	res := d.logs.FindOne(ctx, bson.M{
		"_id": logOid,
	})
	var logMsg LogMessage
	if err := res.Decode(&logMsg); err == mongo.ErrNoDocuments {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("find log message: %v", err)
	}
	return &logMsg, nil
}

func (d *WithContext) AllLogMessages(caseId string) ([]LogMessage, error) {
	ctx, cancel := d.newContext()
	defer cancel()
	cursor, err := d.logs.Find(ctx, bson.M{
		"case": caseId,
	}, options.Find().SetSort(bson.D{
		{"timestamp", 1},
		{"_id", 1},
	}))
	if err != nil {
		return nil, fmt.Errorf("find all log messages for case: %v", err)
	}

	logMsgs := make([]LogMessage, 0)
	if err := cursor.All(ctx, &logMsgs); err != nil {
		return nil, fmt.Errorf("decode all log messages for case: %v", err)
	}
	return logMsgs, nil
}
