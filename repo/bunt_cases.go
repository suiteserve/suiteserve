package repo

import "github.com/tidwall/buntdb"

type buntCaseRepo struct {
	*buntRepo
}

func (r *buntRepo) newCaseRepo() (*buntCaseRepo, error) {
	err := r.db.ReplaceIndex("cases_suite", "cases:*",
		buntdb.IndexJSON("suite"))
	if err != nil {
		return nil, err
	}
	err = r.db.ReplaceIndex("cases_suite_num", "cases:*",
		buntdb.IndexJSON("suite"), buntdb.IndexJSON("num"))
	if err != nil {
		return nil, err
	}
	err = r.db.ReplaceIndex("cases_deleted", "cases:*",
		buntdb.IndexJSON("deleted"))
	if err != nil {
		return nil, err
	}
	return &buntCaseRepo{r}, nil
}

func (r *buntCaseRepo) Save(c Case) (string, error) {
	return r.save(&c, CaseCollection)
}

func (r *buntCaseRepo) SaveAttachments(id string, attachments ...string) error {
	m := make(map[string]interface{})
	for _, a := range attachments {
		m["attachments.-1"] = a
	}
	return r.set(CaseCollection, id, m)
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
	if err := r.find(CaseCollection, id, Case{}); err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *buntCaseRepo) FindAllBySuite(suiteId string, num *int64) ([]Case, error) {
	m := map[string]interface{}{
		"suite": suiteId,
	}
	index := "cases_suite"
	if num != nil {
		m["num"] = *num
		index = "cases_suite_num"
	}
	var cases []Case
	if err := r.findAllBy(index, m, &cases); err != nil {
		return nil, err
	}
	return cases, nil
}
