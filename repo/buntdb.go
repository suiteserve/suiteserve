package repo

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/suiteserve/suiteserve/event"
	"github.com/tidwall/buntdb"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"log"
	"reflect"
	"strconv"
	"sync/atomic"
	"time"
)

type BuntDb struct {
	changes event.Publisher
	db      *buntdb.DB
	files   *FileRepo
}

func OpenBuntDb(filename string, files *FileRepo) (*BuntDb, error) {
	db, err := buntdb.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("open: %v", err)
	}
	d := BuntDb{
		db:    db,
		files: files,
	}
	if err := d.createIndexes(); err != nil {
		return nil, fmt.Errorf("create indexes: %v", err)
	}
	return &d, nil
}

func (d *BuntDb) createIndexes() error {
	// attachments
	err := d.db.ReplaceIndex("attachments_id", "attachments:*",
		buntdb.IndexJSON("id"))
	if err != nil {
		return err
	}
	err = d.db.ReplaceIndex("attachments_deleted", "attachments:*",
		buntdb.IndexJSON("deleted"),
		indexJSONOptional("id"))
	if err != nil {
		return err
	}
	// cases
	err = d.db.ReplaceIndex("cases_suite", "cases:*",
		buntdb.IndexJSON("suite"),
		indexJSONOptional("num"),
		indexJSONOptional("created_at"),
		indexJSONOptional("id"))
	if err != nil {
		return err
	}
	// logs
	err = d.db.ReplaceIndex("logs_case", "logs:*",
		buntdb.IndexJSON("case"),
		buntdb.IndexJSON("timestamp"),
		buntdb.IndexJSON("seq"),
		buntdb.IndexJSON("id"))
	if err != nil {
		return err
	}
	// suites
	err = d.db.ReplaceIndex("suites_status", "suites:*",
		buntdb.IndexJSON("status"))
	if err != nil {
		return err
	}
	err = d.db.ReplaceIndex("suites_deleted", "suites:*",
		buntdb.IndexJSON("deleted"),
		indexJSONOptional("started_at"),
		indexJSONOptional("id"))
	if err != nil {
		return err
	}
	return nil
}

func indexJSONOptional(path string) func(a, b string) bool {
	return func(a, b string) bool {
		aRes := gjson.Get(a, path)
		bRes := gjson.Get(b, path)
		if aRes.Exists() && bRes.Exists() {
			return aRes.Less(bRes, false)
		}
		return false
	}
}

func (d *BuntDb) Changes() *event.Bus {
	return &d.changes.Bus
}

func (d *BuntDb) Seedable() (bool, error) {
	n := 0
	err := d.db.View(func(tx *buntdb.Tx) error {
		var err error
		n, err = tx.Len()
		return err
	})
	if err != nil {
		return false, err
	}
	return n == 0, nil
}

func (d *BuntDb) Close() error {
	if err := d.db.Close(); err != nil {
		return fmt.Errorf("close: %v", err)
	}
	return nil
}

type buntDbUpdate struct {
	oldV string
	newV string
}

func (d *BuntDb) onUpdate(tx *buntdb.Tx, coll Coll, updates []buntDbUpdate) error {
	switch coll {
	case CollSuites:
		return d.onSuiteUpdate(tx, updates)
	default:
		return nil
	}
}

func (d *BuntDb) onSuiteUpdate(tx *buntdb.Tx, updates []buntDbUpdate) error {
	runningDelta := int64(0)
	finishedDelta := int64(0)
	for _, u := range updates {
		oldStatus := gjson.Get(u.oldV, "status").String()
		newStatus := gjson.Get(u.newV, "status").String()
		runningStr := string(SuiteStatusRunning)
		if u.oldV == "" {
			if newStatus == runningStr {
				runningDelta++
			} else {
				finishedDelta++
			}
		} else if u.newV == "" {
			if oldStatus == runningStr {
				runningDelta--
			} else {
				finishedDelta--
			}
		} else if oldStatus == runningStr && newStatus != runningStr {
			runningDelta--
			finishedDelta++
		} else if oldStatus != runningStr && newStatus == runningStr {
			runningDelta++
			finishedDelta--
		}
	}
	var changed bool
	m := map[string]interface{}{
		"version":  0,
		"running":  0,
		"finished": 0,
	}
	if runningDelta != 0 {
		n, err := incInt(tx, buntDbKey(CollSuiteAggs, "running"), runningDelta)
		if err != nil {
			return err
		}
		m["running"] = n
		changed = true
	}
	if finishedDelta != 0 {
		n, err := incInt(tx, buntDbKey(CollSuiteAggs, "finished"), finishedDelta)
		if err != nil {
			return err
		}
		m["finished"] = n
		changed = true
	}
	if !changed {
		return nil
	}
	var err error
	m["version"], err = incInt(tx, buntDbKey(CollSuiteAggs, "version"), 1)
	if err != nil {
		return err
	}
	d.changes.Publish(newUpdateDocChange(CollSuiteAggs, "", m, []string{}))
	return nil
}

