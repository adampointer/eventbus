package eventbus

import (
	"testing"
	"time"

	bus1 "github.com/asaskevich/EventBus"
)

func BenchmarkEventBus_Publish_A(b *testing.B) {
	bus := bus1.New()
	if err := bus.SubscribeAsync("test", func() {
		time.Sleep(time.Millisecond)
	}, true); err != nil {
		panic(err)
	}

	for n := 0; n < b.N; n++ {
		bus.Publish("test")
	}
}

func BenchmarkEventBus_Publish_B(b *testing.B) {
	bus := NewEventBus()
	c := bus.Subscribe("test")

	go func() {
		for range c {
			time.Sleep(time.Millisecond)
		}
	}()

	for n := 0; n < b.N; n++ {
		bus.Publish("test", "foo")
	}
	bus.Unsubscribe("test", c)
}
