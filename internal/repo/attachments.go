package repo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

type Attachment struct {
	Entity          `bson:",inline"`
	VersionedEntity `bson:",inline"`
	SuiteId         *Id     `json:"suiteId,omitempty" bson:"suite_id"`
	CaseId          *Id     `json:"caseId,omitempty" bson:"case_id"`
	Filename        *string `json:"filename,omitempty"`
	ContentType     *string `json:"contentType,omitempty" bson:"content_type"`
	Size            *int64  `json:"size,omitempty"`
	Timestamp       *MsTime `json:"timestamp,omitempty"`
}

var attachmentType = reflect.TypeOf(Attachment{})

func (r *Repo) InsertAttachment(ctx context.Context, a Attachment) (Id, error) {
	return r.insert(ctx, Attachments, a)
}

func (r *Repo) Attachment(ctx context.Context, id Id) (Attachment, error) {
	var a Attachment
	err := r.findById(ctx, Attachments, id, &a)
	return a, err
}

func (r *Repo) AllAttachments(ctx context.Context) ([]Attachment, error) {
	as := []Attachment{}
	return as, readAll(ctx, &as, func() (*mongo.Cursor, error) {
		return r.db.Collection(attachments).Find(ctx, bson.D{})
	})
}

func (r *Repo) SuiteAttachments(ctx context.Context,
	suiteId Id) ([]Attachment, error) {
	as := []Attachment{}
	return as, readAll(ctx, &as, func() (*mongo.Cursor, error) {
		return r.db.Collection(attachments).Find(ctx, bson.D{
			{"suite_id", suiteId},
		})
	})
}

func (r *Repo) CaseAttachments(ctx context.Context,
	caseId Id) ([]Attachment, error) {
	as := []Attachment{}
	return as, readAll(ctx, &as, func() (*mongo.Cursor, error) {
		return r.db.Collection(attachments).Find(ctx, bson.D{
			{"suite_id", nil},
			{"case_id", caseId},
		})
	})
}
