package repo

import (
	"github.com/tidwall/buntdb"
)

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

type buntAttachmentRepo struct {
	*buntRepo
}

func (r *buntRepo) newAttachmentRepo() (*buntAttachmentRepo, error) {
	err := r.db.ReplaceIndex("attachments_deleted", "attachments:*",
		buntdb.IndexJSON("deleted"))
	if err != nil {
		return nil, err
	}
	return &buntAttachmentRepo{r}, nil
}

func (r *buntAttachmentRepo) Save(a Attachment) (string, error) {
	return r.save(&a, AttachmentCollection)
}

func (r *buntAttachmentRepo) Find(id string) (*Attachment, error) {
	var a Attachment
	if err := r.find(AttachmentCollection, id, &a); err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *buntAttachmentRepo) FindAll(includeDeleted bool) ([]Attachment, error) {
	var attachments []Attachment
	if err := r.findAll("attachments_deleted", includeDeleted, &attachments); err != nil {
		return nil, err
	}
	return attachments, nil
}

func (r *buntAttachmentRepo) Delete(id string) error {
	return r.delete(AttachmentCollection, id)
}

func (r *buntAttachmentRepo) DeleteAll() error {
	return r.deleteAll(AttachmentCollection, "attachments_deleted")
}
