package repo

import (
	"context"
	"github.com/tidwall/buntdb"
	"io"
)

type buntAttachmentRepo struct {
	*buntRepo
	files *fileRepo
}

func (r *buntRepo) newAttachmentRepo(files *fileRepo) (*buntAttachmentRepo, error) {
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
	return &buntAttachmentRepo{r, files}, nil
}

func (r *buntAttachmentRepo) Save(_ context.Context, unsavedA UnsavedAttachmentInfo, src io.Reader) (AttachmentFile, error) {
	a := AttachmentInfo{
		UnsavedAttachmentInfo: unsavedA,
	}
	var file *fileAccessor
	_, err := r.funcSave(AttachmentColl, &a, func(id string) error {
		a.Id = id
		file = r.files.newFileAccessor(id)
		var err error
		a.Size, err = file.save(src)
		return err
	})
	if err != nil {
		file.delete()
		return nil, err
	}
	return r.files.newAttachmentFile(&a), nil
}

func (r *buntAttachmentRepo) Find(_ context.Context, id string) (AttachmentFile, error) {
	var info AttachmentInfo
	if err := r.find(AttachmentColl, id, &info); err != nil {
		return nil, err
	}
	return r.files.newAttachmentFile(&info), nil
}

func (r *buntAttachmentRepo) FindAll(_ context.Context, includeDeleted bool) ([]AttachmentFile, error) {
	var infos []AttachmentInfo
	index := "attachments_deleted"
	if includeDeleted {
		index = "attachments_id"
	}
	if err := r.findAll(index, includeDeleted, &infos); err != nil {
		return nil, err
	}
	files := make([]AttachmentFile, len(infos))
	for i := range infos {
		files[i] = r.files.newAttachmentFile(&infos[i])
	}
	return files, nil
}

func (r *buntAttachmentRepo) Delete(_ context.Context, id string, at int64) error {
	if err := r.delete(AttachmentColl, id, at); err != nil {
		return err
	}
	r.files.newFileAccessor(id).delete()
	return nil
}

func (r *buntAttachmentRepo) DeleteAll(_ context.Context, at int64) error {
	ids, err := r.deleteAll(AttachmentColl, "attachments_deleted", at)
	if err != nil {
		return err
	}
	for _, id := range ids {
		r.files.newFileAccessor(id).delete()
	}
	return nil
}
