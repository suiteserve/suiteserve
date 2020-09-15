package repo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Attachment struct {
	Entity           `bson:",inline"`
	VersionedEntity  `bson:",inline"`
	SoftDeleteEntity `bson:",inline"`
	SuiteId          Id     `json:"suite_id,omitempty" bson:"suite_id,omitempty"`
	CaseId           Id     `json:"case_id,omitempty" bson:"case_id,omitempty"`
	Filename         string `json:"filename"`
	ContentType      string `json:"content_type" bson:"content_type"`
	Size             int64  `json:"size"`
	Timestamp        int64  `json:"timestamp"`
}

func (r *Repo) InsertAttachment(ctx context.Context, a Attachment) (Id, error) {
	return r.insert(ctx, "attachments", a)
}

func (r *Repo) Attachment(ctx context.Context, id Id) (*Attachment, error) {
	var a Attachment
	if err := r.findById(ctx, "attachments", id, &a); err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *Repo) Attachments(ctx context.Context) ([]Attachment, error) {
	opts := options.Find().SetSort(bson.D{
		{"suite_id", 1},
		{"case_id", 1},
		{"timestamp", 1},
		{"_id", 1},
	})
	res, err := r.db.Collection("attachments").Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, err
	}
	a := make([]Attachment, 0)
	if err := res.All(ctx, &a); err != nil {
		return nil, err
	}
	return a, nil
}

func (r *Repo) SuiteAttachments(ctx context.Context,
	suiteId Id) ([]Attachment, error) {
	filter := bson.D{{"suite_id", suiteId}}
	opts := options.Find().SetSort(bson.D{
		{"case_id", 1},
		{"timestamp", 1},
		{"_id", 1},
	})
	res, err := r.db.Collection("attachments").Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	a := make([]Attachment, 0)
	if err := res.All(ctx, &a); err != nil {
		return nil, err
	}
	return a, nil
}

func (r *Repo) CaseAttachments(ctx context.Context,
	caseId Id) ([]Attachment, error) {
	filter := bson.D{{"suite_id", nil}, {"case_id", caseId}}
	opts := options.Find().SetSort(bson.D{
		{"timestamp", 1},
		{"_id", 1},
	})
	res, err := r.db.Collection("attachments").Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	a := make([]Attachment, 0)
	if err := res.All(ctx, &a); err != nil {
		return nil, err
	}
	return a, nil
}
