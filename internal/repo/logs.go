package repo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type LogLine struct {
	Entity `bson:",inline"`
	CaseId Id     `json:"caseId" bson:"case_id"`
	Idx    int64  `json:"idx"`
	Error  bool   `json:"error,omitempty" bson:",omitempty"`
	Line   string `json:"line,omitempty" bson:",omitempty"`
}

func (r *Repo) InsertLogLine(ctx context.Context, ll LogLine) (Id, error) {
	return r.insert(ctx, "logs", ll)
}

func (r *Repo) LogLine(ctx context.Context, id Id) (interface{}, error) {
	return r.findById(ctx, "logs", id, LogLine{})
}

func (r *Repo) CaseLogLines(ctx context.Context, caseId Id) (interface{}, error) {
	return readAll(ctx, []LogLine{}, func() (*mongo.Cursor, error) {
		return r.db.Collection("logs").Find(ctx, bson.D{
			{"case_id", bsonId{caseId}},
		})
	})
}
