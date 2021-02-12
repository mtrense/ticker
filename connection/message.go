package connection

import (
	"github.com/ugorji/go/codec"
)

const (
	MsgTypeServerInfo  = "SERVER_INFO"
	MsgTypeConnect     = "CONNECT"
	MsgTypeOk          = "OK"
	MsgTypeError       = "ERROR"
	MsgTypeAppend      = "APPEND"
	MsgTypeAppended    = "APPENDED"
	MsgTypeStored      = "STORED"
	MsgTypeSubscribe   = "SUBSCRIBE"
	MsgTypeStream      = "STREAM"
	MsgTypeEvent       = "EVENT"
	MsgTypeAck         = "ACK"
	MsgTypeUnsubscribe = "UNSUBSCRIBE"
	MsgTypeClose       = "CLOSE"
)

type Message struct {
	Serial     int       `codec:"s,omitempty"`
	Type       string    `codec:"t,omitempty"`
	RawPayload codec.Raw `codec:"p,omitempty"`
}

func Msg(typ string, payload interface{}) *Message {
	msg := &Message{
		Type: typ,
	}
	msg.Pack(payload)
	return msg
}

func (s *Message) Pack(payload interface{}) error {
	var data []byte
	encoder := codec.NewEncoderBytes(&data, codecHandle)
	if err := encoder.Encode(payload); err != nil {
		return err
	}
	s.RawPayload = data
	return nil
}

func (s *Message) Unpack(payload interface{}) error {
	decoder := codec.NewDecoderBytes(s.RawPayload, codecHandle)
	return decoder.Decode(payload)
}
