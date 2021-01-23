package connection

import (
	"net"

	log "github.com/mtrense/soil/logging"
)

type MessageHandler func(command string, args []string)

type TickerServer struct {
	listen      string
	connections map[*Connection]struct{}
}

func NewServer(addr string) *TickerServer {
	return &TickerServer{
		listen:      addr,
		connections: make(map[*Connection]struct{}),
	}
}

func (s *TickerServer) Start() error {
	log.L().Info().Str("address", s.listen).Msg("Server starting")
	listener, err := net.Listen("tcp", s.listen)
	if err != nil {
		return err
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.L().Err(err).Msg("Error while accepting socket connection")
		}
		s.addConnection(conn)
		log.L().Info().Int("activeConnections", len(s.connections)).Str("remoteAddress", conn.RemoteAddr().String()).Msg("Client connected")
	}
	return nil
}

func (s *TickerServer) addConnection(c net.Conn) *Connection {
	connection := &Connection{
		server: s,
		conn:   c,
	}
	s.connections[connection] = struct{}{}
	go connection.handleConnection()
	return connection
}

func (s *TickerServer) quitConnection(c *Connection) {
	delete(s.connections, c)
}
