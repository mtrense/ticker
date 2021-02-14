package server

import (
	"context"

	"github.com/mtrense/ticker/rpc"
)

type eventStreamServer struct {
	rpc.UnimplementedEventStreamServer
	server *Server
}

func (s *eventStreamServer) Emit(context.Context, *rpc.Event) (*rpc.Ack, error) {
	return &rpc.Ack{}, nil
}

func (s *eventStreamServer) Subscribe(subscription *rpc.Subscription, stream rpc.EventStream_SubscribeServer) error {

	return nil
}
