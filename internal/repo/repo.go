package repo

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/suiteserve/suiteserve/event"
	"github.com/tidwall/buntdb"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"log"
	"strconv"
	"sync/atomic"
	"time"
)

type Coll string

const (
	AttachmentColl Coll = "attachments"
	SuiteColl      Coll = "suites"
	CaseColl       Coll = "cases"
	LogColl        Coll = "logs"

	attachmentIndexOwner = "attachments/owner"
	suiteKeyVersion      = "suites_version"
	suiteKeyRunning      = "suites_running"
	suiteKeyTotal        = "suites_total"
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
	DeletedAt int64 `json:"deleted_at,omitempty"`
}

type Repo struct {
	db  *buntdb.DB
	pub event.Publisher
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

func (r *Repo) Changefeed() *event.Bus {
	return &r.pub.Bus
}

func (r *Repo) Close() error {
	return r.db.Close()
}

func (r *Repo) setIndexes() error {
	err := r.db.ReplaceIndex(attachmentIndexOwner, string(AttachmentColl)+":*",
		buntdb.IndexJSON("suite_id"), buntdb.IndexJSON("case_id"))
	if err != nil {
		return err
	}
	return nil
}

type diff struct {
	old    []byte
	new    []byte
	change *Change
}

func (r *Repo) onUpdate(tx *buntdb.Tx, coll Coll, diffs ...*diff) error {
	var err error
	var changefeed Changefeed
	if coll == SuiteColl {
		changefeed, err = r.onSuitesUpdate(tx, diffs)
	}
	if err != nil {
		return err
	}
	if len(changefeed) > 0 {
		r.pub.Publish(changefeed)
	}
	return nil
}

func (r *Repo) onSuitesUpdate(tx *buntdb.Tx, diffs []*diff) (changefeed Changefeed, err error) {
	var runningDelta, totalDelta int64
	for _, d := range diffs {
		if len(d.old) == 0 {
			// entity was inserted
			totalDelta++
		} else if gjson.GetBytes(d.old, "status").String() == string(SuiteStatusStarted) {
			// entity was updated or deleted; old status was `started`
			runningDelta--
		}
		if len(d.new) == 0 {
			// entity was deleted
			totalDelta--
		} else if gjson.GetBytes(d.new, "status").String() == string(SuiteStatusStarted) {
			// entity was inserted or updated; new status is `started`
			runningDelta++
		}
		changefeed = append(changefeed, d.change)
	}

	agg := make(map[string]int64)
	if agg["version"], err = incInt(tx, suiteKeyVersion, 1); err != nil {
		return nil, err
	}
	if agg["running"], err = incInt(tx, suiteKeyRunning, runningDelta); err != nil {
		return nil, err
	}
	if agg["total"], err = incInt(tx, suiteKeyTotal, totalDelta); err != nil {
		return nil, err
	}
	changefeed = append(changefeed, &Change{
		Id:      string(SuiteColl),
		Op:      ChangeOpUpdate,
		Updated: agg,
	})
	return changefeed, nil
}

func (r *Repo) insert(coll Coll, x interface{}) (id string, err error) {
	b, err := json.Marshal(x)
	if err != nil {
		log.Panicf("marshal json: %v", err)
	}
	id = genId()
	if b, err = sjson.SetBytes(b, "id", id); err != nil {
		log.Panicf("set json: %v", err)
	}
	return id, r.db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(string(coll)+":"+id, string(b), nil)
		if err != nil {
			return err
		}
		return r.onUpdate(tx, coll, &diff{
			new: b,
			change: &Change{
				Id:      id,
				Op:      ChangeOpInsert,
				Updated: json.RawMessage(b),
			},
		})
	})
}

func (r *Repo) getById(coll Coll, id string, x interface{}) error {
	var v string
	err := r.db.View(func(tx *buntdb.Tx) error {
		var err error
		v, err = tx.Get(string(coll) + ":" + id)
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

func incInt(tx *buntdb.Tx, k string, delta int64) (new int64, err error) {
	new, err = getInt(tx, k)
	if err != nil {
		return 0, err
	}
	if delta == 0 {
		return new, nil
	}
	new += delta
	if _, _, err := tx.Set(k, strconv.FormatInt(new, 10), nil); err != nil {
		return 0, err
	}
	return new, nil
}

func getInt(tx *buntdb.Tx, k string) (int64, error) {
	v, err := tx.Get(k)
	if err == buntdb.ErrNotFound {
		return 0, nil
	} else if err != nil {
		return 0, err
	}
	i, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		log.Panicf("parse int: %v", err)
	}
	return i, err
}

var idInc uint32

func genId() string {
	b := make([]byte, 1)
	if _, err := rand.Read(b); err != nil {
		log.Panicf("read rand: %v", err)
	}
	now := time.Now()
	return fmt.Sprintf("%011x%02x%02x",
		now.Unix()*1e3+int64(now.Nanosecond())/1e6,
		atomic.AddUint32(&idInc, 1)&0xff, b)
}
