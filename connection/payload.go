package connection

import (
	"time"
)

type MsgServerInfo struct {
	ServerVersion   string `codec:"server_version,omitempty"`
	ProtocolVersion int    `codec:"protocol_version,omitempty"`
}

func ServerInfo(serverVersion string, protocolVersion int) *Message {
	return Msg(MsgTypeServerInfo, MsgServerInfo{
		ServerVersion:   serverVersion,
		ProtocolVersion: protocolVersion,
	})
}

type MsgConnect struct {
	ClientName      string `codec:"client_name,omitempty"`
	ProtocolVersion int    `codec:"protocol_version,omitempty"`
}

func Connect(clientName string, protocolVersion int) *Message {
	return Msg(MsgTypeConnect, MsgConnect{
		ClientName:      clientName,
		ProtocolVersion: protocolVersion,
	})
}

type MsgError struct {
	Details string `codec:"details,omitempty"`
}

func Error(details string) *Message {
	return Msg(MsgTypeError, MsgError{
		Details: details,
	})
}

type MsgAppend struct {
	ID         string                 `codec:"id,omitempty"`
	Aggregate  []string               `codec:"aggregate,omitempty"`
	EventType  string                 `codec:"event_type,omitempty"`
	OccurredAt time.Time              `codec:"occurred_at,omitempty"`
	Revision   int                    `codec:"revision,omitempty"`
	Event      map[string]interface{} `codec:"event,omitempty"`
}

type MsgAppended struct {
	ID             string `codec:"id,omitempty"`
	GlobalSequence int    `codec:"global_sequence,omitempty"`
}

func Appended(id string, globalSequence int) *Message {
	return Msg(MsgTypeAppended, &MsgAppended{
		ID:             id,
		GlobalSequence: globalSequence,
	})
}

type MsgAck struct {
	ID string `codec:"id,omitempty"`
}

func Close() *Message {
	return Msg(MsgTypeClose, nil)
}
