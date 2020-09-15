package repo

import (
	"github.com/asdine/storm/v3"
	"sync"
)

type Change interface {
	Type() string
	Data() []byte
}

type ChangeList []Change

type commit struct{
	tx storm.Node
	cl ChangeList
}

type bufWatcher struct {
	in  chan<- ChangeList
	out <-chan ChangeList
}

func newBufWatcher() bufWatcher {
	in := make(chan ChangeList)
	out := make(chan ChangeList)
	var buf []ChangeList
	getNext := func() ChangeList {
		if len(buf) == 0 {
			return nil
		}
		return buf[0]
	}
	getOut := func() chan<- ChangeList {
		if len(buf) == 0 {
			return nil
		}
		return out
	}
	go func() {
		defer close(out)
		for {
			select {
			case v, ok := <-in:
				if !ok {
					return
				}
				buf = append(buf, v)
			case getOut() <- getNext():
				buf = buf[1:]
			}
		}
	}()
	return bufWatcher{in, out}
}

type watcher interface {
	onCommit(c commit) error
}

type changeBroker struct {
	mu       sync.Mutex
	watchers []watcher
}

func (b *changeBroker) publish(c commit) (err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, w := range b.watchers {
		if err2 := w.onCommit(c); err2 != nil && err == nil {
			err = err2
		}
	}
	return
}

func (b *changeBroker) watch(w watcher) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.watchers = append(b.watchers, w)
}

func (b *changeBroker) unwatch(w watcher) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for i, w2 := range b.watchers {
		if w2 == w {
			b.watchers = append(b.watchers[:i], b.watchers[i+1:]...)
			return
		}
	}
}

type SuiteUpsert struct {
	s1 *Suite
	s0 *Suite
}

// func (SuiteUpsert) Type() string {
// 	return "suite_upsert"
// }
//
// func (u *SuiteUpsert) Data() []byte {
// 	return mustMarshalJson(u.s1)
// }

type SuiteDelete string

// func (SuiteDelete) Type() string {
// 	return "suite_delete"
// }
//
// func (d *SuiteDelete) Data() []byte {
// 	return mustMarshalJson(d)
// }

// type SuiteAggUpsert SuiteAgg
//
// func (SuiteAggUpsert) Type() string {
// 	return "suite_agg_upsert"
// }
//
// func (u *SuiteAggUpsert) Data() []byte {
// 	return mustMarshalJson(u)
// }
