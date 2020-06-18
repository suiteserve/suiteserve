package repo

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/tidwall/buntdb"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/tmazeika/testpass/event"
	"log"
	"reflect"
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
	if err := createBuntDbIndexes(db); err != nil {
		return nil, fmt.Errorf("create indexes: %v", err)
	}
	return &BuntDb{
		db:    db,
		files: files,
	}, nil
}

func createBuntDbIndexes(db *buntdb.DB) error {
	// attachments
	err := db.ReplaceIndex("attachments_id", "attachments:*",
		buntdb.IndexJSON("id"))
	if err != nil {
		return err
	}
	err = db.ReplaceIndex("attachments_deleted", "attachments:*",
		buntdb.IndexJSON("deleted"),
		buntDbIndexJSONOptional("id"))
	if err != nil {
		return err
	}
	// cases
	err = db.ReplaceIndex("cases_suite", "cases:*",
		buntdb.IndexJSON("suite"),
		buntDbIndexJSONOptional("num"),
		buntDbIndexJSONOptional("created_at"),
		buntDbIndexJSONOptional("id"))
	if err != nil {
		return err
	}
	// logs
	err = db.ReplaceIndex("logs_case", "logs:*",
		buntdb.IndexJSON("case"),
		buntdb.IndexJSON("timestamp"),
		buntdb.IndexJSON("seq"),
		buntdb.IndexJSON("id"))
	if err != nil {
		return err
	}
	// suites
	err = db.ReplaceIndex("suites_status", "suites:*",
		buntdb.IndexJSON("status"))
	if err != nil {
		return err
	}
	err = db.ReplaceIndex("suites_deleted", "suites:*",
		buntdb.IndexJSON("deleted"),
		buntDbIndexJSONOptional("started_at"),
		buntDbIndexJSONOptional("id"))
	if err != nil {
		return err
	}
	return nil
}

func buntDbIndexJSONOptional(path string) func(a, b string) bool {
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

func (d *BuntDb) Seedable() bool {
	n := 0
	err := d.db.View(func(tx *buntdb.Tx) error {
		var err error
		n, err = tx.Len()
		return err
	})
	if err != nil {
		log.Fatalf("check buntdb seedability: %v\n", err)
	}
	return n == 0
}

func (d *BuntDb) Close() error {
	if err := d.db.Close(); err != nil {
		return fmt.Errorf("close: %v", err)
	}
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
		_, _, err = tx.Set(buntDbKey(coll, id), string(v), nil)
		return err
	})
	if err != nil {
		return "", err
	}
	d.changes.Publish(newInsertChange(coll, v))
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
		v, err := tx.Get(k)
		if err != nil {
			return err
		}
		if updated, err = fn(v); err != nil {
			return err
		}
		updated["version"] = gjson.Get(v, "version").Int() + 1
		v, updated = applyUpdateMap(v, updated)
		_, _, err = tx.Set(k, v, nil)
		return err
	})
	if err == buntdb.ErrNotFound {
		return ErrNotFound
	} else if err != nil {
		return err
	}
	d.changes.Publish(newUpdateChange(coll, id, updated, []string{}))
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
	if err := json.Unmarshal([]byte(v), &e); err != nil {
		log.Panicf("unmarshal json: %v\n", err)
	}
	return nil
}

func (d *BuntDb) findAll(index string, entities interface{}) error {
	var vals []string
	err := d.db.View(func(tx *buntdb.Tx) error {
		return tx.AscendEqual(index, `{"deleted":false}`, func(k, v string) bool {
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
		v, err := tx.Get(k)
		if err != nil {
			return err
		}
		v, updated = applyUpdateMap(v, map[string]interface{}{
			"deleted":    true,
			"deleted_at": at,
			"version":    gjson.Get(v, "version").Int() + 1,
		})
		_, _, err = tx.Set(k, v, nil)
		return err
	})
	if err == buntdb.ErrNotFound {
		return ErrNotFound
	} else if err != nil {
		return err
	}
	d.changes.Publish(newUpdateChange(coll, id, updated, []string{}))
	return nil
}

func (d *BuntDb) deleteAll(coll Coll, index string, at int64) error {
	var updated map[string]map[string]interface{}
	var setErr error
	err := d.db.Update(func(tx *buntdb.Tx) error {
		return tx.AscendEqual(index, `{"deleted":false}`, func(k, v string) bool {
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
	})
	if setErr != nil {
		return setErr
	} else if err != nil {
		return err
	}
	for id, innerUpdated := range updated {
		d.changes.Publish(newUpdateChange(coll, id, innerUpdated, []string{}))
	}
	return nil
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
	return fmt.Sprintf("%s:%s", coll, id)
}
