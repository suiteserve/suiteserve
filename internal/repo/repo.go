package repo

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/suiteserve/suiteserve/event"
	"github.com/tidwall/buntdb"
	"log"
	"sync/atomic"
	"time"
)

type Coll string

const (
	AttachmentColl Coll = "attachments"
	SuiteColl      Coll = "suites"
	SuiteAggColl   Coll = "suite_agg"
	CaseColl       Coll = "cases"
	LogColl        Coll = "logs"
)

const (
	attachmentIndexOwner = "attachments/owner"
	suiteIndexStartedAt  = "suites/started_at"
)

var ErrNotFound = errors.New("not found")

type Entity struct {
	Id string `json:"id"`
}

type VersionedEntity struct {
	Version int64 `json:"version"`
}

type SoftDeleteEntity struct {
	Deleted   bool  `json:"deleted"`
	DeletedAt int64 `json:"deleted_at"`
}

type Repo struct {
	db    *buntdb.DB
	pub   event.Publisher
	idInc uint32
}

func Open(filename string) (*Repo, error) {
	db, err := buntdb.Open(filename)
	if err != nil {
		return nil, err
	}
	repo := Repo{db: db}
	return &repo, repo.setIndexes()
}

func (r *Repo) Changefeed() *event.Bus {
	return &r.pub.Bus
}

func (r *Repo) Close() error {
	return r.db.Close()
}

func (r *Repo) setIndexes() error {
	err := r.db.ReplaceIndex(attachmentIndexOwner, key(AttachmentColl, "*"),
		buntdb.IndexJSON("suite_id"), buntdb.IndexJSON("case_id"))
	if err != nil {
		return err
	}
	return r.db.ReplaceIndex(suiteIndexStartedAt, key(SuiteColl, "*"),
		buntdb.IndexJSON("started_at"))
}

type insertable interface {
	setId(id string)
}

func (r *Repo) insert(coll Coll, x insertable) (id string, err error) {
	return r.insertFunc(coll, x, func(tx *buntdb.Tx) error {
		return nil
	})
}

func (r *Repo) insertFunc(coll Coll, x insertable, after func(tx *buntdb.Tx) error) (id string, err error) {
	id = r.genId()
	x.setId(id)
	return id, r.setFunc(coll, id, x, after)
}

func (r *Repo) setFunc(coll Coll, id string, x interface{}, after func(tx *buntdb.Tx) error) error {
	b, err := json.Marshal(x)
	if err != nil {
		log.Panicf("marshal json: %v", err)
	}
	return r.db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(key(coll, id), string(b), nil)
		if err != nil {
			return err
		}
		return after(tx)
	})
}

func (r *Repo) update(tx *buntdb.Tx, coll Coll, id string, x interface{}, updateX func()) error {
	k := key(coll, id)
	v, err := tx.Get(k)
	if err == nil {
		if err := json.Unmarshal([]byte(v), x); err != nil {
			log.Panicf("unmarshal json: %v", err)
		}
	} else if err != buntdb.ErrNotFound {
		return err
	}
	updateX()
	b, err := json.Marshal(x)
	if err != nil {
		log.Panicf("marshal json: %v", err)
	}
	_, _, err = tx.Set(k, string(b), nil)
	return err
}

func (r *Repo) getById(coll Coll, id string, x interface{}) error {
	var v string
	err := r.db.View(func(tx *buntdb.Tx) error {
		var err error
		v, err = tx.Get(key(coll, id))
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

func (r *Repo) genId() string {
	b := make([]byte, 1)
	if _, err := rand.Read(b); err != nil {
		log.Panicf("read rand: %v", err)
	}
	now := time.Now()
	return fmt.Sprintf("%011x%02x%02x",
		now.Unix()*1e3+int64(now.Nanosecond())/1e6,
		atomic.AddUint32(&r.idInc, 1)&0xff, b)
}

func unmarshalJsonVals(vals []string, f func(i int) interface{}) {
	for i, v := range vals {
		if err := json.Unmarshal([]byte(v), f(i)); err != nil {
			log.Panicf("unmarshal json: %v", err)
		}
	}
}

type less func(v string) bool
type itr func(k, v string) bool

func newPageItr(limit int, vals *[]string, hasMore *bool, less less) itr {
	var n int
	return func(k, v string) bool {
		if less(v) {
			if n == limit {
				*hasMore = true
				return false
			}
			n++
		}
		*vals = append(*vals, v)
		return true
	}
}

func key(coll Coll, id string) string {
	return string(coll) + ":" + id
}
