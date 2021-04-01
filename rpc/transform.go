package rpc

import (
	"github.com/golang/protobuf/ptypes"
	es "github.com/mtrense/ticker/eventstream/base"
	"google.golang.org/protobuf/types/known/structpb"
)

func EventToProto(e *es.Event) *Event {
	occurredAt, _ := ptypes.TimestampProto(e.OccurredAt)
	payload, _ := structpb.NewStruct(e.Payload)
	ev := &Event{
		Sequence:   e.Sequence,
		Aggregate:  e.Aggregate,
		Type:       e.Type,
		OccurredAt: occurredAt,
		Payload:    payload,
	}
	return ev
}

func ProtoToEvent(e *Event) *es.Event {
	return &es.Event{
		Sequence:   e.Sequence,
		Aggregate:  e.Aggregate,
		Type:       e.Type,
		OccurredAt: e.OccurredAt.AsTime(),
		Payload:    e.Payload.AsMap(),
	}
}

func BracketToProto(b *es.Bracket) *Bracket {
	return &Bracket{
		FirstSequence: b.NextSequence,
		LastSequence:  b.LastSequence,
	}
}

func ProtoToBracket(b *Bracket) *es.Bracket {
	return &es.Bracket{
		NextSequence: b.FirstSequence,
		LastSequence: b.LastSequence,
	}
}

func SelectorToProto(s *es.Selector) *Selector {
	return &Selector{
		Aggregate: s.Aggregate,
		Type:      s.Type,
	}
}

func ProtoToSelector(s *Selector) *es.Selector {
	return &es.Selector{
		Aggregate: s.Aggregate,
		Type:      s.Type,
	}
}
