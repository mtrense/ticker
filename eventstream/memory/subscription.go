// +build memory

package memory

import (
	"context"
	"sync/atomic"
	"time"

	es "github.com/mtrense/ticker/eventstream/base"
)

type Subscription struct {
	stream                   *EventStream
	clientID                 string
	live                     bool
	active                   bool
	inactiveSince            time.Time
	activeSelector           es.Selector
	lastAcknowledgedSequence int64
	buffer                   chan *es.Event
	handler                  es.EventHandler
	lastError                error
}

func newSubscription(stream *EventStream, clientID string, sel es.Selector) *Subscription {
	return &Subscription{
		stream:         stream,
		clientID:       clientID,
		live:           false,
		activeSelector: sel,
	}
}

func (s *Subscription) PersistentClientID() string {
	return s.clientID
}

func (s *Subscription) ActiveSelector() es.Selector {
	return s.activeSelector
}

func (s *Subscription) LastAcknowledgedSequence() int64 {
	return s.lastAcknowledgedSequence
}

func (s *Subscription) Acknowledge(sequence int64) error {
	atomic.StoreInt64(&s.lastAcknowledgedSequence, sequence)
	return nil
}

func (s *Subscription) publishEvent(event *es.Event) {
	if s.live {
		select {
		case s.buffer <- event:
		default:
			close(s.buffer)
			s.live = false
		}
	}
}

func (s *Subscription) handleSubscription(ctx context.Context, handler es.EventHandler) error {
	// TODO Handle already subscribed
	nextSequence := int64(1)
	lastKnownSequence, err := s.stream.attachSubscription(s)
	if err != nil {
		return err
	}
	go func() {
		for {
			err := s.stream.Stream(ctx, s.activeSelector, es.Range(nextSequence, lastKnownSequence), func(e *es.Event) {
				if s.activeSelector.Matches(e) {
					handler(e)
				}
				nextSequence = e.Sequence + 1
			})
			if err != nil {
				s.lastError = err
				return
			}
		liveStream:
			for {
				select {
				case event, live := <-s.buffer:
					if !live {
						if seq, err := s.stream.attachSubscription(s); err == nil {
							lastKnownSequence = seq
							break liveStream
						} else {
							s.lastError = err
							return
						}
					} else {
						if s.activeSelector.Matches(event) {
							handler(event)
						}
						nextSequence = event.Sequence + 1
					}
				case <-ctx.Done():
					s.stream.unsubscribe(s)
					return
				}
			}
		}
	}()
	return nil
}
