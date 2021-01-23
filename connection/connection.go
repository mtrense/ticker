package connection

import (
	"bufio"
	"fmt"
	"io"
	"net"

	log "github.com/mtrense/soil/logging"
)

type Subscription struct {
	topicSegments []string
}

type Connection struct {
	server        *TickerServer
	conn          net.Conn
	subscriptions []Subscription
}

func (s *Connection) RemoteAddr() net.Addr {
	return s.conn.RemoteAddr()
}

func (s *Connection) HasSubscriptions() bool {
	return s.subscriptions != nil && len(s.subscriptions) > 0
}

func (s *Connection) handleConnection() {
	reader := bufio.NewReader(s.conn)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.L().Info().Str("remoteAddress", s.conn.RemoteAddr().String()).Msg("Client has closed connection")
			} else {
				log.L().Err(err).Msg("Error while reading from client connection")
			}
			s.close()
			break
		}
		fmt.Printf("%v\n", line)
	}
}

func (s *Connection) close() {
	s.conn.Close()
	s.server.quitConnection(s)
}
