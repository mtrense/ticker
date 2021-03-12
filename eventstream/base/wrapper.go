package base

import "time"

type EventStreamWrapper struct {
	wrappedStream   EventStream
	startTime       time.Time
	lastEmittedTime time.Time
}

func NewWrapper(stream EventStream) *EventStreamWrapper {
	return NewWrapperWithStartTime(stream, time.Unix(0, 0))
}

func NewWrapperWithStartTime(stream EventStream, startTime time.Time) *EventStreamWrapper {
	return &EventStreamWrapper{
		wrappedStream:   stream,
		startTime:       startTime,
		lastEmittedTime: startTime,
	}
}

type EventBuilder = func(e *Event)

func (s *EventStreamWrapper) Emit(builders ...EventBuilder) (*Event, error) {
	event := &Event{}
	for _, b := range builders {
		b(event)
	}
	_, err := s.wrappedStream.Store(event)
	return event, err
}

func (s *EventStreamWrapper) Stream() EventStream {
	return s.wrappedStream
}

func (s *EventStreamWrapper) Agg(a ...string) EventBuilder {
	return func(e *Event) {
		e.Aggregate = a
	}
}

func (s *EventStreamWrapper) DefAgg() EventBuilder {
	return s.Agg("test", "aggregate", "1")
}

func (s *EventStreamWrapper) Type(t string) EventBuilder {
	return func(e *Event) {
		e.Type = t
	}
}

func (s *EventStreamWrapper) IncrBy(duration time.Duration) EventBuilder {
	return func(e *Event) {
		s.lastEmittedTime = s.lastEmittedTime.Add(duration)
		e.OccurredAt = s.lastEmittedTime
	}
}

func (s *EventStreamWrapper) After(duration time.Duration) EventBuilder {
	return func(e *Event) {
		e.OccurredAt = s.lastEmittedTime.Add(duration)
	}
}
