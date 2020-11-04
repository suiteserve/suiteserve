package repo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

type Attachment struct {
	Entity           `bson:",inline"`
	VersionedEntity  `bson:",inline"`
	SoftDeleteEntity `bson:",inline"`
	SuiteId          Id      `json:"suiteId,omitempty" bson:"suite_id"`
	CaseId           Id      `json:"caseId,omitempty" bson:"case_id"`
	Filename         *string `json:"filename"`
	ContentType      *string `json:"contentType" bson:"content_type"`
	Size             *int64  `json:"size"`
	Timestamp        *Time   `json:"timestamp"`
}

var attachmentType = reflect.TypeOf(Attachment{})

func (r *Repo) InsertAttachment(ctx context.Context, a Attachment) (Id, error) {
	return r.insert(ctx, attachmentsColl, a)
}

func (r *Repo) Attachment(ctx context.Context, id Id) (interface{}, error) {
	return r.findById(ctx, attachmentsColl, id, Attachment{})
}

func (r *Repo) AllAttachments(ctx context.Context) (interface{}, error) {
	return readAll(ctx, []Attachment{}, func() (*mongo.Cursor, error) {
		return r.db.Collection(attachmentsColl).Find(ctx, bson.D{
			{"deleted", false},
		})
	})
}

func (r *Repo) SuiteAttachments(ctx context.Context,
	suiteId Id) (interface{}, error) {
	return readAll(ctx, []Attachment{}, func() (*mongo.Cursor, error) {
		return r.db.Collection(attachmentsColl).Find(ctx, bson.D{
			{"deleted", false},
			{"suite_id", bsonId{suiteId}},
		})
	})
}

func (r *Repo) CaseAttachments(ctx context.Context,
	caseId Id) (interface{}, error) {
	return readAll(ctx, []Attachment{}, func() (*mongo.Cursor, error) {
		return r.db.Collection(attachmentsColl).Find(ctx, bson.D{
			{"deleted", false},
			{"suite_id", nil},
			{"case_id", bsonId{caseId}},
		})
	})
}
