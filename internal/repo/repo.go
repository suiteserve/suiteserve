package repo

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tidwall/buntdb"
	"github.com/tidwall/gjson"
	"sync"
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

func suiteIndexStartedAtPivot(v string) string {
	return `{"started_at":` + gjson.Get(v, "started_at").String() +
		`,"id":"` + gjson.Get(v, "id").String() + `"}`
}

type foundErr interface {
	Found() bool
}

type errNotFound struct{}

func (e errNotFound) Error() string {
	return "not found"
}

func (e errNotFound) Is(target error) bool {
	var foundErr foundErr
	return errors.As(target, &foundErr) && e.Found() == foundErr.Found()
}

func (e errNotFound) Found() bool {
	return false
}

type Entity struct {
	Id string `json:"id"`
}

func (e *Entity) setId(id string) {
	e.Id = id
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
	idInc uint32

	mu       sync.Mutex
	handlers []changeHandler
}

func Open(filename string) (*Repo, error) {
	db, err := buntdb.Open(filename)
	if err != nil {
		return nil, err
	}
	repo := Repo{db: db}
	return &repo, repo.setIndexes()
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
		buntdb.IndexJSON("started_at"), buntdb.IndexJSON("id"))
}

func (r *Repo) addHandler(h changeHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers = append(r.handlers, h)
}

func (r *Repo) removeHandler(h changeHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, v := range r.handlers {
		if v == h {
			r.handlers = append(r.handlers[:i], r.handlers[i+1:]...)
			return
		}
	}
}

func (r *Repo) notifyHandlers(changes []Change) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, h := range r.handlers {
		h.handleChanges(changes)
	}
}

type insertable interface {
	setId(id string)
}

func (r *Repo) insert(coll Coll, x insertable) (id string, err error) {
	return r.insertFunc(coll, x, func(tx *buntdb.Tx) error {
		return nil
	})
}

func (r *Repo) insertFunc(coll Coll, x insertable,
	after func(tx *buntdb.Tx) error) (id string, err error) {
	id = r.genId()
	x.setId(id)
	return id, r.setFunc(coll, id, x, after)
}

func (r *Repo) setFunc(coll Coll, id string, x interface{},
	after func(tx *buntdb.Tx) error) error {
	b, err := json.Marshal(x)
	if err != nil {
		panic(err)
	}
	return r.db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(key(coll, id), string(b), nil)
		if err != nil {
			return err
		}
		return after(tx)
	})
}

func (r *Repo) update(tx *buntdb.Tx, coll Coll, id string, x interface{},
	updateX func()) error {
	k := key(coll, id)
	v, err := tx.Get(k)
	if err == nil {
		if err := json.Unmarshal([]byte(v), x); err != nil {
			panic(err)
		}
	} else if err != buntdb.ErrNotFound {
		return err
	}
	updateX()
	b, err := json.Marshal(x)
	if err != nil {
		panic(err)
	}
	_, _, err = tx.Set(k, string(b), nil)
	return err
}

func (r *Repo) byId(coll Coll, id string, x interface{}) error {
	return r.db.View(func(tx *buntdb.Tx) error {
		return r.byIdTx(tx, coll, id, x)
	})
}

func (r *Repo) byIdTx(tx *buntdb.Tx, coll Coll, id string,
	x interface{}) error {
	v, err := tx.Get(key(coll, id))
	if err == buntdb.ErrNotFound {
		return errNotFound{}
	} else if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(v), x); err != nil {
		panic(err)
	}
	return nil
}

func (r *Repo) genId() string {
	b := make([]byte, 1)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	now := time.Now()
	return fmt.Sprintf("%011x%02x%02x",
		now.Unix()*1e3+int64(now.Nanosecond())/1e6,
		atomic.AddUint32(&r.idInc, 1)&0xff, b)
}

func unmarshalJsonVals(vals []string, fn func(i int) interface{}) {
	for i, v := range vals {
		if err := json.Unmarshal([]byte(v), fn(i)); err != nil {
			panic(err)
		}
	}
}

type entry struct {
	k string
	v string
}
type itr func(k, v string) bool

func limitCond(limit int) itr {
	if limit < 0 {
		return func(k, v string) bool {
			return true
		}
	}
	var n int
	return func(k, v string) bool {
		if n == limit {
			return false
		}
		n++
		return true
	}
}

func firstCond(first itr) itr {
	var fn itr
	rest := func(k, v string) bool {
		return true
	}
	fn = func(k, v string) bool {
		fn = rest
		return first(k, v)
	}
	return func(k, v string) bool {
		return fn(k, v)
	}
}

func restCond(rest itr) itr {
	var fn itr
	fn = func(k, v string) bool {
		fn = rest
		return true
	}
	return func(k, v string) bool {
		return fn(k, v)
	}
}

// andCond returns a new itr that calls each given cond, returning true iff all
// of the given conds return true for that entry. Short-circuits as logical AND
// would.
func andCond(conds ...itr) itr {
	return func(k, v string) bool {
		for _, fn := range conds {
			if !fn(k, v) {
				return false
			}
		}
		return true
	}
}

func key(coll Coll, id string) string {
	return string(coll) + ":" + id
}
