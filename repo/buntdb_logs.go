package repo

import (
	"context"
	"encoding/json"
	"github.com/tidwall/buntdb"
	"github.com/tidwall/gjson"
	"log"
	"strconv"
)

func (d *BuntDb) InsertLogLine(_ context.Context, l *UnsavedLogLine) (string, error) {
	return d.insert(CollLogs, l)
}

func (d *BuntDb) LogPage(_ context.Context, caseId, fromId string, limit int) (*LogPage, error) {
	var page LogPage
	itr := func(k, v string) bool {
		if gjson.Get(v, "case").String() != caseId {
			return false
		}
		if len(page.Lines) == limit {
			id := gjson.Get(v, "id").String()
			page.NextId = &id
			return false
		}
		var l LogLine
		if err := json.Unmarshal([]byte(v), &l); err != nil {
			log.Panicf("unmarshal json: %v\n", err)
		}
		page.Lines = append(page.Lines, l)
		return true
	}
	err := d.db.View(func(tx *buntdb.Tx) error {
		v, err := tx.Get(buntDbKey(CollLogs, fromId))
		if err != nil {
			return err
		}
		timestamp := gjson.Get(v, "timestamp").String()
		seq := gjson.Get(v, "seq").String()
		return tx.DescendLessOrEqual("logs_case",
			`{"case":`+strconv.Quote(caseId)+`,"timestamp":`+timestamp+`,`+
				`"seq":`+seq+`,"id":`+strconv.Quote(fromId)+`}`, itr)
	})
	if err == buntdb.ErrNotFound {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}
	return &page, nil
}
