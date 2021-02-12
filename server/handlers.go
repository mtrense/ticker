package server

import (
	log "github.com/mtrense/soil/logging"
	"github.com/mtrense/ticker/connection"
	"github.com/mtrense/ticker/eventstore"
)

func (s *TickerServer) handleConnect(conn *connection.Connection, state *ConnectionState, msg connection.Message) error {
	var payload connection.MsgConnect
	if err := msg.Unpack(&payload); err != nil {
		return err
	}
	if payload.ProtocolVersion != 1 {
		log.L().Error().Int("requestedProtocolVersion", payload.ProtocolVersion).Msg("Protocol version not supported")
		conn.Send(connection.Error("Protocol version not supported"))
		return ErrProtocolVersionUnsupported
	}
	state.ClientName = payload.ClientName
	state.ProtocolVersion = payload.ProtocolVersion
	log.L().Debug().
		Str("remoteAddress", conn.RemoteAddr().String()).
		Str("clientName", payload.ClientName).
		Int("protocolVersion", payload.ProtocolVersion).
		Msg("Client connection initialized")
	return nil
}

func (s *TickerServer) handleSubscribe(conn *connection.Connection, state *ConnectionState, msg connection.Message) error {

	return nil
}

func (s *TickerServer) handleAppend(conn *connection.Connection, state *ConnectionState, msg connection.Message) error {
	var payload connection.MsgAppend
	if err := msg.Unpack(&payload); err != nil {
		return err
	}
	event := &eventstore.Event{
		ID:         payload.ID,
		Aggregate:  payload.Aggregate,
		Type:       payload.EventType,
		OccurredAt: payload.OccurredAt,
		Revision:   payload.Revision,
		Payload:    payload.Event,
	}
	globalSequence := s.eventStore.Append(event)
	return conn.Send(connection.Appended(payload.ID, globalSequence))
}
