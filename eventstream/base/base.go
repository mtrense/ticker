package base

import (
	"context"
	"math"
	"time"
)

type Event struct {
	Sequence   int64                  `json:"sequence,omitempty" yaml:"sequence,omitempty"`
	Aggregate  []string               `json:"aggregate,omitempty" yaml:"aggregate,omitempty"`
	Type       string                 `json:"type,omitempty" yaml:"type,omitempty"`
	OccurredAt time.Time              `json:"occurred_at,omitempty" yaml:"occurred_at,omitempty"`
	Payload    map[string]interface{} `json:"payload,omitempty" yaml:"payload,omitempty"`
}

type EventHandler func(e *Event)

type EventStream interface {
	Store(event *Event) (int64, error)
	LastSequence() int64
	Get(sequence int64) (*Event, error)
	Slice(ctx context.Context, startSequence, endSequence int64, handler EventHandler) error
	Subscribe(ctx context.Context, persistentClientID string, handler EventHandler, sel Selector) (Subscription, error)
}

type Subscription interface {
	PersistentClientID() string
	ActiveSelector() Selector
	LastAcknowledgedSequence() int64
	Acknowledge(sequence int64) error
}

type Selector struct {
	Aggregate    []string
	Type         string
	NextSequence int64
	LastSequence int64
}

type SelectOption func(s *Selector)

func Select(options ...SelectOption) Selector {
	sel := Selector{
		Aggregate:    []string{},
		Type:         "",
		NextSequence: 1,
		LastSequence: math.MaxInt64,
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

func SelectStart(next int64) SelectOption {
	return func(s *Selector) {
		s.NextSequence = next
	}
}

func SelectRange(first, last int64) SelectOption {
	return func(s *Selector) {
		s.NextSequence = first
		s.LastSequence = last
	}
}

func SelectAggregate(agg ...string) SelectOption {
	return func(s *Selector) {
		s.Aggregate = agg
	}
}

func (s *Selector) Matches(event *Event) bool {
	if s.Type != "" && s.Type != event.Type {
		return false
	}
	if s.NextSequence > event.Sequence {
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
