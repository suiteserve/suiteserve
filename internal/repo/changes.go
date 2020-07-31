package repo

import (
	"encoding/json"
	"github.com/tidwall/buntdb"
)

type Mask []string

type Change interface {
	isChange()
}

type SuiteAggUpdate struct {
	SuiteAgg
}

func (SuiteAggUpdate) isChange() {}

type SuiteUpsert struct {
	Suite
	Mask
}

func (SuiteUpsert) isChange() {}

type SuiteDelete struct {
	Id string
}

func (SuiteDelete) isChange() {}

type changeHandler interface {
	handleChanges(changes []Change)
}

type watcher struct {
	in  chan<- []Change
	out <-chan []Change
}

func newWatcher() *watcher {
	in := make(chan []Change)
	out := make(chan []Change)
	var buf [][]Change
	getNext := func() []Change {
		if len(buf) == 0 {
			return nil
		}
		return buf[0]
	}
	getOut := func() chan<- []Change {
		if len(buf) == 0 {
			return nil
		}
		return out
	}
	go func() {
		defer close(out)
		for {
			select {
			case changes, ok := <-in:
				if !ok {
					return
				}
				buf = append(buf, changes)
			case getOut() <- getNext():
				buf = buf[1:]
			}
		}
	}()
	return &watcher{
		in:  in,
		out: out,
	}
}

type SuiteWatcher struct {
	r *Repo
	*watcher

	id    string
	padLt int
	padGt int

	n   int
	min entry
	max entry
}

func (r *Repo) WatchSuites() *SuiteWatcher {
	w := SuiteWatcher{
		r:       r,
		watcher: newWatcher(),
	}
	r.addHandler(&w)
	return &w
}

func (w *SuiteWatcher) Close() {
	w.r.removeHandler(w)
	close(w.in)
}

func (w *SuiteWatcher) Changes() <-chan []Change {
	return w.out
}

func (w *SuiteWatcher) SetQuery(id string, padLt, padGt int) error {
	return w.r.db.View(func(tx *buntdb.Tx) error {
		var changes []Change
		add := func(k, v string) bool {
			var upsert SuiteUpsert
			if err := json.Unmarshal([]byte(v), &upsert.Suite); err != nil {
				panic(err)
			}
			changes = append(changes, upsert)
			return true
		}
		rm := func(k, v string) bool {
			changes = append(changes, SuiteDelete{Id: getId(k)})
			return true
		}
		// rm old region
		if w.id != "" || w.padGt > 0 || w.padLt > 0 {
			err := tx.AscendGreaterOrEqual(suiteIndexStartedAt,
				suiteIndexStartedAtPivot(w.min.v),
				newAndCond(rm, newUntilKeyCond(w.max.k)))
			if err != nil {
				return err
			}
		}
		var n int
		// add new region
		lt := func(k, v string) bool {
			w.min = entry{k, v}
			n++
			return add(k, v)
		}
		eq := func(k, v string) bool {
			w.min = entry{k, v}
			w.max = entry{k, v}
			n++
			return add(k, v)
		}
		gt := func(k, v string) bool {
			w.max = entry{k, v}
			n++
			return add(k, v)
		}
		if err := itrAroundSuite(tx, id, padLt, padGt, lt, eq, gt); err != nil {
			return err
		}
		w.id = id
		w.padLt = padLt
		w.padGt = padGt
		w.n = n

		var agg SuiteAgg
		err := w.r.getById(SuiteAggColl, "", &agg)
		if err != nil && err != ErrNotFound {
			return err
		}
		w.in <- append(changes, SuiteAggUpdate{agg})
		return nil
	})
}

func itrAroundSuite(tx *buntdb.Tx, id string, padLt, padGt int,
	lt, eq, gt itr) error {
	firstEq := newFirstCond(eq)
	restLt := newRestCond(lt)
	restGt := newRestCond(gt)
	if id == "" {
		return tx.Descend(suiteIndexStartedAt,
			newAndCond(firstEq, restLt, newLimitCond(padGt+padLt)))
	}
	idVal, err := tx.Get(key(SuiteColl, id))
	if err != nil {
		return err
	}
	pivot := suiteIndexStartedAtPivot(idVal)
	err = tx.AscendGreaterOrEqual(suiteIndexStartedAt, pivot,
		newAndCond(firstEq, restGt, newLimitCond(padGt)))
	if err != nil {
		return err
	}
	if padLt == 0 {
		return nil
	}
	return tx.DescendLessOrEqual(suiteIndexStartedAt, pivot,
		newAndCond(restLt, newLimitCond(padLt)))
}

func (w *SuiteWatcher) handleChanges(changes []Change) {

}
