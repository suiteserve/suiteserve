package repo

import (
	"context"
	"github.com/tidwall/buntdb"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"strconv"
	"strings"
	"time"
)

type buntSuiteRepo struct {
	*buntRepo
}

func (r *buntRepo) newSuiteRepo() (*buntSuiteRepo, error) {
	err := r.db.ReplaceIndex("suites_id", "suites:*",
		buntdb.IndexJSON("id"))
	if err != nil {
		return nil, err
	}
	err = r.db.ReplaceIndex("suites_status", "suites:*",
		buntdb.IndexJSON("status"))
	if err != nil {
		return nil, err
	}
	err = r.db.ReplaceIndex("suites_deleted", "suites:*",
		buntdb.IndexJSON("deleted"), indexJSONOptional("id"))
	if err != nil {
		return nil, err
	}
	return &buntSuiteRepo{r}, nil
}

func (r *buntSuiteRepo) Save(_ context.Context, s UnsavedSuite) (string, error) {
	return r.save(SuiteColl, &s)
}

func (r *buntSuiteRepo) SaveAttachment(_ context.Context, id string, attachmentId string) error {
	return r.set(CaseColl, id, map[string]interface{}{
		"attachments.-1": attachmentId,
	})
}

func (r *buntSuiteRepo) SaveStatus(_ context.Context, id string, status SuiteStatus, opts *SuiteRepoSaveStatusOptions) error {
	return r.set(SuiteColl, id, map[string]interface{}{
		"status":      status,
		"finished_at": opts.finishedAt,
		"disconnected_at": opts.disconnectedAt,
	})
}

func (r *buntSuiteRepo) Page(_ context.Context, fromId *string, n int64, includeDeleted bool) (*SuitePage, error) {
	index := "suites_id"
	var running int64
	var finished int64
	var nextId *string
	values := make([]string, 0)
	iterator := func(k, v string) bool {
		id := gjson.Get(v, "id").String()
		if int64(len(values)) == n {
			nextId = &id
			return false
		}
		if includeDeleted || !gjson.Get(v, "deleted").Bool() {
			values = append(values, v)
		}
		return true
	}
	err := r.db.View(func(tx *buntdb.Tx) error {
		var err error
		if fromId == nil {
			err = tx.Descend(index, iterator)
		} else {
			escapedFromId := strconv.Quote(*fromId)
			err = tx.DescendLessOrEqual(index, `{"id":`+escapedFromId+`}`, iterator)
		}
		if err != nil {
			return err
		}

		running, err = r.count(tx, "suites_status", "status", string(SuiteStatusRunning))
		if err != nil {
			return err
		}
		passed, err := r.count(tx, "suites_status", "status", string(SuiteStatusPassed))
		if err != nil {
			return err
		}
		failed, err := r.count(tx, "suites_status", "status", string(SuiteStatusFailed))
		if err != nil {
			return err
		}
		finished = passed + failed
		return nil
	})
	if err != nil {
		return nil, err
	}
	var suites []Suite
	if err := jsonValuesToArr(values, &suites); err != nil {
		return nil, err
	}
	return &SuitePage{
		RunningCount:  running,
		FinishedCount: finished,
		NextId:        nextId,
		Suites:        suites,
	}, nil
}

func (r *buntSuiteRepo) Find(_ context.Context, id string) (*Suite, error) {
	var s Suite
	if err := r.find(SuiteColl, id, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *buntSuiteRepo) Reconnect(_ context.Context, id string, at int64, ttl time.Duration) error {
	k := string(SuiteColl) + ":" + id
	err := r.db.Update(func(tx *buntdb.Tx) error {
		v, err := tx.Get(k)
		if err != nil {
			return err
		}
		if gjson.Get(v, "status").String() != string(SuiteStatusDisconnected) {
			return ErrNotReconnectable
		}
		now := time.Unix(at, 0)
		disconnectedAt := time.Unix(gjson.Get(v, "disconnected_at").Int(), 0)
		exp := disconnectedAt.Add(ttl)
		if !now.Before(exp) {
			return ErrExpired
		}
		if v, err = sjson.Set(v, "status", SuiteStatusRunning); err != nil {
			return err
		}
		if v, err = sjson.Set(v, "disconnected_at", 0); err != nil {
			return err
		}
		if _, _, err = tx.Set(k, v, nil); err != nil {
			return err
		}
		change, err := newChangeFromJson(ChangeOpUpdate, SuiteColl, v)
		if err != nil {
			return err
		}
		r.changes <- *change
		return nil
	})
	if err == buntdb.ErrNotFound {
		return ErrNotFound
	}
	return err
}

func (r *buntSuiteRepo) FuzzyFind(_ context.Context, fuzzyIdOrName string, includeDeleted bool) ([]Suite, error) {
	values := make([]string, 0)
	err := r.db.View(func(tx *buntdb.Tx) error {
		return tx.Ascend("suites_id", func(k, v string) bool {
			id := gjson.Get(v, "id").String()
			name := gjson.Get(v, "name").String()
			deleted := gjson.Get(v, "deleted").Bool()
			idMatched := strings.Contains(id, fuzzyIdOrName)
			nameMatched := strings.Contains(name, fuzzyIdOrName)
			if (includeDeleted || !deleted) && (idMatched || nameMatched) {
				values = append(values, v)
			}
			return true
		})
	})
	if err != nil {
		return nil, err
	}
	var suites []Suite
	if err := jsonValuesToArr(values, &suites); err != nil {
		return nil, err
	}
	return suites, nil
}

func (r *buntSuiteRepo) FindAll(_ context.Context, includeDeleted bool) ([]Suite, error) {
	index := "suites_deleted"
	if includeDeleted {
		index = "suites_id"
	}
	var suites []Suite
	if err := r.findAll(index, includeDeleted, &suites); err != nil {
		return nil, err
	}
	return suites, nil
}

func (r *buntSuiteRepo) Delete(_ context.Context, id string, at int64) error {
	return r.delete(SuiteColl, id, at)
}

func (r *buntSuiteRepo) DeleteAll(_ context.Context, at int64) error {
	_, err := r.deleteAll(SuiteColl, "suites_deleted", at)
	return err
}

func (r *buntSuiteRepo) count(tx *buntdb.Tx, index, k, v string) (int64, error) {
	k = strconv.Quote(k)
	v = strconv.Quote(v)
	var n int64
	err := tx.AscendEqual(index, `{`+k+`:`+v+`}`, func(k, v string) bool {
		n++
		return true
	})
	if err != nil {
		return 0, err
	}
	return n, nil
}
