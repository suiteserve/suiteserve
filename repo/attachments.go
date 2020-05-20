package repo

import (
	"encoding/json"
	"github.com/tidwall/buntdb"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Attachment struct {
	Id          string `json:"id" bson:"_id,omitempty"`
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type" bson:"content_type"`
	Deleted     bool   `json:"deleted"`
	DeletedAt   int64  `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
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
	b, err := json.Marshal(&a)
	if err != nil {
		return "", err
	}

	var id string
	err = r.db.Update(func(tx *buntdb.Tx) error {
		id = primitive.NewObjectID().Hex()
		a.Id = id
		_, _, err = tx.Set("attachments:"+id, string(b), nil)
		if err != nil {
			return err
		}
		r.changes <- Change{
			Op:      ChangeOpInsert,
			Coll:    ChangeCollAttachments,
			Payload: a,
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *buntAttachmentRepo) Find(id string) (*Attachment, error) {
	var a Attachment
	err := r.db.View(func(tx *buntdb.Tx) error {
		v, err := tx.Get("attachments:" + id)
		if err != nil {
			return err
		}
		return json.Unmarshal([]byte(v), &a)
	})
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *buntAttachmentRepo) FindAll(includeDeleted bool) ([]Attachment, error) {
	values := make([]string, 0)
	err := r.db.View(func(tx *buntdb.Tx) error {
		iterator := func(k, v string) bool {
			values = append(values, v)
			return true
		}
		if includeDeleted {
			return tx.Ascend("", iterator)
		}
		return tx.AscendEqual("attachments_deleted", "false", iterator)
	})
	if err != nil {
		return nil, err
	}
	attachments := make([]Attachment, len(values))
	for i, v := range values {
		var a Attachment
		if err := json.Unmarshal([]byte(v), &a); err != nil {
			return nil, err
		}
		attachments[i] = a
	}
	return attachments, nil
}

func (r *buntAttachmentRepo) Delete(id string) error {
	return r.db.Update(func(tx *buntdb.Tx) error {
		v, err := tx.Get("attachments:" + id)
		if err != nil {
			return err
		}
		var a Attachment
		if err := json.Unmarshal([]byte(v), &a); err != nil {
			return err
		}
		a.Deleted = true
		a.DeletedAt = nowTimeMillis()
		b, err := json.Marshal(&a)
		if err != nil {
			return err
		}
		_, _, err = tx.Set("attachments:"+id, string(b), nil)
		if err != nil {
			return err
		}
		r.changes <- Change{
			Op:      ChangeOpUpdate,
			Coll:    ChangeCollAttachments,
			Payload: a,
		}
		return nil
	})
}

func (r *buntAttachmentRepo) DeleteAll() error {
	return r.db.Update(func(tx *buntdb.Tx) error {
		values := make([]string, 0)
		err := tx.AscendEqual("attachments_deleted", "false", func(k, v string) bool {
			values = append(values, v)
			return true
		})
		if err != nil {
			return err
		}
		deletedAt := nowTimeMillis()
		for _, v := range values {
			var a Attachment
			if err := json.Unmarshal([]byte(v), &a); err != nil {
				return err
			}
			a.Deleted = true
			a.DeletedAt = deletedAt
			b, err := json.Marshal(&a)
			if err != nil {
				return err
			}
			if _, _, err := tx.Set("attachments:"+a.Id, string(b), nil); err != nil {
				return err
			}
			r.changes <- Change{
				Op:      ChangeOpUpdate,
				Coll:    ChangeCollAttachments,
				Payload: a,
			}
		}
		return nil
	})
}
