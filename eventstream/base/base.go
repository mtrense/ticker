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
	Stream(ctx context.Context, sel Selector, bracket Bracket, handler EventHandler) error
	Subscribe(ctx context.Context, persistentClientID string, sel Selector, handler EventHandler) (Subscription, error)
	// Returns all currently known Subscriptions
	Subscriptions() []Subscription
}

type Subscription interface {
	PersistentID() string
	ActiveSelector() Selector
	LastAcknowledgedSequence() int64
	Acknowledge(sequence int64) error
	// Returns whether this Subscription is currently active.
	Active() bool
	// Returns the time this Subscription last became inactive.
	InactiveSince() time.Time
	// Returns how often this Subscription has dropped out of the live stream.
	DropOuts() int
	// Closes this Subscription and removes all associated state. A Subscription can not be resumed after this call.
	Shutdown()
}

type Selector struct {
	Aggregate []string
	Type      string
}

type SelectOption func(s *Selector)

func Select(options ...SelectOption) Selector {
	sel := Selector{
		Aggregate: []string{},
		Type:      "",
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

func SelectAggregate(agg ...string) SelectOption {
	return func(s *Selector) {
		s.Aggregate = agg
	}
}

func (s *Selector) Matches(event *Event) bool {
	if s.Type != "" && s.Type != event.Type {
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

type Bracket struct {
	NextSequence int64
	LastSequence int64
}

func All() Bracket {
	return Bracket{
		NextSequence: 1,
		LastSequence: math.MaxInt64,
	}
}

func Range(next, last int64) Bracket {
	return Bracket{
		NextSequence: next,
		LastSequence: last,
	}
}

func From(next int64) Bracket {
	return Bracket{
		NextSequence: next,
		LastSequence: math.MaxInt64,
	}
}
