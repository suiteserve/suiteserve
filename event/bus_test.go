package event

import (
	"testing"
)

func TestPublisher_Publish(t *testing.T) {
	var pub Publisher
	sub := pub.Subscribe()

	pub.Publish("hello")
	pub.Publish(nil)
	pub.Publish(3)
	pub.Publish("world")

	if e := <-sub.Ch(); e != "hello" {
		t.Errorf("got %q, want %q", e, "hello")
	}
	if e := <-sub.Ch(); e != nil {
		t.Errorf("got %q, want %v", e, nil)
	}
	if e := <-sub.Ch(); e != 3 {
		t.Errorf("got %q, want %q", e, 3)
	}

	sub.Unsubscribe()
	<-sub.Ch()
}

func TestBus_Subscribe(t *testing.T) {
	var pub Publisher
	pub.Publish("one")
	sub := pub.Subscribe()
	pub.Publish(2)
	if e := <-sub.Ch(); e != 2 {
		t.Errorf("got %q, want %q", e, 2)
	}
	sub.Unsubscribe()
	<-sub.Ch()
}

func TestSubscriber_Unsubscribe(t *testing.T) {
	var pub Publisher
	sub := pub.Subscribe()
	pub.Publish("one")
	sub.Unsubscribe()
	pub.Publish("two")
	<-sub.Ch()
}
