package event

import (
	"runtime"
	"sync"
	"testing"
)

var wg sync.WaitGroup

func TestBus_Publish(t *testing.T) {
	var bus Bus
	sub := bus.Subscribe()
	defer sub.Unsubscribe()
	wg.Add(1)
	go func() {
		defer wg.Done()
		bus.Publish("event")
	}()
	runtime.Gosched()
	if e := <-sub.Ch(); e != "event" {
		t.Errorf("got %q, want %q", e, "event")
	}
	wg.Wait()
}
