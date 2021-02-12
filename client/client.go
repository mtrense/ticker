package client

import (
	"net"
	"sync"

	"github.com/mtrense/ticker/eventstore"

	"github.com/mtrense/ticker/connection"
)

type TickerClient struct {
	clientName string
	address    string
	connection *connection.Connection
	wg         sync.WaitGroup
}

func NewClient(clientName, addr string) *TickerClient {
	return &TickerClient{
		clientName: clientName,
		address:    addr,
	}
}

func (s *TickerClient) Connect() error {
	conn, err := net.Dial("tcp", s.address)
	if err != nil {
		return err
	}
	s.connection = connection.NewConnection(s, s.handleMessage, conn)
	s.connection.Handle()
	return nil
}

func (s *TickerClient) Close() {

}

func (s *TickerClient) ConnectionClosing(conn *connection.Connection) {
}

func (s *TickerClient) ConnectionClosed(conn *connection.Connection) {
}

func (s *TickerClient) ConnectionOpened(conn *connection.Connection) {
}

func (s *TickerClient) SendAppend(event eventstore.Event) error {
	msg := connection.Msg(connection.MsgTypeAppend, connection.MsgAppend{
		ID:         event.ID,
		Aggregate:  event.Aggregate,
		EventType:  event.Type,
		OccurredAt: event.OccurredAt,
		Revision:   event.Revision,
		Event:      event.Payload,
	})
	return s.connection.Send(msg)
}

func (s *TickerClient) handleMessage(conn *connection.Connection, msg connection.Message) error {
	switch msg.Type {
	case connection.MsgTypeServerInfo:
		conn.Send(connection.Connect(s.clientName, 1))
	}
	return nil
}
