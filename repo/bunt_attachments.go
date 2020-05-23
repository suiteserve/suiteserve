package repo

import (
	"context"
	"github.com/tidwall/buntdb"
)

type buntAttachmentRepo struct {
	*buntRepo
}

func (r *buntRepo) newAttachmentRepo() (*buntAttachmentRepo, error) {
	err := r.db.ReplaceIndex("attachments_id", "attachments:*",
		buntdb.IndexJSON("id"))
	if err != nil {
		return nil, err
	}
	err = r.db.ReplaceIndex("attachments_deleted", "attachments:*",
		buntdb.IndexJSON("deleted"), indexJSONOptional("id"))
	if err != nil {
		return nil, err
	}
	return &buntAttachmentRepo{r}, nil
}

func (r *buntAttachmentRepo) Save(_ context.Context, a Attachment) (string, error) {
	return r.save(&a, AttachmentCollection)
}

func (r *buntAttachmentRepo) Find(_ context.Context, id string) (*Attachment, error) {
	var a Attachment
	if err := r.find(AttachmentCollection, id, &a); err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *buntAttachmentRepo) FindAll(_ context.Context, includeDeleted bool) ([]Attachment, error) {
	var attachments []Attachment
	index := "attachments_deleted"
	if includeDeleted {
		index = "attachments_id"
	}
	if err := r.findAll(index, includeDeleted, &attachments); err != nil {
		return nil, err
	}
	return attachments, nil
}

func (r *buntAttachmentRepo) Delete(_ context.Context, id string, at int64) error {
	return r.delete(AttachmentCollection, id, at)
}

func (r *buntAttachmentRepo) DeleteAll(_ context.Context, at int64) error {
	return r.deleteAll(AttachmentCollection, "attachments_deleted", at)
}
