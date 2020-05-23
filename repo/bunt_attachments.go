package repo

import "github.com/tidwall/buntdb"

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
		buntdb.IndexJSON("deleted"), IndexOptionalJSON("id"))
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
	index := "attachments_deleted"
	if includeDeleted {
		index = "attachments_id"
	}
	if err := r.findAll(index, includeDeleted, &attachments); err != nil {
		return nil, err
	}
	return attachments, nil
}

func (r *buntAttachmentRepo) Delete(id string, at int64) error {
	return r.delete(AttachmentCollection, id, at)
}

func (r *buntAttachmentRepo) DeleteAll(at int64) error {
	return r.deleteAll(AttachmentCollection, "attachments_deleted", at)
}
