package connection

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/ugorji/go/codec"

	log "github.com/mtrense/soil/logging"
)

var (
	ErrChecksumNotMatching = fmt.Errorf("checksum is not matching")
	ErrAlreadyStarted      = fmt.Errorf("connection has already been started")
	codecHandle            = &codec.MsgpackHandle{}
)

type MessageHandler func(conn *Connection, msg Message) error

type StateListener interface {
	ConnectionOpened(conn *Connection)
	ConnectionClosing(conn *Connection)
	ConnectionClosed(conn *Connection)
}

type Connection struct {
	// Handlers & Network
	stateListener  StateListener
	messageHandler MessageHandler
	conn           net.Conn
	// IO
	reader  *bufio.Reader
	decoder *codec.Decoder
	writer  *bufio.Writer
	encoder *codec.Encoder
	// Connection State
	started      bool
	localSerial  int
	remoteSerial int
	writeMonitor sync.Mutex
}

func init() {
	codecHandle.WriteExt = true
	codecHandle.Raw = true
}

func NewConnection(stateListener StateListener, handler MessageHandler, c net.Conn) *Connection {
	conn := &Connection{
		stateListener:  stateListener,
		messageHandler: handler,
		conn:           c,
		localSerial:    0,
		remoteSerial:   0,
	}

	return conn
}

func (s *Connection) RemoteAddr() net.Addr {
	return s.conn.RemoteAddr()
}

func (s *Connection) Handle() error {
	if !s.started {
		s.started = true
		s.handleConnection()
		return nil
	} else {
		return ErrAlreadyStarted
	}
}

func (s *Connection) handleConnection() {
	codecHandle.Raw = true
	defer s.close()
	s.reader = bufio.NewReader(s.conn)
	s.writer = bufio.NewWriter(s.conn)
	s.decoder = codec.NewDecoder(s.reader, codecHandle)
	s.encoder = codec.NewEncoder(s.writer, codecHandle)
	s.stateListener.ConnectionOpened(s)
	for {
		var msg Message
		if err := s.decoder.Decode(&msg); err != nil {
			if err == io.EOF {
				log.L().Info().Str("remoteAddress", s.conn.RemoteAddr().String()).Msg("Remote peer has closed connection")
			} else {
				log.L().Err(err).Str("remoteAddress", s.conn.RemoteAddr().String()).Msg("MsgError while reading from remote connection")
			}
			break
		}
		if err := s.messageHandler(s, msg); err != nil {
			log.L().Err(err).Str("remoteAddress", s.conn.RemoteAddr().String()).Int("remoteSerial", msg.Serial).Msg("MsgError while handling Message")
			break
		}
	}
}

func (s *Connection) Send(msg *Message) error {
	s.writeMonitor.Lock()
	defer s.writeMonitor.Unlock()
	s.localSerial += 1
	msg.Serial = s.localSerial
	if err := s.encoder.Encode(msg); err != nil {
		log.L().Err(err).Msg("Failed to encode message")
		return err
	}
	s.writer.Flush()
	return nil
}

func (s *Connection) close() {
	s.stateListener.ConnectionClosing(s)
	s.conn.Close()
	s.stateListener.ConnectionClosed(s)
}
