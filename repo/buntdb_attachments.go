package repo

import (
	"context"
	"io"
)

func (d *BuntDb) InsertAttachment(_ context.Context, a *UnsavedAttachment, src io.Reader) (AttachmentFile, error) {
	savedAttachment := Attachment{
		UnsavedAttachment: *a,
	}
	var file *fileAccessor
	_, err := d.funcInsert(CollAttachments, &a, func(id string) error {
		savedAttachment.Id = id
		file = d.files.newFileAccessor(id)
		var err error
		savedAttachment.Size, err = file.save(src)
		return err
	})
	if err != nil {
		file.delete()
		return nil, err
	}
	return d.files.newAttachmentFile(&savedAttachment), nil
}

func (d *BuntDb) Attachment(_ context.Context, id string) (AttachmentFile, error) {
	var a Attachment
	if err := d.find(CollAttachments, id, &a); err != nil {
		return nil, err
	}
	return d.files.newAttachmentFile(&a), nil
}

func (d *BuntDb) AllAttachments(_ context.Context) ([]AttachmentFile, error) {
	var attachments []Attachment
	if err := d.findAll("attachments_deleted", &attachments); err != nil {
		return nil, err
	}
	files := make([]AttachmentFile, len(attachments))
	for i, a := range attachments {
		files[i] = d.files.newAttachmentFile(&a)
	}
	return files, nil
}

func (d *BuntDb) DeleteAttachment(_ context.Context, id string, at int64) error {
	return d.delete(CollAttachments, id, at)
}

func (d *BuntDb) DeleteAllAttachments(_ context.Context, at int64) error {
	return d.deleteAll(CollAttachments, "attachments_deleted", at)
}
