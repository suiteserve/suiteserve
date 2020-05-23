package repo

import (
	"github.com/tidwall/buntdb"
	"github.com/tidwall/gjson"
)

type buntCaseRepo struct {
	*buntRepo
}

func IndexOptionalJSON(path string) func(a, b string) bool {
	return func(a, b string) bool {
		aResult := gjson.Get(a, path)
		bResult := gjson.Get(b, path)
		if aResult.Exists() && bResult.Exists() {
			return aResult.Less(bResult, false)
		}
		return false
	}
}

func (r *buntRepo) newCaseRepo() (*buntCaseRepo, error) {
	err := r.db.ReplaceIndex("cases_suite", "cases:*",
		buntdb.IndexJSON("suite"),
		buntdb.Desc(IndexOptionalJSON("num")),
		buntdb.Desc(IndexOptionalJSON("created_at")))
	if err != nil {
		return nil, err
	}
	return &buntCaseRepo{r}, nil
}

func (r *buntCaseRepo) Save(c Case) (string, error) {
	return r.save(&c, CaseCollection)
}

func (r *buntCaseRepo) SaveAttachment(id string, attachmentId string) error {
	return r.set(CaseCollection, id, map[string]interface{}{
		"attachments.-1": attachmentId,
	})
}

func (r *buntCaseRepo) SaveStatus(id string, status CaseStatus, opts *CaseRepoSaveStatusOptions) error {
	return r.set(CaseCollection, id, map[string]interface{}{
		"status":      status,
		"flaky":       opts.flaky,
		"started_at":  opts.startedAt,
		"finished_at": opts.finishedAt,
	})
}

func (r *buntCaseRepo) Find(id string) (*Case, error) {
	var c Case
	if err := r.find(CaseCollection, id, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *buntCaseRepo) FindAllBySuite(suiteId string, num *int64) ([]Case, error) {
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
