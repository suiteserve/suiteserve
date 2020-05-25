package repo

import (
	"context"
	"github.com/tidwall/buntdb"
)

type buntCaseRepo struct {
	*buntRepo
}

func (r *buntRepo) newCaseRepo() (*buntCaseRepo, error) {
	err := r.db.ReplaceIndex("cases_suite", "cases:*",
		buntdb.IndexJSON("suite"),
		buntdb.Desc(indexJSONOptional("num")),
		buntdb.Desc(indexJSONOptional("created_at")),
		buntdb.Desc(indexJSONOptional("id")))
	if err != nil {
		return nil, err
	}
	return &buntCaseRepo{r}, nil
}

func (r *buntCaseRepo) Save(_ context.Context, c Case) (string, error) {
	return r.save(CaseCollection, &c)
}

func (r *buntCaseRepo) SaveAttachment(_ context.Context, id string, attachmentId string) error {
	return r.set(CaseCollection, id, map[string]interface{}{
		"attachments.-1": attachmentId,
	})
}

func (r *buntCaseRepo) SaveStatus(_ context.Context, id string, status CaseStatus, opts *CaseRepoSaveStatusOptions) error {
	return r.set(CaseCollection, id, map[string]interface{}{
		"status":      status,
		"flaky":       opts.flaky,
		"started_at":  opts.startedAt,
		"finished_at": opts.finishedAt,
	})
}

func (r *buntCaseRepo) Find(_ context.Context, id string) (*Case, error) {
	var c Case
	if err := r.find(CaseCollection, id, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *buntCaseRepo) FindAllBySuite(_ context.Context, suiteId string, num *int64) ([]Case, error) {
	m := map[string]interface{}{
		"suite": suiteId,
	}
	if num != nil {
		m["num"] = *num
	}
	var cases []Case
	if err := r.findAllBy("cases_suite", m, &cases); err != nil {
		return nil, err
	}
	return cases, nil
}
