package main

import (
	"context"
	"fmt"
	"syscall"
	"time"

	es "github.com/mtrense/ticker/eventstream"

	"github.com/mtrense/ticker/support"

	"github.com/mtrense/soil/logging"

	"github.com/mtrense/ticker/rpc"
	"github.com/mtrense/ticker/server"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	. "github.com/mtrense/soil/config"
	"github.com/spf13/cobra"
)

var (
	version = "none"
	commit  = "none"
	app     = NewCommandline("ticker",
		SubCommand("server",
			Short("Run the ticker server"),
			Flag("listen", Str(":6677"), Description("Address to listen for grpc connections"), Mandatory(), Persistent(), Env()),
			Flag("database", Str("localhost:5432"), Description("Database server to connect to"), Mandatory(), Persistent(), Env()),
			Run(executeServer),
		),
		SubCommand("client",
			Short("Run the ticker client"),
			Flag("connect", Str("localhost:6677"), Description("Server to connect to"), Mandatory(), Persistent(), Env()),
			SubCommand("emit",
				Short("Emit specified event"),
				Run(executeClientEmit),
			),
			SubCommand("sample",
				Short("Emit sample events"),
				Run(executeClientSample),
			),
			SubCommand("subscribe",
				Short("Subscribe to a specific event stream"),
				Run(executeClientSubscribe),
			),
			SubCommand("metrics",
				Short("Show live metrics of the ticker server"),
				Run(executeClientMetrics),
			),
		),
		Version(version, commit),
		Completion(),
	).GenerateCobra()
)

func init() {
	EnvironmentConfig("TICKER")
	ApplyLogFlags(app)
	logging.ConfigureLogging("info", "-")
}

func main() {
	if err := app.Execute(); err != nil {
		panic(err)
	}
}

func executeServer(cmd *cobra.Command, args []string) {
	listen := viper.GetString("listen")
	stream := es.NewMemoryEventStream(es.NewMemorySequenceStore())
	srv := server.NewServer(listen, version, stream)
	if err := srv.Start(); err != nil {
		panic(err)
	}
}

func executeClientEmit(cmd *cobra.Command, args []string) {
	conn := clientConnect()
	_ = rpc.NewEventStreamClient(conn)
}

func executeClientSample(cmd *cobra.Command, args []string) {

}

func executeClientSubscribe(cmd *cobra.Command, args []string) {

}

func executeClientMetrics(cmd *cobra.Command, args []string) {
	conn := clientConnect()
	admin := rpc.NewMaintenanceClient(conn)
	ctx := support.CancelContextOnSignals(context.Background(), syscall.SIGINT)
	for {
		if state, err := admin.GetServerState(ctx, &emptypb.Empty{}); err == nil {
			fmt.Printf("uptime: %5d   |   active connections: %3d   |   events stored: %8d\n", state.Uptime, state.ConnectionCount, state.EventCount)
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(2 * time.Second):
		}
	}
}

func clientConnect() *grpc.ClientConn {
	connect := viper.GetString("connect")
	if conn, err := grpc.Dial(connect, grpc.WithInsecure()); err != nil {
		panic(err)
	} else {
		return conn
	}
	return nil
}
