// +build memory

package memory

import (
	"sync"

	es "github.com/mtrense/ticker/eventstream/base"
)

type SequenceStore struct {
	sequences map[string]int64
	mutex     sync.Mutex
}

func NewMemorySequenceStore() es.SequenceStore {
	return &SequenceStore{
		sequences: make(map[string]int64),
	}
}

func (s *SequenceStore) Get(persistentClientID string) (int64, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if seq, ok := s.sequences[persistentClientID]; ok {
		return seq, nil
	} else {
		return 0, nil
	}
}

func (s *SequenceStore) Store(persistentClientID string, sequence int64) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.sequences[persistentClientID] = sequence
	return nil
}
