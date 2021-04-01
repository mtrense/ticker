package base

import (
	"errors"
	"strings"
)

type Selector struct {
	Aggregate []string
	Type      string
}

func ParseSelector(s string) (*Selector, error) {
	splitSelector := strings.Split(s, "/")
	if len(splitSelector) > 2 {
		return nil, errors.New("expected at most one slash")
	}
	aggregates := strings.Split(splitSelector[0], ".")
	if len(splitSelector) == 2 {
		return &Selector{
			Aggregate: aggregates,
			Type:      splitSelector[1],
		}, nil
	} else {
		return &Selector{
			Aggregate: aggregates,
			Type:      "",
		}, nil
	}
}

func (s *Selector) IsComplete() bool {
	if s.Type == "" {
		return false
	}
	for _, a := range s.Aggregate {
		if a == "" {
			return false
		}
	}
	return true
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
