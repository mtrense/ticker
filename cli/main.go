package main

import (
	"context"
	"fmt"

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
			Flag("connect", Str(""), Description("Server to connect to"), Mandatory(), Persistent(), Env()),
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
	srv := server.NewServer(listen, version)
	if err := srv.Start(); err != nil {
		panic(err)
	}
}

func executeClientMetrics(cmd *cobra.Command, args []string) {
	connect := viper.GetString("connect")
	if conn, err := grpc.Dial(connect, grpc.WithInsecure()); err != nil {
		panic(err)
	} else {
		admin := rpc.NewAdminClient(conn)
		ctx, cancel := context.WithCancel(context.Background())
		num := 0
		if state, err := admin.GetServerState(ctx, &emptypb.Empty{}); err != nil {
			panic(err)
		} else {
			fmt.Printf("uptime: %d\nactive connections: %d\n", state.Uptime, state.ConnectionCount)
			num += 1
			if num > 10 {
				cancel()
			}
		}
	}
}
