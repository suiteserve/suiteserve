package event

import "sync"

// A Bus is a publish/subscribe method of transportation for Events of any type.
// Many concurrent subscribers can subscribe to events published to the Bus.
// Subscribers will not receive events published prior to the time of
// subscription, and slow receivers will block the publishers. This guarantees
// that all subscribers receive events in the order in which they were
// published.
type Bus struct {
	mu   sync.RWMutex
	subs []*Subscriber
}

// Subscribe creates a new Subscriber that receives new events published to the
// Bus. When the subscription is done, the returned Subscriber's Unsubscribe
// method should be called.
func (b *Bus) Subscribe() *Subscriber {
	sub := newSubscriber()
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subs = append(b.subs, sub)
	return sub
}

// A Publisher wraps a Bus and adds event publishing capability. Its purpose is
// to separate publishing from subscribing; the contained Bus, which only allows
// subscribing, may be handed out more liberally than a Publisher that allows
// both.
type Publisher struct {
	Bus
}

// Publish distributes the given event to all current subscribers. This method
// blocks for slow subscribers.
func (b *Publisher) Publish(e Event) {
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
