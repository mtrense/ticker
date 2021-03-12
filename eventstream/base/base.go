package base

import (
	"math"
	"time"
)

type Event struct {
	Sequence   int
	Aggregate  []string
	Type       string
	OccurredAt time.Time
	Payload    map[string]interface{}
}

type EventHandler func(e *Event)
type EventHandlerWithCancel func(e *Event) bool

type EventStream interface {
	Store(event *Event) (int, error)
	LastSequence() int
	Get(sequence int) (*Event, error)
	Slice(startSequence, endSequence int, handler EventHandlerWithCancel) error
	Subscribe(handler EventHandler, sel Selector) func()
}

type Selector struct {
	Aggregate     []string
	Type          string
	FirstSequence int
	LastSequence  int
}

type SelectOption func(s *Selector)

func Select(options ...SelectOption) Selector {
	sel := Selector{
		Aggregate:     nil,
		Type:          "",
		FirstSequence: 1,
		LastSequence:  math.MaxInt64,
	}
	for _, opt := range options {
		opt(&sel)
	}
	return sel
}

func SelectType(t string) SelectOption {
	return func(s *Selector) {
		s.Type = t
	}
}

func SelectRange(first, last int) SelectOption {
	return func(s *Selector) {
		s.FirstSequence = first
		s.LastSequence = last
	}
}

func SelectAggregate(agg ...string) SelectOption {
	return func(s *Selector) {
		s.Aggregate = agg
	}
}

func (s *Selector) Matches(event *Event) bool {
	if s.Type != event.Type {
		return false
	}
	if s.FirstSequence > event.Sequence {
		return false
	}
	if s.LastSequence < event.Sequence {
		return false
	}
	if len(s.Aggregate) > len(event.Aggregate) {
		return false
	}
	for index, token := range s.Aggregate {
		if token != "" && event.Aggregate[index] != token {
			return false
		}
	}
	return true
}
