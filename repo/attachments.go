package repo

type Attachment struct {
	*SoftDeleteEntity `bson:",inline"`
	Filename          string `json:"filename"`
	Size              int64  `json:"size"`
	ContentType       string `json:"content_type" bson:"content_type"`
}

type AttachmentRepo interface {
	Save(Attachment) (string, error)
	Find(id string) (*Attachment, error)
	FindAll(includeDeleted bool) ([]Attachment, error)
	Delete(id string) error
	DeleteAll() error
}
