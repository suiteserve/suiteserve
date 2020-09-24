package repo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
)

type Attachment struct {
	Entity           `bson:",inline"`
	VersionedEntity  `bson:",inline"`
	SoftDeleteEntity `bson:",inline"`
	SuiteId          Id     `json:"suite_id,omitempty" bson:"suite_id"`
	CaseId           Id     `json:"case_id,omitempty" bson:"case_id"`
	Filename         string `json:"filename"`
	ContentType      string `json:"content_type" bson:"content_type"`
	Size             int64  `json:"size"`
	Timestamp        int64  `json:"timestamp"`
}

func (r *Repo) InsertAttachment(ctx context.Context, a Attachment) (Id, error) {
	return r.insert(ctx, "attachments", a)
}

func (r *Repo) Attachment(ctx context.Context, id Id) (interface{}, error) {
	var a Attachment
	if err := r.findById(ctx, "attachments", id, &a); err != nil {
		return nil, err
	}
	return a, nil
}

func (r *Repo) AllAttachments(ctx context.Context) (interface{}, error) {
	res, err := r.db.Collection("attachments").Find(ctx, bson.D{
		{"deleted", false},
	})
	if err != nil {
		return nil, err
	}
	a := []Attachment{}
	if err := res.All(ctx, &a); err != nil {
		return nil, err
	}
	return a, nil
}

func (r *Repo) SuiteAttachments(ctx context.Context, suiteId Id) (interface{}, error) {
	res, err := r.db.Collection("attachments").Find(ctx, bson.D{
		{"deleted", false},
		{"suite_id", suiteId},
	})
	if err != nil {
		return nil, err
	}
	a := []Attachment{}
	if err := res.All(ctx, &a); err != nil {
		return nil, err
	}
	return a, nil
}

func (r *Repo) CaseAttachments(ctx context.Context, caseId Id) (interface{}, error) {
	res, err := r.db.Collection("attachments").Find(ctx, bson.D{
		{"deleted", false},
		{"suite_id", nil},
		{"case_id", caseId},
	})
	if err != nil {
		return nil, err
	}
	a := []Attachment{}
	if err := res.All(ctx, &a); err != nil {
		return nil, err
	}
	return a, nil
}
