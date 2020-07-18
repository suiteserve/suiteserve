package event

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPublisher_Publish(t *testing.T) {
	var pub Publisher
	sub := pub.Subscribe()

	pub.Publish("hello")
	pub.Publish(nil)
	pub.Publish(3)
	pub.Publish("world")

	assert.Equal(t, "hello", <-sub.Ch())
	assert.Equal(t, nil, <-sub.Ch())
	assert.Equal(t, 3, <-sub.Ch())

	sub.Unsubscribe()
	<-sub.Ch()
}

func TestBus_Subscribe(t *testing.T) {
	var pub Publisher
	pub.Publish("one")
	sub := pub.Subscribe()
	pub.Publish(2)
	assert.Equal(t, 2, <-sub.Ch())
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
