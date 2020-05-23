package repo

import "context"

type Attachment struct {
	SoftDeleteEntity `bson:",inline"`
	Filename         string `json:"filename"`
	Size             int64  `json:"size"`
	ContentType      string `json:"content_type" bson:"content_type"`
}

type AttachmentRepo interface {
	Save(ctx context.Context, a Attachment) (string, error)
	Find(ctx context.Context, id string) (*Attachment, error)
	FindAll(ctx context.Context, includeDeleted bool) ([]Attachment, error)
	Delete(ctx context.Context, id string, at int64) error
	DeleteAll(ctx context.Context, at int64) error
}
