package base

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
