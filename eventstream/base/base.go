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
	// Returns all currently known Subscriptions.
	Subscriptions() []Subscription
}

type Subscription interface {
	PersistentID() string
	// Returns the currently active Selector.
	ActiveSelector() Selector
	LastAcknowledgedSequence() (int64, error)
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

type SequenceStore interface {
	Get(persistentClientID string) (int64, error)
	Store(persistentClientID string, sequence int64) error
}
