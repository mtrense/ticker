package client

import (
	"context"
	"errors"
	"io"

	es "github.com/mtrense/ticker/eventstream/base"

	"github.com/mtrense/ticker/rpc"
	"google.golang.org/grpc"
)

type Client struct {
	connection        *grpc.ClientConn
	eventStreamClient rpc.EventStreamClient
}

type Option = func(c *Client)

func NewClient(conn *grpc.ClientConn, opts ...Option) *Client {
	cl := &Client{
		connection:        conn,
		eventStreamClient: rpc.NewEventStreamClient(conn),
	}
	for _, opt := range opts {
		opt(cl)
	}
	return cl
}

func (s *Client) Emit(ctx context.Context, event es.Event) (es.Event, error) {
	ack, err := s.eventStreamClient.Emit(ctx, rpc.EventToProto(&event))
	if ack != nil {
		event.Sequence = ack.Sequence
		return event, err
	} else {
		return event, errors.New("didn't receive an Ack")
	}
}

func (s *Client) Stream(ctx context.Context, selector *es.Selector, bracket *es.Bracket, handler es.EventHandler) error {
	req := &rpc.StreamRequest{
		Bracket:  rpc.BracketToProto(bracket),
		Selector: rpc.SelectorToProto(selector),
	}
	stream, err := s.eventStreamClient.Stream(ctx, req)
	if err != nil {
		return err
	}
	for {
		ev, err := stream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		event := rpc.ProtoToEvent(ev)
		handler(event)
	}
	return nil
}

func (s *Client) Subscribe(ctx context.Context) {

}
