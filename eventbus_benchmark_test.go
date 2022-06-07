package eventbus

import (
	"testing"
	"time"
)

func BenchmarkEventBus_Publish(b *testing.B) {
	bus := NewEventBus()
	c := bus.Subscribe("test")

	go func() {
		for range c {
			time.Sleep(time.Millisecond)
		}
	}()

	for n := 0; n < b.N; n++ {
		bus.Publish("test", &testEvent{})
	}
	if err := bus.Unsubscribe("test", c); err != nil {
		panic(err)
	}
}
