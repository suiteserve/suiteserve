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

	// min int
	// id  string
	// pad int
	//
	// min0 entry
	// id0  entry
	// pad0 entry
	//
	// rightN int
	// right0 entry
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

// func newLteId1(tx *buntdb.Tx, less less, id string) (func(v string) bool,
// 	error) {
// 	if id == "" {
// 		return func(v string) bool {
// 			return true
// 		}, nil
// 	}
// 	id1Val, err := tx.Get(key(SuiteColl, id))
// 	if err != nil {
// 		return nil, err
// 	}
// 	return func(v string) bool {
// 		return !less(id1Val, v)
// 	}, nil
// }
//
// func newLtPad1(tx *buntdb.Tx, less less, id string,
// 	pad int) (func(v string) bool, error) {
// 	if id == "" {
// 		return func(v string) bool {
// 			return true
// 		}, nil
// 	}
// 	id1Val, err := tx.Get(key(SuiteColl, id))
// 	if err != nil {
// 		return nil, err
// 	}
// 	pad1Val := id1Val
// 	err = tx.DescendLessOrEqual(suiteIndexStartedAt,
// 		suiteIndexStartedAtPivot(id1Val),
// 		newAndCond(func(k, v string) bool {
// 			pad1Val = v
// 			return true
// 		}, newLimitCond(pad)))
// 	if err != nil {
// 		return nil, err
// 	}
// 	return func(v string) bool {
// 		return less(v, pad1Val)
// 	}, nil
// }
//
// func (w *SuiteWatcher) newLtPad0(less less) func(v string) bool {
// 	if w.id == "" {
// 		return func(v string) bool {
// 			return true
// 		}
// 	}
// 	return func(v string) bool {
// 		return less(v, w.pad0.v)
// 	}
// }
//
// // newIdPad1Itr returns a new iterator that returns false while outside and true
// // while inside, inclusive, the idPad1 region. Must be called in descending
// // order and the first element must be at or before a non-empty id1.
// func (w *SuiteWatcher) newIdPad1Itr(id string, pad int) itr {
// 	var fn itr
// 	right := func(k, v string) bool {
// 		return false
// 	}
// 	var i int
// 	inside := func(k, v string) bool {
// 		if i == pad {
// 			fn = right
// 			return right(k, v)
// 		}
// 		i++
// 		return true
// 	}
// 	left := func(k, v string) bool {
// 		if k == id {
// 			fn = inside
// 			return inside(k, v)
// 		}
// 		return false
// 	}
// 	fn = left
// 	return func(k, v string) bool {
// 		return fn(k, v)
// 	}
// }
//
// func (w *SuiteWatcher) itr(tx *buntdb.Tx, min int, id string, pad int,
// 	less less, add, rm consumer) error {
// 	ltPad0 := w.newLtPad0(less)
// 	ltPad1, err := newLtPad1(tx, less, id, pad)
// 	if err != nil {
// 		return err
// 	}
//
// 	min1 := w.min0
// 	var rm1, add1 entry
// 	var leftFn, rightFn consumer
// 	rmFn := func(k, v string) {
// 		min1 = entry{k, v}
// 		if ltPad1(v) {
// 			rm(k, v)
// 		}
// 	}
// 	leftFn = func(k, v string) {
// 		// moving left, ascending, remove
// 		min1 = entry{k, v}
// 		if ltPad1(v) {
// 			rm1 = entry{k, v}
// 			leftFn = rmFn
// 			rm(k, v)
// 		}
// 	}
// 	addFn := func(k, v string) {
// 		min1 = entry{k, v}
// 		add(k, v)
// 	}
// 	rightFn = func(k, v string) {
// 		// moving right, descending, add
// 		min1 = entry{k, v}
// 		if ltPad0(v) {
// 			add1 = entry{k, v}
// 			rightFn = addFn
// 			add(k, v)
// 		}
// 	}
// 	err = w.itrMin(tx, min, func(k, v string) {
// 		leftFn(k, v)
// 	}, func(k, v string) {
// 		rightFn(k, v)
// 	})
// 	if err != nil {
// 		return err
// 	}
// }
//
// func (w *SuiteWatcher) itrIdPad(tx *buntdb.Tx, id string, pad int, less less,
// 	add, rm itr) error {
// 	/*
// 		Example 1:
// 		      |-- 0 --|
// 		0 1 2 3 4 5 6 7 8 9
// 		  |-- 1 --|
//
// 		[1,2]: add
// 		[3,5]: skip (overlap)
// 		[6,7]: rm
//
// 		Example 2:
// 		            |- 0 -|
// 		0 1 2 3 4 5 6 7 8 9
// 		  |- 1 -|
//
// 		[1,4]: add
// 		[6,9]: rm
//
// 		Example 3:
// 		    |-- 0 --|
// 		0 1 2 3 4 5 6 7 8 9
// 		  |----- 1 -----|
//
// 		[1,1]: add
// 		[2,6]: skip (overlap)
// 		[7,8]: add
//
// 		Example 4:
// 		    |-- 0 --|
// 		0 1 2 3 4 5 6 7 8 9
// 		        |-- 1 --|
//
// 		[2,3]: rm
// 		[4,6]: skip (overlap)
// 		[7,8]: add
// 	*/
// 	if id == w.id && pad == w.pad {
// 		// no change
// 		return nil
// 	}
// 	if w.id != "" && id == "" {
// 		// newly empty
// 		return idPadRm(tx, w.id0.v, w.pad0.v, less, rm)
// 	}
// 	id1, err := getSuiteIdEntry(tx, id)
// 	if err != nil {
// 		return err
// 	}
// 	if w.id == "" && id != "" {
// 		// newly non-empty
// 		return idPadAdd(tx, id1.v, pad, add)
// 	}
// 	// new id and/or pad
// 	if less(id1.v, w.pad0.v) {
// 		// no overlap: remove old region and add new region
// 		if err := idPadRm(tx, w.id0.v, w.pad0.v, less, rm); err != nil {
// 			return err
// 		}
// 		return idPadAdd(tx, id1.v, pad, add)
// 	}
// 	n := -1
// 	var fn itr
// 	afterFn := func() error {
// 		return nil
// 	}
// 	// addRest adds each element after the end of the old region and up to the
// 	// end of the new region. Precondition: pad0 > v >= pad1, i.e. we're already
// 	// past the end of the old region but not yet past the end of the new
// 	// region.
// 	addRest := func(k, v string) bool {
// 		if n == pad {
// 			// we're past the end of the new region
// 			return false
// 		}
// 		n++
// 		return add(k, v)
// 	}
// 	// rmRest removes each element after the end of the new region and up to the
// 	// end of the old region. Precondition: pad1 > v >= pad0, i.e. we're already
// 	// past the end of the new region but not yet past the end of the old
// 	// region.
// 	rmRest := func(k, v string) bool {
// 		if less(v, w.pad0.v) {
// 			// we're past the end of the old region
// 			return false
// 		}
// 		return rm(k, v)
// 	}
// 	// skipOverlap skips each element at or after the start of the old region
// 	// and before the end of the new region and before the end of the old
// 	// region, whichever comes first. If the end of the old region comes first,
// 	// then the rest of the new region is added according to addRest. If the end
// 	// of the new region comes first, then the rest of the old region is removed
// 	// according to rmRest. If the end of both regions is equal, then rmRest
// 	// will be called by default, but nothing more will be added or removed.
// 	// Precondition: id0 >= v >= pad0 && v >= pad1, i.e. we're somewhere within
// 	// the old region, inclusive, but not yet past the end of the new region.
// 	skipOverlap := func(k, v string) bool {
// 		if n == pad {
// 			// we're past the end of the new region
// 			fn = rmRest
// 			return rmRest(k, v)
// 		}
// 		if less(v, w.pad0.v) {
// 			// we're past the end of the old region
// 			fn = addRest
// 			return addRest(k, v)
// 		}
// 		n++
// 		return true
// 	}
// 	// addFirst adds each element at or after the start of the new region and
// 	// before the end of the new region and before the start of the old region,
// 	// whichever comes first. If the start of the old region comes first, then
// 	// the iterator morphs into skipOverlap. If the end of the new region comes
// 	// first, then the iterator stops and afterFn is set to remove the old
// 	// region. Precondition: id1 >= v >= pad1 && v > id0, i.e. we're somewhere
// 	// within the new region, inclusive, but not yet at or past the start of the
// 	// old region.
// 	addFirst := func(k, v string) bool {
// 		if n == pad {
// 			// we're past the end of the new region
// 			afterFn = func() error {
// 				// remove old region
// 				return idPadRm(tx, w.id0.v, w.pad0.v, less, rm)
// 			}
// 			return false
// 		}
// 		if !less(w.id0.v, v) {
// 			// we're at the start of the old region
// 			fn = skipOverlap
// 			return skipOverlap(k, v)
// 		}
// 		n++
// 		return add(k, v)
// 	}
// 	leftVal := id1.v
// 	if less(w.id0.v, id1.v) {
// 		// the new region starts before the old region
// 		fn = addFirst
// 	} else {
// 		// the old region starts at or before the new region
// 		leftVal = w.id0.v
// 		// we first need to remove the elements before the start of the new
// 		// region
// 		fn = func(k, v string) bool {
// 			if !less(id1.v, v) {
// 				// we're at or past the start of the new region
// 				fn = addFirst
// 				return addFirst(k, v)
// 			}
// 			// we're before the start of the new region
// 			return rm(k, v)
// 		}
// 	}
// 	err = tx.DescendLessOrEqual(suiteIndexStartedAt,
// 		suiteIndexStartedAtPivot(leftVal),
// 		func(k, v string) bool {
// 			return fn(k, v)
// 		})
// 	if err != nil {
// 		return err
// 	}
// 	return afterFn()
// }
//
// func getSuiteIdEntry(tx *buntdb.Tx, id string) (entry, error) {
// 	e := entry{k: key(SuiteColl, id)}
// 	var err error
// 	e.v, err = tx.Get(e.k)
// 	return e, err
// }
//
// func idPadAdd(tx *buntdb.Tx, leftVal string, pad int, add itr) error {
// 	return tx.DescendLessOrEqual(suiteIndexStartedAt,
// 		suiteIndexStartedAtPivot(leftVal),
// 		newAndCond(add, newLimitCond(pad)))
// }
//
// func idPadRm(tx *buntdb.Tx, leftVal, rightVal string, less less, rm itr) error {
// 	leftPivot := suiteIndexStartedAtPivot(leftVal)
// 	rightPivot := suiteIndexStartedAtPivot(rightVal)
// 	return tx.DescendLessOrEqual(suiteIndexStartedAt, leftPivot,
// 		newAndCond(newGreaterOrEqual(rightPivot, less), rm))
// }
//
// func (w *SuiteWatcher) itrMin(tx *buntdb.Tx, min int,
// 	leftItr, rightItr consumer) error {
// 	if min == w.min {
// 		// no change
// 		return nil
// 	}
// 	if min > w.min {
// 		// new min > 0
// 		itr := newAndCond(newLimitCond(min-w.min), func(k, v string) bool {
// 			rightItr(k, v)
// 			return true
// 		})
// 		if w.min == 0 {
// 			return tx.Descend(suiteIndexStartedAt, itr)
// 		}
// 		return tx.DescendLessOrEqual(suiteIndexStartedAt,
// 			suiteIndexStartedAtPivot(w.min0.v),
// 			newSkipKeyCond(w.min0.k, itr))
// 	}
// 	// old min > 0
// 	return tx.AscendGreaterOrEqual(suiteIndexStartedAt,
// 		suiteIndexStartedAtPivot(w.min0.v),
// 		newAndCond(newLimitCond(w.min-min), func(k, v string) bool {
// 			leftItr(k, v)
// 			return true
// 		}))
// }

func (w *SuiteWatcher) handleChanges(changes []Change) {

}
