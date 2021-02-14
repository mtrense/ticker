package server

import (
	"context"
	"net"
	"os"
	"os/signal"
	"sync/atomic"
	"time"

	"google.golang.org/grpc/stats"

	"google.golang.org/grpc/peer"

	"github.com/mtrense/soil/logging"
	"google.golang.org/grpc/reflection"

	"github.com/mtrense/ticker/rpc"
	"google.golang.org/grpc"
)

type Server struct {
	listen          string
	version         string
	stream          *eventStreamServer
	admin           *adminServer
	connectionCount int32
	startTime       time.Time
}

func NewServer(listen string, version string) *Server {
	srv := &Server{
		listen:  listen,
		version: version,
	}
	srv.stream = &eventStreamServer{
		server: srv,
	}
	srv.admin = &adminServer{
		server: srv,
	}
	return srv
}

func (s *Server) Start() error {
	s.startTime = time.Now()
	sigs := make(chan os.Signal)
	signal.Notify(sigs, os.Interrupt, os.Kill)
	listener, err := net.Listen("tcp", ":6677")
	if err != nil {
		return err
	}
	srv := grpc.NewServer(grpc.StatsHandler(s))
	go func() {
		sig := <-sigs
		switch sig {
		case os.Kill:
			srv.Stop()
		case os.Interrupt:
			srv.GracefulStop()
		}
	}()
	rpc.RegisterEventStreamServer(srv, s.stream)
	rpc.RegisterAdminServer(srv, s.admin)
	reflection.Register(srv)
	return srv.Serve(listener)
}

func (s *Server) TagRPC(ctx context.Context, i *stats.RPCTagInfo) context.Context {
	return ctx
}

func (s *Server) HandleRPC(ctx context.Context, st stats.RPCStats) {
}

func (s *Server) TagConn(ctx context.Context, i *stats.ConnTagInfo) context.Context {
	return ctx
}

func (s *Server) HandleConn(ctx context.Context, st stats.ConnStats) {
	l := logging.L().Info()
	if p, ok := peer.FromContext(ctx); ok {
		l.Str("clientAddr", p.Addr.String())
	}
	switch st.(type) {
	case *stats.ConnBegin:
		atomic.AddInt32(&s.connectionCount, 1)
		l.Msg("Client connected")
	case *stats.ConnEnd:
		atomic.AddInt32(&s.connectionCount, -1)
		l.Msg("Client disconnected")
	}
}
