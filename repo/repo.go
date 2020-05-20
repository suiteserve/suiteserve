package repo

import (
	"github.com/tidwall/buntdb"
	"github.com/tmazeika/testpass/config"
	"time"
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

func nowTimeMillis() int64 {
	now := time.Now()
	// Doesn't use now.UnixNano() to avoid Y2K262.
	return now.Unix()*time.Second.Milliseconds() +
		time.Duration(now.Nanosecond()).Milliseconds()
}
