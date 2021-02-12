package eventstore

type Publisher struct {
	eventStore *EventStore
	consumers  map[EventConsumer]struct{}
}

func NewPublisher(store *EventStore) *Publisher {
	return &Publisher{
		eventStore: store,
		consumers:  make(map[EventConsumer]struct{}),
	}
}

type EventConsumer interface {
	Consume(event *Event)
}

func (s *Publisher) AddConsumer(consumer EventConsumer) {
	s.consumers[consumer] = struct{}{}
}

func (s *Publisher) RemoveConsumer(consumer EventConsumer) {
	delete(s.consumers, consumer)
}
