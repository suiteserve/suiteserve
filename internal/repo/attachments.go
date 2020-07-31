package repo

import (
	"fmt"
	"github.com/tidwall/buntdb"
)

type Attachment struct {
	Entity
	VersionedEntity
	SoftDeleteEntity
	SuiteId     string `json:"suite_id"`
	CaseId      string `json:"case_id"`
	Filename    string `json:"filename"`
	Url         string `json:"url"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
	Timestamp   int64  `json:"timestamp"`
}

func (r *Repo) InsertAttachment(a Attachment) (id string, err error) {
	return r.insert(AttachmentColl, &a)
}

func (r *Repo) Attachment(id string) (Attachment, error) {
	var a Attachment
	err := r.getById(AttachmentColl, id, &a)
	return a, err
}

func (r *Repo) SuiteAttachments(suiteId string) ([]Attachment, error) {
	return r.attachmentsByOwner(suiteId, "")
}

func (r *Repo) CaseAttachments(caseId string) ([]Attachment, error) {
	return r.attachmentsByOwner("", caseId)
}

func (r *Repo) attachmentsByOwner(suiteId, caseId string) ([]Attachment, error) {
	var vals []string
	pivot := fmt.Sprintf(`{"suite_id":%q,"case_id":%q}`, suiteId, caseId)
	err := r.db.View(func(tx *buntdb.Tx) error {
		return tx.DescendEqual(attachmentIndexOwner, pivot, func(k, v string) bool {
			vals = append(vals, v)
			return true
		})
	})
	if err != nil {
		return nil, err
	}
	all := make([]Attachment, len(vals))
	unmarshalJsonVals(vals, func(i int) interface{} {
		return &all[i]
	})
	return all, nil
}
