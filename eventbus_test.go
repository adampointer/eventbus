package eventbus

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testEvent struct {
	t Topic
}

func newTestEvent() *testEvent {
	return &testEvent{}
}

func (e *testEvent) Type() EventType {
	return "test"
}

func (e *testEvent) Topic() Topic {
	return e.t
}

func (e *testEvent) SetTopic(t Topic) {
	e.t = t
}

func TestEventBus_Subscribe(t *testing.T) {
	bus := NewEventBus()
	var ev1 Topic = "foo"
	var ev2 Topic = "bar"

	_ = bus.Subscribe(ev1)
	_ = bus.Subscribe(ev1)
	_ = bus.Subscribe(ev2)

	assert.Len(t, bus.topics, 2)
	assert.Len(t, bus.topics[ev1].subscribers, 2)
	assert.Len(t, bus.topics[ev2].subscribers, 1)
}

func TestEventBus_Unsubscribe(t *testing.T) {
	bus := NewEventBus()
	var ev1 Topic = "foo"

	c := bus.Subscribe(ev1)

	assert.Len(t, bus.topics, 1)
	assert.Len(t, bus.topics[ev1].subscribers, 1)

	assert.NoError(t, bus.Unsubscribe(ev1, c))

	assert.Len(t, bus.topics, 0)
}

func TestEventBus_Publish(t *testing.T) {
	bus := NewEventBus()
	var ev1 Topic = "foo"
	e := newTestEvent()
	c := bus.Subscribe(ev1)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		select {
		case evt := <-c:
			assert.Equal(t, e, evt)
		case <-time.After(time.Second):
			assert.Fail(t, "test timed out")
		}
		wg.Done()
	}()
	bus.Publish(ev1, e)
	wg.Wait()
}

func TestEventBus_Concurrency(t *testing.T) {
	bus := NewEventBus()
	var ev1 Topic = "foo"
	e := newTestEvent()
	n := 10
	var wg sync.WaitGroup
	wg.Add(n)
	chans := make([]chan Event, n)

	for i := 0; i < n; i++ {
		c := bus.Subscribe(ev1)
		chans[i] = c

		go func() {
			select {
			case evt := <-c:
				assert.Equal(t, e, evt)
			case <-time.After(time.Second):
				assert.Fail(t, "test timed out")
			}
			wg.Done()
		}()
	}
	bus.Publish(ev1, e)
	wg.Wait()

	for _, c := range chans {
		go func(ch chan Event) {
			assert.NoError(t, bus.Unsubscribe(ev1, ch))
		}(c)
	}
}

func TestEventBus_ClosedChannel(t *testing.T) {
	bus := NewEventBus()
	var ev1 Topic = "foo"
	e := newTestEvent()
	n := 10
	var wg sync.WaitGroup
	wg.Add(n)
	chans := make([]chan Event, n)

	for i := 0; i < n; i++ {
		c := bus.Subscribe(ev1)
		chans[i] = c

		go func() {
			select {
			case evt := <-c:
				if evt != nil {
					assert.Equal(t, e, evt)
				}
			case <-time.After(time.Second):
				assert.Fail(t, "test timed out")
			}
			wg.Done()
		}()
	}
	close(chans[5])
	bus.Publish(ev1, e)
	wg.Wait()

	assert.Len(t, bus.topics[ev1].subscribers, n-1)
}
