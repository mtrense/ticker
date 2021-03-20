// +build memory

package memory

import (
	"context"
	"errors"
	"sync"
	"time"

	es "github.com/mtrense/ticker/eventstream/base"
)

type EventStream struct {
	upstream          es.EventStream
	events            []*es.Event
	writeLock         sync.Mutex
	subscriptions     map[string]*Subscription
	defaultBufferSize int
}

func New() *EventStream {
	return &EventStream{
		defaultBufferSize: 100,
		subscriptions:     make(map[string]*Subscription),
	}
}

func (s *EventStream) Store(event *es.Event) (int64, error) {
	s.writeLock.Lock()
	defer s.writeLock.Unlock()
	// Sequence starts at 1
	seq := int64(len(s.events) + 1)
	event.Sequence = seq
	s.events = append(s.events, event)
	for _, sub := range s.subscriptions {
		if sub.active {
			sub.publishEvent(event)
		}
	}
	return event.Sequence, nil
}

func (s *EventStream) LastSequence() int64 {
	return int64(len(s.events))
}

func (s *EventStream) Get(sequence int64) (*es.Event, error) {
	return s.events[sequence-1], nil
}

func (s *EventStream) Stream(ctx context.Context, sel es.Selector, bracket es.Bracket, handler es.EventHandler) error {
	if bracket.NextSequence < 1 {
		bracket.NextSequence = 1
	}
	if bracket.LastSequence > s.LastSequence() {
		bracket.NextSequence = s.LastSequence()
	}
	for _, event := range s.events[bracket.NextSequence-1 : bracket.LastSequence] {
		if err := ctx.Err(); errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil
		}
		if sel.Matches(event) {
			handler(event)
		}
	}
	return nil
}

func (s *EventStream) Subscribe(ctx context.Context, persistentClientID string, sel es.Selector, handler es.EventHandler) (es.Subscription, error) {
	s.writeLock.Lock()
	sub, present := s.subscriptions[persistentClientID]
	if !present {
		sub = newSubscription(s, persistentClientID, sel)
		s.subscriptions[persistentClientID] = sub
	}
	s.writeLock.Unlock()
	err := sub.handleSubscription(ctx, handler)
	return sub, err
}

func (s *EventStream) attachSubscription(sub *Subscription) (int64, error) {
	s.writeLock.Lock()
	defer s.writeLock.Unlock()
	sub.buffer = make(chan *es.Event, s.defaultBufferSize)
	sub.active = true
	sub.live = true
	return s.LastSequence(), nil
}

func (s *EventStream) unsubscribe(sub *Subscription) {
	s.writeLock.Lock()
	defer s.writeLock.Unlock()
	sub.active = false
	sub.inactiveSince = time.Now()
	close(sub.buffer)
}

func (s *EventStream) Subscriptions() []es.Subscription {
	result := make([]es.Subscription, 0, len(s.subscriptions))
	for _, sub := range s.subscriptions {
		result = append(result, sub)
	}
	return result
}
