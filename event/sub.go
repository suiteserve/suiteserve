package event

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

func (s *Subscriber) Ch() <-chan Event {
	return s.ch
}

func (s *Subscriber) Unsubscribe() {
	close(s.done)
}
