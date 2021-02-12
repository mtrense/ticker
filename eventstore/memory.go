package eventstore

import (
	"sync"

	log "github.com/mtrense/soil/logging"
)

type MemoryEventStore struct {
	mutex         sync.Mutex
	events        []*Event
	clientCursors map[string]int
}

func NewMemoryEventStore() *MemoryEventStore {
	return &MemoryEventStore{
		events:        make([]*Event, 0),
		clientCursors: make(map[string]int),
	}
}

func (m *MemoryEventStore) Append(event *Event) int {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	event.GlobalSequence = len(m.events) + 1
	m.events = append(m.events, event)
	log.L().Debug().Int("globalSequence", event.GlobalSequence).Msg("Appended event")
	return event.GlobalSequence
}

func (m *MemoryEventStore) Stream(fn func(event *Event), aggregate []string, typ string, begin int) {
	for _, e := range m.events[begin:] {
		if e.Matches(typ, aggregate...) {
			fn(e)
		}
	}
}

func (m *MemoryEventStore) SetClientCursor(clientID string, sequence int) {
	m.clientCursors[clientID] = sequence
}

func (m *MemoryEventStore) GetClientCursor(clientID string) int {
	return m.clientCursors[clientID]
}
