package event

import "sync"

// A Bus is a publish/subscribe method of transportation for Events of any type.
// Many subscribers can subscribe to Events published to the Bus. Subscribers
// do not receive Events published prior to the time of subscription, and slow
// receivers do not block publishers. Subscribers are also guaranteed to receive
// events in the order in which they were published. Bus only allows
// subscribing: use Publisher to add the Event publishing capability.
type Bus struct {
	mu   sync.Mutex
	subs []*Subscriber
}

// Subscribe returns a new Subscriber that receives new Events published to the
// Bus. When the subscription is done, the returned Subscriber's Unsubscribe
// method must be called.
func (b *Bus) Subscribe() *Subscriber {
	sub := newSubscriber()
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subs = append(b.subs, sub)
	return sub
}

// A Publisher wraps a Bus and adds the Event publishing capability. Its purpose
// is to separate publishing from subscribing. The wrapped Bus, which only
// allows subscribing, may be handed out more liberally than a Publisher that
// allows both. The zero value for Publisher is ready to use.
type Publisher struct {
	Bus
}

// Publish distributes the given Event to all current subscribers.
func (b *Publisher) Publish(e Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	subs := b.subs[:0]
	for _, s := range b.subs {
		if s.publish(e) {
			subs = append(subs, s)
		}
	}
	b.subs = subs
}
