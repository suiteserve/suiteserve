package repo

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tidwall/buntdb"
	"github.com/tidwall/sjson"
	"log"
	"sync/atomic"
	"time"
)

const (
	attachmentColl       = "attachments"
	attachmentOwnerIndex = "attachments_owner"
	caseColl             = "cases"
	suiteColl            = "suites"
)

var (
	ErrNotFound = errors.New("not found")
)

type Entity struct {
	Id string `json:"id"`
}

type VersionedEntity struct {
	Version int64 `json:"version"`
}

type SoftDeleteEntity struct {
	Deleted   bool  `json:"deleted,omitempty"`
	DeletedAt int64 `json:"deleted_at,omitempty"`
}

type Repo struct {
	db *buntdb.DB
}

func Open(filename string) (*Repo, error) {
	db, err := buntdb.Open(filename)
	if err != nil {
		return nil, err
	}
	repo := &Repo{
		db: db,
	}
	if err := repo.setIndexes(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *Repo) Close() error {
	return r.db.Close()
}

func (r *Repo) setIndexes() error {
	err := r.db.ReplaceIndex(attachmentOwnerIndex, attachmentColl+":*",
		buntdb.IndexJSON("suite_id"), buntdb.IndexJSON("case_id"))
	if err != nil {
		return err
	}
	return nil
}

func (r *Repo) insert(coll string, x interface{}) (id string, err error) {
	b, err := json.Marshal(x)
	if err != nil {
		log.Panicf("marshal json: %v", err)
	}
	id = genId()
	if b, err = sjson.SetBytes(b, "id", id); err != nil {
		log.Panicf("set json: %v", err)
	}
	return id, r.db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(coll+":"+id, string(b), nil)
		return err
	})
}

func (r *Repo) getById(coll string, id string, x interface{}) error {
	var v string
	err := r.db.View(func(tx *buntdb.Tx) error {
		var err error
		v, err = tx.Get(coll + ":" + id)
		return err
	})
	if err == buntdb.ErrNotFound {
		return ErrNotFound
	} else if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(v), x); err != nil {
		log.Panicf("unmarshal json: %v", err)
	}
	return nil
}

func unmarshalJsonVals(vals []string, f func(i int) interface{}) {
	for i, v := range vals {
		if err := json.Unmarshal([]byte(v), f(i)); err != nil {
			log.Panicf("unmarshal json: %v", err)
		}
	}
}

var idInc uint32

func genId() string {
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		log.Panicf("read rand: %v", err)
	}
	now := time.Now()
	return fmt.Sprintf("%012x%04x%06x",
		now.Unix()*1e3+int64(now.Nanosecond())/1e6,
		atomic.AddUint32(&idInc, 1)%(1<<16), b)
}
