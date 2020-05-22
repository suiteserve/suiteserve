package repo

import (
	"encoding/json"
	"github.com/tidwall/buntdb"
	"github.com/tidwall/sjson"
	"github.com/tmazeika/testpass/config"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strings"
	"time"
)

type Collection string

const (
	AttachmentCollection Collection = "attachments"
	CaseCollection       Collection = "cases"
	LogCollection        Collection = "logs"
	SuiteCollection      Collection = "suites"
)

type Repos interface {
	Attachments() AttachmentRepo
	Cases() CaseRepo
	Changes() <-chan Change
	Logs() LogRepo
	Suites() SuiteRepo
	Close() error
}

type buntRepo struct {
	changes chan Change
	db      *buntdb.DB

	attachments AttachmentRepo
	cases       CaseRepo
	logs        LogRepo
	suites      SuiteRepo
}

func NewBuntRepos(cfg *config.Config) (Repos, error) {
	db, err := buntdb.Open(cfg.Storage.Bunt.File)
	if err != nil {
		return nil, err
	}
	r := &buntRepo{
		changes: make(chan Change),
		db:      db,
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

func (r *buntRepo) Attachments() AttachmentRepo {
	return r.attachments
}

func (r *buntRepo) Cases() CaseRepo {
	return r.cases
}

func (r *buntRepo) Changes() <-chan Change {
	return r.changes
}

func (r *buntRepo) Logs() LogRepo {
	return r.logs
}

func (r *buntRepo) Suites() SuiteRepo {
	return r.suites
}

func (r *buntRepo) Close() error {
	return r.db.Close()
}

func (r *buntRepo) save(e interface{}, collection Collection) (string, error) {
	var id string
	err := r.db.Update(func(tx *buntdb.Tx) error {
		id = primitive.NewObjectID().Hex()
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

func (r *buntRepo) findAll(deletedIndex string, includeDeleted bool, entities interface{}) error {
	var values []string
	err := r.db.View(func(tx *buntdb.Tx) error {
		iterator := func(k, v string) bool {
			values = append(values, v)
			return true
		}
		if includeDeleted {
			return tx.Ascend("", iterator)
		}
		return tx.AscendEqual(deletedIndex, `{"deleted":false}`, iterator)
	})
	if err != nil {
		return err
	}
	return r.valuesToSlice(values, &entities)
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
	return r.valuesToSlice(values, &entities)
}

func (r *buntRepo) delete(collection Collection, id string) error {
	return r.db.Update(func(tx *buntdb.Tx) error {
		k := string(collection) + ":" + id
		v, err := tx.Get(k)
		if err != nil {
			return err
		}
		if v, err = sjson.Set(v, "deleted", true); err != nil {
			return err
		}
		if v, err = sjson.Set(v, "deleted_at", nowTimeMillis()); err != nil {
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

func (r *buntRepo) deleteAll(collection Collection, deletedIndex string) error {
	return r.db.Update(func(tx *buntdb.Tx) error {
		entries := make(map[string]string)
		err := tx.AscendEqual(deletedIndex, `{"deleted":false}`, func(k, v string) bool {
			entries[k] = v
			return true
		})
		if err != nil {
			return err
		}
		deletedAt := nowTimeMillis()
		for k, v := range entries {
			if v, err = sjson.Set(v, "deleted", true); err != nil {
				return err
			}
			if v, err = sjson.Set(v, "deleted_at", deletedAt); err != nil {
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

func (r *buntRepo) valuesToSlice(values []string, slice interface{}) error {
	v := "[" + strings.Join(values, ",") + "]"
	return json.Unmarshal([]byte(v), slice)
}

type Entity struct {
	Id string `json:"id" bson:"_id,omitempty"`
}

type SoftDeleteEntity struct {
	*Entity   `bson:",inline"`
	Deleted   bool  `json:"deleted"`
	DeletedAt int64 `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
}

func nowTimeMillis() int64 {
	now := time.Now()
	// Doesn't use now.UnixNano() to avoid Y2K262.
	return now.Unix()*time.Second.Milliseconds() +
		time.Duration(now.Nanosecond()).Milliseconds()
}
