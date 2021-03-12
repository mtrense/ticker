// +build memory

package memory

import (
	"sync"

	es "github.com/mtrense/ticker/eventstream/base"
)

type EventStream struct {
	upstream          es.EventStream
	events            []*es.Event
	writeLock         sync.Mutex
	listeners         map[chan *es.Event]struct{}
	defaultBufferSize int
}

func New() *EventStream {
	return &EventStream{
		listeners:         make(map[chan *es.Event]struct{}),
		defaultBufferSize: 100,
	}
}

func NewCachingMemoryEventStream(upstream es.EventStream) *EventStream {
	return &EventStream{}
}

func (s *EventStream) Store(event *es.Event) (int, error) {
	s.writeLock.Lock()
	defer s.writeLock.Unlock()
	// Sequence starts at 1
	seq := len(s.events) + 1
	event.Sequence = seq
	s.events = append(s.events, event)
	for l, _ := range s.listeners {
		select {
		case l <- event:
		default:
			close(l)
			delete(s.listeners, l)
		}
	}
	return event.Sequence, nil
}

func (s *EventStream) LastSequence() int {
	return len(s.events)
}

func (s *EventStream) Get(sequence int) (*es.Event, error) {
	return s.events[sequence-1], nil
}

func (s *EventStream) Slice(startSequence, endSequence int, handler es.EventHandlerWithCancel) error {
	for _, event := range s.events[startSequence-1 : endSequence] {
		if !handler(event) {
			return nil
		}
	}
	return nil
}

func (s *EventStream) Subscribe(handler es.EventHandler, sel es.Selector) func() {
	unsubscriber := make(chan bool)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go s.handleSubscription(handler, sel, s.defaultBufferSize, unsubscriber, wg)
	wg.Wait()
	return func() {
		unsubscriber <- true
		close(unsubscriber)
	}
}

func (s *EventStream) handleSubscription(handler es.EventHandler, sel es.Selector, bufferSize int, unsubscriber chan bool, wg *sync.WaitGroup) {
	startSequence := sel.FirstSequence
	if bufferSize <= 0 {
		bufferSize = 100
	}
	subscriberNotified := false
	for {
		if s.unsubscribe(unsubscriber) {
			// Client unsubscribed, good bye
			return
		}
		listener := make(chan *es.Event, bufferSize)
		s.writeLock.Lock()
		s.listeners[listener] = struct{}{}
		endSequence := s.LastSequence()
		s.writeLock.Unlock()
		if !subscriberNotified {
			subscriberNotified = true
			wg.Done()
		}
		if cancel := s.streamHistory(handler, sel, startSequence, endSequence, unsubscriber); cancel {
			return
		}
		if seq, cancel := s.streamLiveUpdates(listener, handler, sel, startSequence, unsubscriber); cancel {
			return
		} else {
			startSequence = seq
		}
	}
}

func (s *EventStream) streamHistory(handler es.EventHandler, sel es.Selector, startSequence int, endSequence int, unsubscriber chan bool) bool {
	var result bool
	s.Slice(startSequence, endSequence, func(event *es.Event) bool {
		if s.unsubscribe(unsubscriber) {
			// Client unsubscribed, good bye
			result = true
			return false
		}
		if sel.Matches(event) {
			handler(event)
		}
		return true
	})
	return result
}

func (s *EventStream) streamLiveUpdates(listener chan *es.Event, handler es.EventHandler, sel es.Selector, startSequence int, unsubscriber chan bool) (int, bool) {
	for {
		select {
		case event, live := <-listener:
			if !live {
				// Channel was closed due to buffer overflow, return new start sequence and rewind
				return startSequence, false
			} else {
				startSequence = event.Sequence + 1
				if sel.Matches(event) {
					handler(event)
				}
			}
		case <-unsubscriber:
			// Client unsubscribed, good bye
			return startSequence, true
		}
	}
}

func (s *EventStream) unsubscribe(unsubscriber chan bool) bool {
	select {
	case <-unsubscriber:
		return true
	default:
		return false
	}
}
