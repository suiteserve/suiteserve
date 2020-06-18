package event

// A Subscriber receives events published to its parent Bus. The Unsubscribe
// method should always be called when the subscription is done.
type Subscriber struct {
	ch   chan Event
	done chan interface{}
}

func newSubscriber() *Subscriber {
	return &Subscriber{
		ch:   make(chan Event),
		done: make(chan interface{}),
	}
}

// Ch returns the singleton receive-only channel of in-order Events. The channel
// may or may not be closed after the Unsubscribe method is called. It is an
// error to call Ch or receive from the returned channel after Unsubscribe is
// called.
func (s *Subscriber) Ch() <-chan Event {
	return s.ch
}

// Unsubscribe removes this Subscriber from its parent Bus. If an Event is
// waiting to be received, the Event is thrown out and no more will be received.
func (s *Subscriber) Unsubscribe() {
	close(s.done)
}
