package repo

import (
	"context"
)

func (d *BuntDb) InsertCase(_ context.Context, c *UnsavedCase) (string, error) {
	return d.insert(CollCases, c)
}

func (d *BuntDb) Case(_ context.Context, id string) (*Case, error) {
	var c Case
	if err := d.find(CollCases, id, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func (d *BuntDb) CasesBySuite(_ context.Context, suiteId string) ([]Case, error) {
	var cases []Case
	err := d.findAllBy("cases_suite", map[string]interface{}{
		"suite": suiteId,
	}, &cases)
	if err != nil {
		return nil, err
	}
	return cases, nil
}

func (d *BuntDb) UpdateCaseStatus(_ context.Context, id string, status CaseStatus, at int64) error {
	m := map[string]interface{}{
		"status": status,
	}
	if status == CaseStatusRunning {
		m["started_at"] = at
	} else {
		m["finished_at"] = at
	}
	return d.update(CollCases, id, m)
}
