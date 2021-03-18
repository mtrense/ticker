package server

import (
	"context"

	"github.com/golang/protobuf/ptypes"

	"github.com/mtrense/ticker/eventstream/base"

	"github.com/mtrense/ticker/rpc"
)

type eventStreamServer struct {
	rpc.UnimplementedEventStreamServer
	server *Server
}

func (s *eventStreamServer) Emit(ctx context.Context, event *rpc.Event) (*rpc.Ack, error) {
	occurredAt, err := ptypes.Timestamp(event.OccurredAt)
	if err != nil {
		return nil, err
	}
	payload := event.Payload.AsMap()
	ev := base.Event{
		Sequence:   -1,
		Aggregate:  event.Aggregate,
		Type:       event.Type,
		OccurredAt: occurredAt,
		Payload:    payload,
	}
	seq, err := s.server.streamBackend.Store(&ev)
	return &rpc.Ack{
		Sequence: int64(seq),
	}, err
}

func (s *eventStreamServer) Subscribe(subscription *rpc.Subscription, stream rpc.EventStream_SubscribeServer) error {

	return nil
}
