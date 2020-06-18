package event

import (
	"sync"
	"testing"
	"time"
)

func TestPublisher_Publish(t *testing.T) {
	wg := &sync.WaitGroup{}
	var pub Publisher
	sub := pub.Subscribe()
	defer sub.Unsubscribe()
	wg.Add(1)
	go func() {
		defer wg.Done()
		pub.Publish("event")
		pub.Publish(nil)
		pub.Publish(3)
	}()
	if e := <-sub.Ch(); e != "event" {
		t.Errorf("got %q, want %q", e, "event")
	}
	if e := <-sub.Ch(); e != nil {
		t.Errorf("got %q, want %v", e, nil)
	}
	if e := <-sub.Ch(); e != 3 {
		t.Errorf("got %q, want %q", e, 3)
	}
	wg.Wait()
}

func TestBus_Subscribe(t *testing.T) {
	wg := &sync.WaitGroup{}
	var pub Publisher
	wg.Add(1)
	go func() {
		defer wg.Done()
		pub.Publish("event")
		wg.Add(1)
		time.AfterFunc(2 * time.Millisecond, func() {
			defer wg.Done()
			pub.Publish(3)
		})
	}()
	wg.Add(1)
	time.AfterFunc(1 * time.Millisecond, func() {
		defer wg.Done()
		sub := pub.Subscribe()
		defer sub.Unsubscribe()
		if e := <-sub.Ch(); e != 3 {
			t.Errorf("got %q, want %q", e, 3)
		}
	})
	wg.Wait()
}

func TestSubscriber_Unsubscribe(t *testing.T) {
	wg := &sync.WaitGroup{}
	var pub Publisher
	sub := pub.Subscribe()
	wg.Add(1)
	go func() {
		defer wg.Done()
		pub.Publish("event")
	}()
	wg.Add(1)
	time.AfterFunc(1 * time.Millisecond, func() {
		defer wg.Done()
		sub.Unsubscribe()
	})
	wg.Wait()
}
