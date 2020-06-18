package repo

import (
	"io"
)

type UnsavedAttachment struct {
	SoftDeleteEntity `bson:",inline"`
	Filename         string `json:"filename"`
	ContentType      string `json:"content_type" bson:"content_type"`
}

type Attachment struct {
	SavedEntity       `bson:",inline"`
	VersionedEntity   `bson:",inline"`
	UnsavedAttachment `bson:",inline"`
	Size              int64 `json:"size"`
}

type AttachmentFile interface {
	Info() *Attachment
	Open() (io.ReadCloser, error)
}
