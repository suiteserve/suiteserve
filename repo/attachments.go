package repo

import (
	"context"
	"io"
)

type UnsavedAttachmentInfo struct {
	SoftDeleteEntity `bson:",inline"`
	Filename         string `json:"filename"`
	ContentType      string `json:"content_type" bson:"content_type"`
}

type AttachmentInfo struct {
	SavedEntity           `bson:",inline"`
	VersionedEntity       `bson:",inline"`
	UnsavedAttachmentInfo `bson:",inline"`
	Size                  int64 `json:"size"`
}

type AttachmentFile interface {
	Info() *AttachmentInfo
	Open() (io.ReadCloser, error)
}

type AttachmentRepo interface {
	Save(ctx context.Context, a UnsavedAttachmentInfo, src io.Reader) (AttachmentFile, error)
	Find(ctx context.Context, id string) (AttachmentFile, error)
	FindAll(ctx context.Context, includeDeleted bool) ([]AttachmentFile, error)
	Delete(ctx context.Context, id string, at int64) error
	DeleteAll(ctx context.Context, at int64) error
}
