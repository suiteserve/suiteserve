package repo

import (
	bolt "go.etcd.io/bbolt"
)

type Attachment struct {
	Entity
	VersionedEntity
	SoftDeleteEntity
	SuiteId     string `json:"suite_id"`
	CaseId      string `json:"case_id"`
	Filename    string `json:"filename"`
	Url         string `json:"url"`
	ContentType string `json:"content_type" bson:"content_type"`
	Size        int64  `json:"size"`
	Timestamp   int64  `json:"timestamp,omitempty" bson:",omitempty"`
}

func (r *Repo) InsertAttachment(a Attachment) (string, error) {
	err := r.db.Update(func(tx *bolt.Tx) error {
		// k, err := database.Insert(tx, attachmentBkt, database.IdKvGen(&a))
		// if err != nil {
		// 	return err
		// }
		// _, err = database.Insert(tx, attachmentSuiteOwnerIdxBkt,
		// 	database.IdIndexKvGen(k, a.SuiteId))
		// if err != nil {
		// 	return err
		// }
		// _, err = database.Insert(tx, attachmentCaseOwnerIdxBkt,
		// 	database.IdIndexKvGen(k, a.CaseId))
		// return err
		return nil
	})
	// return a.Id, err
	return "", err
}

func (r *Repo) Attachment(id string) (a Attachment, err error) {
	err = r.db.View(func(tx *bolt.Tx) error {
		// return database.Find(tx, attachmentBkt, database.IdKGen(id), &a)
		return nil
	})
	return
}

func (r *Repo) SuiteAttachments(suiteId string) (a []Attachment, err error) {
	err = r.db.View(func(tx *bolt.Tx) error {
		// database.Ascend(tx, attachment)
		return nil
	})

	// err = wrapNotFoundErr(r.db.Find("SuiteId", suiteId, &a))
	return
}

func (r *Repo) CaseAttachments(caseId string) (a []Attachment, err error) {
	// err = wrapNotFoundErr(r.db.Find("CaseId", caseId, &a))
	return
}
