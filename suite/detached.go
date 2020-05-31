package suite

import (
	"sync"
	"time"
)

const sweepPeriod = 30 * time.Second

type detachedSuiteStore struct {
	sync.RWMutex
	ttl time.Duration
	exp map[string]time.Time
}

func newDetachedSuiteStore(ttl time.Duration, done <-chan interface{}) *detachedSuiteStore {
	s := detachedSuiteStore{
		ttl: ttl,
		exp: make(map[string]time.Time),
	}
	go s.sweepPeriodically(done)
	return &s
}

func (s *detachedSuiteStore) detach(id string) {
	now := time.Now()
	s.Lock()
	defer s.Unlock()
	s.exp[id] = now.Add(s.ttl)
}

func (s *detachedSuiteStore) reattach(id string) bool {
	now := time.Now()
	s.RLock()
	defer s.RUnlock()
	exp, ok := s.exp[id]
	if !ok || !now.Before(exp) {
		return false
	}
	delete(s.exp, id)
	return true
}

func (s *detachedSuiteStore) sweepPeriodically(done <-chan interface{}) {
	ticker := time.NewTicker(sweepPeriod)
	defer ticker.Stop()
	for {
		select {
		case tick := <-ticker.C:
			s.sweep(tick)
		case <-done:
			return
		}
	}
}

func (s *detachedSuiteStore) sweep(now time.Time) {
	s.Lock()
	defer s.Unlock()
	for id, exp := range s.exp {
		if !now.Before(exp) {
			delete(s.exp, id)
		}
	}
}
