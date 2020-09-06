package repo

import (
	"encoding/json"
	"github.com/asdine/storm/v3"
)

type SuiteStatus string

const (
	SuiteStatusStarted      SuiteStatus = "started"
	SuiteStatusFinished     SuiteStatus = "finished"
	SuiteStatusDisconnected SuiteStatus = "disconnected"
)

type SuiteResult string

const (
	SuiteResultPassed SuiteResult = "passed"
	SuiteResultFailed SuiteResult = "failed"
)

type Suite struct {
	Entity
	VersionedEntity
	SoftDeleteEntity
	Name             string      `json:"name,omitempty"`
	Tags             []string    `json:"tags,omitempty"`
	PlannedCases     int64       `json:"planned_cases,omitempty"`
	Status           SuiteStatus `json:"status,omitempty"`
	Result           SuiteResult `json:"result,omitempty"`
	DisconnectedAt   int64       `json:"disconnected_at,omitempty"`
	StartedAt        int64       `json:"started_at,omitempty"`
	FinishedAt       int64       `json:"finished_at,omitempty"`
}

type SuiteAgg struct {
	Entity          `storm:"inline"`
	VersionedEntity `storm:"inline"`
	TotalCount      int64 `json:"total_count"`
	StartedCount    int64 `json:"started_count"`
}

func (a *SuiteAgg) applyOne(s *Suite) {
	a.Version++
	if !s.Deleted {
		a.TotalCount++
		if s.Status == SuiteStatusStarted {
			a.StartedCount++
		}
	}
}

func (a *SuiteAgg) applyDiff(s1, s0 *Suite) {
	a.Version++
	if s1.Deleted != s0.Deleted {
		if s1.Deleted {
			// newly deleted
			a.TotalCount--
			if s0.Status == SuiteStatusStarted {
				// was started
				a.StartedCount--
			}
		} else {
			// no longer deleted
			a.TotalCount++
			if s1.Status == SuiteStatusStarted {
				// newly started
				a.StartedCount++
			}
		}
	} else if s1.Status != s0.Status {
		if s1.Status == SuiteStatusStarted {
			// newly started
			a.StartedCount++
		} else {
			// no longer started
			a.StartedCount--
		}
	}
}

func (r *Repo) InsertSuite(s Suite) (string, error) {
	// err := r.update(func(tx storm.Node) error {
	// 	if err := tx.Save(&s); err != nil {
	// 		return err
	// 	}
	// 	var agg SuiteAgg
	// 	err := updateAgg(tx, suiteAggKey, &agg, func() {
	// 		agg.applyOne(&s)
	// 	})
	// 	if err != nil {
	// 		return err
	// 	}
	// 	return r.cb.publish(commit{
	// 		tx: tx,
	// 		cl: ChangeList{&SuiteUpsert{&s, nil}, (*SuiteAggUpsert)(&agg)},
	// 	})
	// })
	// return s.Id, err
	return "", nil
}

func (r *Repo) Suite(id string) (s Suite, err error) {
	// err = wrapNotFoundErr(r.db.One("Id", id, &s))
	return
}

func (r *Repo) UpdateSuite(id string, mask json.RawMessage) (s Suite, err error) {
	// err = r.update(func(tx storm.Node) error {
	// 	if err := tx.One("Id", id, &s); err != nil {
	// 		return wrapNotFoundErr(err)
	// 	}
	// 	s0 := s
	// 	if err := json.Unmarshal(mask, &s); err != nil {
	// 		return errBadJson{err}
	// 	}
	// 	if err := tx.Update(&s); err != nil {
	// 		return err
	// 	}
	// 	var agg SuiteAgg
	// 	err := updateAgg(tx, suiteAggKey, &agg, func() {
	// 		agg.applyDiff(&s, &s0)
	// 	})
	// 	if err != nil {
	// 		return err
	// 	}
	// 	return r.cb.publish(commit{
	// 		tx: tx,
	// 		cl: ChangeList{&SuiteUpsert{&s, &s0}, (*SuiteAggUpsert)(&agg)},
	// 	})
	// })
	return
}

func (r *Repo) WatchSuites(done chan chan struct{}, id string,
	padLt, padGt int) (<-chan ChangeList, error) {
	const indexF = "StartedAt"
	// var pvt Suite
	// var all, lt, eq, gt []Suite
	bw := newBufWatcher()
	var w *suiteWatcher
	// err := r.view(func(tx storm.Node) error {
	// 	if err := tx.One("Id", id, &pvt); err != nil {
	// 		return wrapNotFoundErr(err)
	// 	}
	// 	if err := tx.Select(q.Lt(indexF, pvt.StartedAt)).
	// 		OrderBy(indexF, "Id").Limit(padLt).Find(&lt); err != nil {
	// 		return err
	// 	}
	// 	if err := tx.Select(q.Eq(indexF, pvt.StartedAt)).
	// 		OrderBy(indexF, "Id").Find(&eq); err != nil {
	// 		return err
	// 	}
	// 	if err := tx.Select(q.Gt(indexF, pvt.StartedAt)).
	// 		OrderBy(indexF, "Id").Limit(padGt).Find(&gt); err != nil {
	// 		return err
	// 	}
	// 	all = append(lt, append(eq, gt...)...)
	// 	w = &suiteWatcher{
	// 		r:     r,
	// 		out:   bw.in,
	// 		min:   &all[0],
	// 		pvt:   &pvt,
	// 		max:   &all[len(all)-1],
	// 		padLt: padLt,
	// 		padGt: padGt,
	// 		ltN:   len(lt),
	// 		eqN:   len(eq),
	// 		gtN:   len(gt),
	// 	}
	// 	w.sendFirst(all)
	// 	r.cb.watch(w)
	// 	return nil
	// })
	// if err != nil {
	// 	if w != nil {
	// 		r.cb.unwatch(w)
	// 	}
	// 	close(bw.in)
	// 	return nil, err
	// }
	go func() {
		ret := <-done
		r.cb.unwatch(w)
		close(bw.in)
		ret <- struct{}{}
	}()
	return bw.out, nil
}

type suiteWatcher struct {
	r     *Repo
	out   chan<- ChangeList
	min   *Suite
	pvt   *Suite
	max   *Suite
	padLt int
	padGt int
	ltN   int
	eqN   int
	gtN   int
}

func (w *suiteWatcher) sendFirst(all []Suite) {
	cl := make(ChangeList, len(all))
	for i := range all {
		cl[i] = &SuiteUpsert{&all[i], nil}
	}
	w.out <- cl
}

func (w *suiteWatcher) belongsLt(tx storm.Node, s Suite) (bool, error) {
	const indexF = "StartedAt"
	// ltN, err := tx.Select(q.Eq("Id", s.Id), q.Lt(indexF, w.pvt.StartedAt)).
	// 	Count(&Suite{})
	// if err != nil || ltN == 0 {
	// 	return false, err
	// }
	// if w.min != nil {
	//
	// }
	return false, nil
}

func (w *suiteWatcher) onCommit(c commit) error {
	// tx := c.tx
	// var cl ChangeList
	// for _, ch := range c.cl {
	// 	switch ch := ch.(type) {
	// 	case *SuiteUpsert:
	// 		if ch.s0 == nil {
	// 			// insert
	//
	// 		} else {
	// 			// update
	// 		}
	// 	case *SuiteAggUpsert:
	// 		cl = append(cl, ch)
	// 	}
	// }
	return nil
}
