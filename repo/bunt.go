package repo

import (
	"context"
	"encoding/json"
	"github.com/tidwall/buntdb"
	"github.com/tidwall/sjson"
)

type buntRepo struct {
	changes chan Change
	db      *buntdb.DB
	idGen   func() string

	attachments AttachmentRepo
	cases       CaseRepo
	logs        LogRepo
	suites      SuiteRepo
}

func NewBuntRepos(filename string, idGen func() string) (Repos, error) {
	db, err := buntdb.Open(filename)
	if err != nil {
		return nil, err
	}
	r := &buntRepo{
		changes: make(chan Change),
		db:      db,
		idGen:   idGen,
	}
	if r.attachments, err = r.newAttachmentRepo(); err != nil {
		return nil, err
	}
	if r.cases, err = r.newCaseRepo(); err != nil {
		return nil, err
	}
	if r.logs, err = r.newLogRepo(); err != nil {
		return nil, err
	}
	if r.suites, err = r.newSuiteRepo(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *buntRepo) Attachments(context.Context) AttachmentRepo {
	return r.attachments
}

func (r *buntRepo) Cases(context.Context) CaseRepo {
	return r.cases
}

func (r *buntRepo) Changes() <-chan Change {
	return r.changes
}

func (r *buntRepo) Logs(context.Context) LogRepo {
	return r.logs
}

func (r *buntRepo) Suites(context.Context) SuiteRepo {
	return r.suites
}

func (r *buntRepo) Close() error {
	if err := r.db.Close(); err != nil {
		return err
	}
	close(r.changes)
	return nil
}

func (r *buntRepo) save(e interface{}, collection Collection) (string, error) {
	var id string
	err := r.db.Update(func(tx *buntdb.Tx) error {
		id = r.idGen()
		b, err := json.Marshal(e)
		if err != nil {
			return err
		}
		v, err := sjson.Set(string(b), "id", id)
		if err != nil {
			return err
		}
		_, _, err = tx.Set(string(collection)+":"+id, v, nil)
		if err != nil {
			return err
		}
		var payload interface{}
		if err := json.Unmarshal([]byte(v), &payload); err != nil {
			return err
		}
		r.changes <- Change{
			Op:         ChangeOpInsert,
			Collection: collection,
			Payload:    payload,
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *buntRepo) set(collection Collection, id string, m map[string]interface{}) error {
	return r.db.Update(func(tx *buntdb.Tx) error {
		k := string(collection) + ":" + id
		v, err := tx.Get(k)
		if err != nil {
			return err
		}
		for mK, mV := range m {
			if mV == nil {
				continue
			}
			if v, err = sjson.Set(v, mK, mV); err != nil {
				return err
			}
		}
		if _, _, err = tx.Set(k, v, nil); err != nil {
			return err
		}
		var payload interface{}
		if err := json.Unmarshal([]byte(v), &payload); err != nil {
			return err
		}
		r.changes <- Change{
			Op:         ChangeOpUpdate,
			Collection: collection,
			Payload:    payload,
		}
		return nil
	})
}

func (r *buntRepo) find(collection Collection, id string, e interface{}) error {
	return r.db.View(func(tx *buntdb.Tx) error {
		v, err := tx.Get(string(collection) + ":" + id)
		if err != nil {
			return err
		}
		return json.Unmarshal([]byte(v), e)
	})
}

func (r *buntRepo) findAll(index string, includeDeleted bool, entities interface{}) error {
	var values []string
	err := r.db.View(func(tx *buntdb.Tx) error {
		iterator := func(k, v string) bool {
			values = append(values, v)
			return true
		}
		if includeDeleted {
			return tx.Ascend(index, iterator)
		}
		return tx.AscendEqual(index, `{"deleted":false}`, iterator)
	})
	if err != nil {
		return err
	}
	return valuesToSlice(values, &entities)
}

func (r *buntRepo) findAllBy(index string, m map[string]interface{}, entities interface{}) error {
	var values []string
	err := r.db.View(func(tx *buntdb.Tx) error {
		iterator := func(k, v string) bool {
			values = append(values, v)
			return true
		}
		b, err := json.Marshal(m)
		if err != nil {
			return err
		}
		return tx.AscendEqual(index, string(b), iterator)
	})
	if err != nil {
		return err
	}
	return valuesToSlice(values, &entities)
}

func (r *buntRepo) delete(collection Collection, id string, at int64) error {
	return r.db.Update(func(tx *buntdb.Tx) error {
		k := string(collection) + ":" + id
		v, err := tx.Get(k)
		if err != nil {
			return err
		}
		if v, err = sjson.Set(v, "deleted", true); err != nil {
			return err
		}
		if v, err = sjson.Set(v, "deleted_at", at); err != nil {
			return err
		}
		if _, _, err := tx.Set(k, v, nil); err != nil {
			return err
		}
		var payload interface{}
		if err := json.Unmarshal([]byte(v), &payload); err != nil {
			return err
		}
		r.changes <- Change{
			Op:         ChangeOpUpdate,
			Collection: collection,
			Payload:    payload,
		}
		return nil
	})
}

func (r *buntRepo) deleteAll(collection Collection, index string, at int64) error {
	return r.db.Update(func(tx *buntdb.Tx) error {
		entries := make(map[string]string)
		err := tx.AscendEqual(index, `{"deleted":false}`, func(k, v string) bool {
			entries[k] = v
			return true
		})
		if err != nil {
			return err
		}
		for k, v := range entries {
			if v, err = sjson.Set(v, "deleted", true); err != nil {
				return err
			}
			if v, err = sjson.Set(v, "deleted_at", at); err != nil {
				return err
			}
			if _, _, err := tx.Set(k, v, nil); err != nil {
				return err
			}
			var payload interface{}
			if err := json.Unmarshal([]byte(v), &payload); err != nil {
				return err
			}
			r.changes <- Change{
				Op:         ChangeOpUpdate,
				Collection: collection,
				Payload:    payload,
			}
		}
		return nil
	})
}
