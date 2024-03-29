// +build memory

package memory

import (
	"context"
	"sync"
	"time"

	es "github.com/mtrense/ticker/eventstream/base"
)

type Subscription struct {
	stream         *EventStream
	clientID       string
	live           bool
	active         bool
	inactiveSince  time.Time
	activeSelector es.Selector
	buffer         chan *es.Event
	handler        es.EventHandler
	dropOuts       int
	lastError      error
	wg             sync.WaitGroup
}

func newSubscription(stream *EventStream, clientID string, sel es.Selector) *Subscription {
	return &Subscription{
		stream:         stream,
		clientID:       clientID,
		live:           false,
		activeSelector: sel,
	}
}

func (s *Subscription) PersistentID() string {
	return s.clientID
}

func (s *Subscription) ActiveSelector() es.Selector {
	return s.activeSelector
}

func (s *Subscription) LastAcknowledgedSequence() (int64, error) {
	return s.stream.sequenceStore.Get(s.clientID)
}

func (s *Subscription) Acknowledge(sequence int64) error {
	return s.stream.sequenceStore.Store(s.clientID, sequence)
}

func (s *Subscription) publishEvent(event *es.Event) {
	if s.live {
		select {
		case s.buffer <- event:
		default:
			close(s.buffer)
			s.live = false
			s.dropOuts += 1
		}
	}
}

func (s *Subscription) handleSubscription(ctx context.Context, handler es.EventHandler) error {
	s.wg.Add(1)
	lastSequence, err := s.LastAcknowledgedSequence()
	if err != nil {
		return err
	}
	nextSequence := lastSequence + 1
	lastKnownSequence, err := s.stream.attachSubscription(s)
	if err != nil {
		return err
	}
	go func() {
		defer s.wg.Done()
		for {
			if nextSequence <= lastKnownSequence {
				err := s.stream.Stream(ctx, s.activeSelector, es.Range(nextSequence, lastKnownSequence), func(e *es.Event) error {
					if s.activeSelector.Matches(e) {
						if err := handler(e); err != nil {
							s.lastError = err
							s.stream.unsubscribe(s)
							return err
						}
					}
					if err := s.Acknowledge(e.Sequence); err != nil {
						s.lastError = err
						s.stream.unsubscribe(s)
						return err
					}
					nextSequence = e.Sequence + 1
					return nil
				})
				if err != nil {
					s.lastError = err
					return
				}
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
							if err := handler(event); err != nil {
								s.lastError = err
								s.stream.unsubscribe(s)
								return
							}
						}
						if err := s.Acknowledge(event.Sequence); err != nil {
							s.lastError = err
							s.stream.unsubscribe(s)
							return
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

func (s *Subscription) Active() bool {
	return s.active
}

func (s *Subscription) InactiveSince() time.Time {
	return s.inactiveSince
}

func (s *Subscription) Wait() error {
	s.wg.Wait()
	return s.lastError
}

func (s *Subscription) DropOuts() int {
	return s.dropOuts
}

func (s *Subscription) Shutdown() {

}