func (d *BuntDb) insert(coll Coll, e interface{}) (string, error) {
	return d.funcInsert(coll, e, func(_ string) error {
		return nil
	})
}

func (d *BuntDb) funcInsert(coll Coll, e interface{}, fn func(id string) error) (string, error) {
	id := newBuntDbId()
	if err := fn(id); err != nil {
		return "", err
	}
	b, err := json.Marshal(&e)
	if err != nil {
		log.Panicf("marshal json: %v\n", err)
	}
	v, err := sjson.SetBytes(b, "id", id)
	if err != nil {
		log.Panicf("set json: %v\n", err)
	}
	err = d.db.Update(func(tx *buntdb.Tx) error {
		v := string(v)
		if err := d.onUpdate(tx, coll, []buntDbUpdate{{"", v}}); err != nil {
			return err
		}
		_, _, err = tx.Set(buntDbKey(coll, id), v, nil)
		return err
	})
	if err != nil {
		return "", err
	}
	d.changes.Publish(newInsertDocChange(coll, id, b))
	return id, nil
}

var buntDbIdInc int64

func newBuntDbId() string {
	b := make([]byte, 3)
	if _, err := rand.Reader.Read(b); err != nil {
		log.Panicf("read rand: %v\n", err)
	}
	inc := atomic.AddInt64(&buntDbIdInc, 1)
	return fmt.Sprintf("%x%x%x", time.Now().Unix(), inc, b)
}

func (d *BuntDb) update(coll Coll, id string, m map[string]interface{}) error {
	return d.funcUpdate(coll, id, func(_ string) (map[string]interface{}, error) {
		return m, nil
	})
}

func (d *BuntDb) funcUpdate(coll Coll, id string, fn func(v string) (map[string]interface{}, error)) error {
	k := buntDbKey(coll, id)
	var updated map[string]interface{}
	err := d.db.Update(func(tx *buntdb.Tx) error {
		oldV, err := tx.Get(k)
		if err != nil {
			return err
		}
		if updated, err = fn(oldV); err != nil {
			return err
		}
		updated["version"] = gjson.Get(oldV, "version").Int() + 1
		var newV string
		newV, updated = applyUpdateMap(oldV, updated)
		if err := d.onUpdate(tx, coll, []buntDbUpdate{{oldV, newV}}); err != nil {
			return err
		}
		_, _, err = tx.Set(k, newV, nil)
		return err
	})
	if err == buntdb.ErrNotFound {
		return ErrNotFound
	} else if err != nil {
		return err
	}
	d.changes.Publish(newUpdateDocChange(coll, id, updated, []string{}))
	return nil
}

func (d *BuntDb) find(coll Coll, id string, e interface{}) error {
	k := buntDbKey(coll, id)
	var v string
	err := d.db.View(func(tx *buntdb.Tx) error {
		var err error
		v, err = tx.Get(k)
		return err
	})
	if err == buntdb.ErrNotFound {
		return ErrNotFound
	} else if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(v), e); err != nil {
		log.Panicf("unmarshal json: %v\n", err)
	}
	return nil
}

func (d *BuntDb) findAllBy(index string, m map[string]interface{}, entities interface{}) error {
	pivot, err := json.Marshal(m)
	if err != nil {
		log.Panicf("marshal json: %v\n", err)
	}
	var vals []string
	err = d.db.View(func(tx *buntdb.Tx) error {
		return tx.AscendEqual(index, string(pivot), func(k, v string) bool {
			vals = append(vals, v)
			return true
		})
	})
	if err != nil {
		return err
	}
	jsonValsToArr(vals, entities)
	return nil
}

