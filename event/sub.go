package event

import "sync"

// A Subscriber receives Events published to its parent Bus. The Unsubscribe
// method must always be called when the subscription is done.
type Subscriber struct {
	mu   sync.Mutex
	in   chan<- Event
	out  <-chan Event
	done bool
}

func newSubscriber() *Subscriber {
	var s Subscriber
	s.in, s.out = newInfCh()
	return &s
}

// Ch returns the singleton receive-only channel of in-order Events.
func (s *Subscriber) Ch() <-chan Event {
	return s.out
}

// Unsubscribe removes the Subscriber from its parent Bus. The channel returned
// by Ch will be closed soon after Unsubscribe is called.
func (s *Subscriber) Unsubscribe() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.done {
		return
	}
	s.done = true
	close(s.in)
}

func (s *Subscriber) publish(e Event) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.done {
		return false
	}
	s.in <- e
	return true
}

// newInfCh returns two new channels. Events sent to the first channel are
// immediately buffered and then consumed later by the second channel. Closing
// the first channel will throw out any buffered events and then will close the
// second channel.
func newInfCh() (chan<- Event, <-chan Event) {
	in := make(chan Event)
	out := make(chan Event)
	var buf []Event
	getNext := func() Event {
		if len(buf) == 0 {
			return nil
		}
		return buf[0]
	}
	getOut := func() chan Event {
		if len(buf) == 0 {
			return nil
		}
		return out
	}
	go func() {
		defer close(out)
		for {
			select {
			case e, ok := <-in:
				if !ok {
					return
				}
				buf = append(buf, e)
			case getOut() <- getNext():
				buf = buf[1:]
			}
		}
	}()
	return in, out
}
