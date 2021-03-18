// +build memory

package memory

import (
	"context"
	"errors"
	"sync"

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
		sub.publishEvent(event)
	}
	return event.Sequence, nil
}

func (s *EventStream) LastSequence() int64 {
	return int64(len(s.events))
}

func (s *EventStream) Get(sequence int64) (*es.Event, error) {
	return s.events[sequence-1], nil
}

func (s *EventStream) Slice(ctx context.Context, startSequence, endSequence int64, handler es.EventHandler) error {
	for _, event := range s.events[startSequence-1 : endSequence] {
		if err := ctx.Err(); errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil
		}
		handler(event)
	}
	return nil
}

func (s *EventStream) Subscribe(ctx context.Context, persistentClientID string, handler es.EventHandler, sel es.Selector) (es.Subscription, error) {
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
	sub.live = true
	return s.LastSequence(), nil
}

func (s *EventStream) unsubscribe(sub *Subscription) {
	s.writeLock.Lock()
	defer s.writeLock.Unlock()
	delete(s.subscriptions, sub.clientID)
}
