package server

import (
	"context"
	"time"

	"github.com/mtrense/soil/logging"
	"github.com/mtrense/ticker/rpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/types/known/emptypb"
)

type maintenanceServer struct {
	rpc.UnimplementedMaintenanceServer
	server *Server
}

func (a *maintenanceServer) GetServerState(ctx context.Context, empty *emptypb.Empty) (*rpc.ServerState, error) {
	if p, ok := peer.FromContext(ctx); ok {
		logging.L().Debug().Str("peerAddr", p.Addr.String()).Msg("GetServerState()")
	}
	s := rpc.ServerState{
		Uptime:          int64(time.Since(a.server.startTime).Seconds()),
		ConnectionCount: uint32(a.server.connectionCount),
		EventCount:      a.server.streamBackend.LastSequence(),
	}
	return &s, nil
}

func (a *maintenanceServer) Shutdown(ctx context.Context, parameters *rpc.ShutdownParameters) (*emptypb.Empty, error) {
	return nil, nil
}
