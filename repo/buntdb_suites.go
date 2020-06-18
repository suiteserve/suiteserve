package repo

import (
	"context"
	"encoding/json"
	"github.com/tidwall/buntdb"
	"github.com/tidwall/gjson"
	"log"
	"strconv"
	"time"
)

func (d *BuntDb) InsertSuite(_ context.Context, s *UnsavedSuite) (string, error) {
	return d.insert(CollSuites, s)
}

func (d *BuntDb) Suite(_ context.Context, id string) (*Suite, error) {
	var s Suite
	if err := d.find(CollSuites, id, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func (d *BuntDb) SuitePage(_ context.Context, fromId string, limit int) (*SuitePage, error) {
	page := SuitePage{
		Suites: make([]Suite, 0),
	}
	itr := func(k, v string) bool {
		if gjson.Get(v, "deleted").Bool() {
			return false
		}
		if len(page.Suites) == limit {
			id := gjson.Get(v, "id").String()
			page.NextId = &id
			return false
		}
		var s Suite
		if err := json.Unmarshal([]byte(v), &s); err != nil {
			log.Panicf("unmarshal json: %v\n", err)
		}
		page.Suites = append(page.Suites, s)
		return true
	}
	err := d.db.View(func(tx *buntdb.Tx) error {
		if fromId == "" {
			err := tx.DescendLessOrEqual("suites_deleted",
				`{"deleted":false}`, itr)
			if err != nil {
				return err
			}
		} else {
			v, err := tx.Get(buntDbKey(CollSuites, fromId))
			if err != nil {
				return err
			}
			startedAt := gjson.Get(v, "started_at").String()
			fromId := strconv.Quote(fromId)
			err = tx.DescendLessOrEqual("suites_deleted",
				`{"deleted":false,"started_at":`+startedAt+`,"id":`+fromId+`}`, itr)
			if err != nil {
				return err
			}
		}
		return tx.Ascend("suites_status", func(k, v string) bool {
			if gjson.Get(v, "status").String() == string(SuiteStatusRunning) {
				page.RunningCount++
			} else {
				page.FinishedCount++
			}
			return true
		})
	})
	if err != nil {
		return nil, err
	}
	return &page, nil
}

func (d *BuntDb) UpdateSuiteStatus(_ context.Context, id string, status SuiteStatus, at int64) error {
	m := map[string]interface{}{
		"status": status,
	}
	switch status {
	case SuiteStatusRunning:
	case SuiteStatusDisconnected:
		m["disconnected_at"] = at
	default:
		m["finished_at"] = at
	}
	return d.update(CollSuites, id, m)
}

func (d *BuntDb) ReconnectSuite(_ context.Context, id string, at int64, expiry time.Duration) error {
	return d.funcUpdate(CollSuites, id, func(v string) (map[string]interface{}, error) {
		if gjson.Get(v, "status").Value() != SuiteStatusDisconnected {
			return nil, ErrNotReconnectable
		}
		if at-gjson.Get(v, "disconnected_at").Int() > expiry.Milliseconds() {
			return nil, ErrExpired
		}
		return map[string]interface{}{
			"status":          SuiteStatusRunning,
			"disconnected_at": 0,
		}, nil
	})
}

func (d *BuntDb) DeleteSuite(_ context.Context, id string, at int64) error {
	return d.delete(CollSuites, id, at)
}

func (d *BuntDb) DeleteAllSuites(_ context.Context, at int64) error {
	return d.deleteAll(CollSuites, "suites_deleted", at)
}
