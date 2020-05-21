package eventbus

import (
	"errors"
	"fmt"
	"log"
	"sync"
)

type Event interface{}

type Topic string

type EventBus struct {
	lock   sync.RWMutex
	topics map[Topic]*Subscriptions
}

func NewEventBus() *EventBus {
	return &EventBus{topics: make(map[Topic]*Subscriptions)}
}

func (b *EventBus) Subscribe(topic Topic) chan Event {
	c := make(chan Event, 1)
	sub := b.get(topic)
	if sub != nil {
		sub.Add(c)
	} else {
		sub := NewSubscriptions()
		sub.Add(c)
		b.set(topic, sub)
	}
	return c
}

func (b *EventBus) Unsubscribe(topic Topic, c chan Event) error {
	sub := b.get(topic)
	if sub != nil {
		if err := sub.Delete(c); err != nil {
			return err
		}
		if sub.Len() == 0 {
			b.delete(topic)
		}
		close(c)
		return nil
	}
	return fmt.Errorf("unknown event type '%s'", topic)
}

func (b *EventBus) Publish(topic Topic, e Event) {
	sub := b.get(topic)
	if sub != nil {
		sub.Publish(e)
	}
}

func (b *EventBus) get(topic Topic) *Subscriptions {
	b.lock.RLock()
	defer b.lock.RUnlock()
	if sub, ok := b.topics[topic]; ok {
		return sub
	}
	return nil
}

func (b *EventBus) set(topic Topic, sub *Subscriptions) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.topics[topic] = sub
}

func (b *EventBus) delete(topic Topic) {
	b.lock.Lock()
	defer b.lock.Unlock()
	delete(b.topics, topic)
}

type Subscriptions struct {
	lock        sync.RWMutex
	subscribers map[chan Event]struct{}
}

func NewSubscriptions() *Subscriptions {
	return &Subscriptions{subscribers: make(map[chan Event]struct{})}
}

func (s *Subscriptions) Add(c chan Event) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.subscribers[c] = struct{}{}
}

func (s *Subscriptions) Delete(c chan Event) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if _, ok := s.subscribers[c]; !ok {
		return errors.New("attempting to remove unknown subscriber")
	}
	delete(s.subscribers, c)
	return nil
}

func (s *Subscriptions) Len() int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return len(s.subscribers)
}

func (s *Subscriptions) Publish(e Event) {
	var wg sync.WaitGroup
	s.lock.RLock()
	wg.Add(len(s.subscribers))
	for c := range s.subscribers {
		go func(ch chan Event) {
			defer s.recoverClosedChannel(ch, &wg)
			ch <- e
		}(c)
	}
	s.lock.RUnlock()
	wg.Wait()
}

func (s *Subscriptions) recoverClosedChannel(c chan Event, wg *sync.WaitGroup) {
	if r := recover(); r != nil {
		log.Println("channel is closed")
		if err := s.Delete(c); err != nil {
			log.Printf("unable to remove closed channel from subscriptions")
		}
	}
	wg.Done()
}
