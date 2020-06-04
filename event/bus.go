package event

import "sync"

type Bus struct {
	mu   sync.RWMutex
	subs []*Subscriber
}

func (b *Bus) Subscribe() *Subscriber {
	sub := newSubscriber()
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subs = append(b.subs, sub)
	return sub
}

func (b *Bus) Publish(e Event) {
	b.mu.Lock()
	defer b.mu.Unlock()

	remainingSubs := b.subs[:0]
	for _, s := range b.subs {
		select {
		case s.ch <- e:
			remainingSubs = append(remainingSubs, s)
		case <-s.done:
			close(s.ch)
		}
	}
	b.subs = remainingSubs
}
