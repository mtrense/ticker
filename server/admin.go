package server

import (
	"context"
	"time"

	"github.com/mtrense/soil/logging"
	"github.com/mtrense/ticker/rpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/types/known/emptypb"
)

type adminServer struct {
	rpc.UnimplementedAdminServer
	server *Server
}

func (a *adminServer) GetServerState(ctx context.Context, empty *emptypb.Empty) (*rpc.ServerState, error) {
	if p, ok := peer.FromContext(ctx); ok {
		logging.L().Info().Str("peerAddr", p.Addr.String()).Msg("Initializing")
	}
	s := rpc.ServerState{
		Uptime:          int64(time.Since(a.server.startTime).Seconds()),
		ConnectionCount: uint32(a.server.connectionCount),
	}
	return &s, nil
}

func (a *adminServer) Shutdown(ctx context.Context, parameters *rpc.ShutdownParameters) (*emptypb.Empty, error) {
	return nil, nil
}