func (d *BuntDb) delete(coll Coll, id string, at int64) error {
	k := buntDbKey(coll, id)
	var updated map[string]interface{}
	err := d.db.Update(func(tx *buntdb.Tx) error {
		oldV, err := tx.Get(k)
		if err != nil {
			return err
		}
		var newV string
		newV, updated = applyUpdateMap(oldV, map[string]interface{}{
			"deleted":    true,
			"deleted_at": at,
			"version":    gjson.Get(oldV, "version").Int() + 1,
		})
		if err := d.onUpdate(tx, coll, []buntDbUpdate{{oldV, newV}}); err != nil {
			return err
		}
		_, _, err = tx.Set(k, newV, nil)
		return err
	})
	if err == buntdb.ErrNotFound {
		return ErrNotFound
	} else if err != nil {
		return err
	}
	d.changes.Publish(newUpdateDocChange(coll, id, updated, []string{}))
	return nil
}

func (d *BuntDb) deleteAll(coll Coll, index string, at int64) error {
	var updated map[string]map[string]interface{}
	var setErr error
	err := d.db.Update(func(tx *buntdb.Tx) error {
		var updates []buntDbUpdate
		err := tx.AscendEqual(index, `{"deleted":false}`, func(k, v string) bool {
			updates = append(updates, buntDbUpdate{v, ""})
			v, innerUpdated := applyUpdateMap(v, map[string]interface{}{
				"deleted":    true,
				"deleted_at": at,
				"version":    gjson.Get(v, "version").Int() + 1,
			})
			updated[gjson.Get(v, "id").String()] = innerUpdated
			if _, _, err := tx.Set(k, v, nil); err != nil {
				setErr = err
				return false
			}
			return true
		})
		if err != nil {
			return err
		}
		return d.onUpdate(tx, coll, updates)
	})
	if setErr != nil {
		return setErr
	} else if err != nil {
		return err
	}
	for id, innerUpdated := range updated {
		d.changes.Publish(newUpdateDocChange(coll, id, innerUpdated, []string{}))
	}
	return nil
}

func incInt(tx *buntdb.Tx, k string, delta int64) (new int64, err error) {
	new, err = getInt(tx, k)
	if err != nil {
		return 0, err
	}
	new += delta
	if _, _, err := tx.Set(k, strconv.FormatInt(new, 10), nil); err != nil {
		return 0, err
	}
	return new, nil
}

func getInt(tx *buntdb.Tx, k string) (int64, error) {
	old, err := tx.Get(k)
	if err == buntdb.ErrNotFound {
		return 0, nil
	} else if err != nil {
		return 0, err
	}
	i, err := strconv.ParseInt(old, 10, 64)
	if err != nil {
		log.Panicf("parse value as int at key %q: %v\n", k, err)
	}
	return i, err
}

func applyUpdateMap(s string, m map[string]interface{}) (string, map[string]interface{}) {
	updated := make(map[string]interface{}, len(m))
	for k, v := range m {
		if v == nil {
			continue
		}
		var err error
		if s, err = sjson.Set(s, k, v); err != nil {
			log.Panicf("set json: %v\n", err)
		}
		updated[k] = v
	}
	return s, updated
}

func jsonValsToArr(vals []string, arr interface{}) {
	arrType := reflect.TypeOf(arr)
	if arrType.Kind() != reflect.Ptr || arrType.Elem().Kind() != reflect.Slice {
		log.Panicf("bad arr type: expected ptr to slice")
	}
	arrVal := reflect.ValueOf(arr).Elem()
	arrVal.Set(reflect.MakeSlice(arrType.Elem(), len(vals), len(vals)))
	for i, v := range vals {
		if err := json.Unmarshal([]byte(v), arrVal.Index(i).Addr().Interface()); err != nil {
			log.Panicf("unmarshal json: %v\n", err)
		}
	}
}

func buntDbKey(coll Coll, id string) string {
	return string(coll) + ":" + id
}
