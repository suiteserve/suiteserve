package repo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

type LogLine struct {
	Entity `bson:",inline"`
	CaseId *Id      `json:"caseId" bson:"case_id"`
	Idx    *int64  `json:"idx"`
	Error  *bool   `json:"error,omitempty" bson:",omitempty"`
	Line   *string `json:"line,omitempty" bson:",omitempty"`
}

var logLineType = reflect.TypeOf(LogLine{})

func (r *Repo) InsertLogLine(ctx context.Context, ll LogLine) (Id, error) {
	return r.insert(ctx, Logs, ll)
}

func (r *Repo) LogLine(ctx context.Context, id Id) (LogLine, error) {
	var ll LogLine
	err := r.findById(ctx, Logs, id, &ll)
	return ll, err
}

func (r *Repo) CaseLogLines(ctx context.Context,
	caseId Id) ([]LogLine, error) {
	lls := []LogLine{}
	return lls, readAll(ctx, &lls, func() (*mongo.Cursor, error) {
		return r.db.Collection(logs).Find(ctx, bson.D{
			{"case_id", caseId},
		})
	})
}
