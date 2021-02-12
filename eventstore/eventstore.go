package eventstore

import "time"

type EventStore interface {
	Append(event *Event) int
	Stream(fn func(event *Event), aggregate []string, typ string, begin int)
	SetClientCursor(clientID string, sequence int)
	GetClientCursor(clientID string) int
}

//type ClientState struct {
//}

type Event struct {
	ID             string                 `json:"id,omitempty"`
	GlobalSequence int                    `json:"global_sequence,omitempty"`
	Aggregate      []string               `json:"aggregate,omitempty"`
	Type           string                 `json:"type,omitempty"`
	OccurredAt     time.Time              `json:"occurred_at,omitempty"`
	Revision       int                    `json:"revision,omitempty"`
	Payload        map[string]interface{} `json:"payload,omitempty"`
}

func (s *Event) Matches(typ string, aggregate ...string) bool {
	if typ != "" && typ != s.Type {
		return false
	}
	for index, segment := range aggregate {
		if segment != "" && segment != s.Aggregate[index] {
			return false
		}
	}
	return true
}
