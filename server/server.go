package server

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"

	"github.com/mtrense/ticker/connection"
	"github.com/mtrense/ticker/eventstore"

	log "github.com/mtrense/soil/logging"
)

var (
	ErrProtocolVersionUnsupported = fmt.Errorf("protocol version unsupported")
)

type TickerServer struct {
	version     string
	listen      string
	connections map[*connection.Connection]*ConnectionState
	handlers    map[string]ServerMessageHandler
	eventStore  eventstore.EventStore
	listener    net.Listener
	stopping    chan interface{}
	signals     chan os.Signal
	wg          sync.WaitGroup
}

type ServerInfo struct {
	ServerVersion   string   `json:"server_version,omitempty"`
	ProtocolVersion int      `json:"protocol_version,omitempty"`
	ValidCommands   []string `json:"valid_commands,omitempty"`
}

type ConnectionState struct {
	ClientName          string
	ProtocolVersion     int
	subscribedAggregate []string
	selectedType        string
}

type ServerMessageHandler func(conn *connection.Connection, state *ConnectionState, msg connection.Message) error

func NewServer(addr, version string, eventStore eventstore.EventStore) *TickerServer {
	srv := &TickerServer{
		version:     version,
		listen:      addr,
		connections: make(map[*connection.Connection]*ConnectionState),
		handlers:    make(map[string]ServerMessageHandler),
		eventStore:  eventStore,
		stopping:    make(chan interface{}),
		signals:     make(chan os.Signal),
	}
	srv.registerHandlers()
	return srv
}

func (s *TickerServer) registerHandlers() {
	s.handlers["CONNECT"] = s.handleConnect
	s.handlers["APPEND"] = s.handleAppend
}

func (s *TickerServer) Run() error {
	log.L().Info().Str("address", s.listen).Msg("Server starting")
	var err error
	s.listener, err = net.Listen("tcp", s.listen)
	if err != nil {
		return err
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, os.Kill)
	go func() {
		sig := <-sigs
		switch sig {
		case os.Kill:
			// Handle hard Kill
			// TODO
		case os.Interrupt:
			// Handle Ctrl-C
			s.Stop()
		}
	}()
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.stopping:
				return nil
			default:
				log.L().Err(err).Msg("MsgError while accepting socket connection")
			}
		} else {
			s.addConnection(conn)
			log.L().Info().Int("activeConnections", len(s.connections)).Str("remoteAddress", conn.RemoteAddr().String()).Msg("Client connected")
		}
	}
}

func (s *TickerServer) Stop() {
	log.L().Info().Int("numberOfConnections", len(s.connections)).Msg("Gracefully shutting down")
	close(s.stopping)
	s.listener.Close()
	for conn, _ := range s.connections {
		conn.Send(connection.Close())
	}
	s.wg.Wait()
}

func (s *TickerServer) handleMessage(conn *connection.Connection, msg connection.Message) error {
	state := s.connections[conn]
	if handler, ok := s.handlers[msg.Type]; ok {
		if err := handler(conn, state, msg); err != nil {
			conn.Send(connection.Error(err.Error()))
		}
	} else {
		log.L().Warn().Str("remoteAddress", conn.RemoteAddr().String()).Str("type", msg.Type).Msg("Unknown message type")
		conn.Send(connection.Error("Unknown message type"))
	}
	return nil
}

func (s *TickerServer) ConnectionClosing(conn *connection.Connection) {
	//conn.Send("CLOSE", nil)
}

func (s *TickerServer) ConnectionClosed(conn *connection.Connection) {
	delete(s.connections, conn)
}

func (s *TickerServer) ConnectionOpened(conn *connection.Connection) {
	conn.Send(connection.ServerInfo(s.version, 1))
}

func (s *TickerServer) addConnection(c net.Conn) *connection.Connection {
	connection := connection.NewConnection(s, s.handleMessage, c)
	s.connections[connection] = &ConnectionState{}
	s.wg.Add(1)
	go func() {
		connection.Handle()
		s.wg.Done()
	}()
	return connection
}

func (s *ConnectionState) HasSubscriptions() bool {
	return s.subscribedAggregate != nil && len(s.subscribedAggregate) > 0
}
