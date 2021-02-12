package main

import (
	"fmt"
	"os"

	. "github.com/mtrense/soil/config"
	l "github.com/mtrense/soil/logging"
	"github.com/mtrense/ticker/client"
	"github.com/mtrense/ticker/eventstore"
	"github.com/mtrense/ticker/server"
	"github.com/spf13/cobra"
)

var (
	version = "none"
	commit  = "none"
	app     = NewCommandline("ticker",
		Short("Ticker EventStream server"),
		FlagLogFile(),
		FlagLogLevel("warn"),
		FlagLogFormat(),
		SubCommand("server",
			Short("Run the ticker server"),
			Run(executeServer),
		),
		SubCommand("client",
			Short("Run client"),
			Flag("client-id", Str(defaultClientId()),
				Description("Client ID to send to the server when connecting")),
			Flag("server", Str("localhost:6677"),
				Description("Address of the Server to connect to")),
			//Run(executeClient),
			SubCommand("shell",
				Alias("repl", "sh"),
				Short("Run a client Shell"),
				Run(executeClientShell),
			),
		),
		Version(version, commit),
		Completion(),
	).GenerateCobra()
)

func init() {
	EnvironmentConfig("TICKER")
	l.ConfigureDefaultLogging()
}

func main() {
	if err := app.Execute(); err != nil {
		panic(err)
	}
}

func executeServer(cmd *cobra.Command, args []string) {
	eventStore := eventstore.NewMemoryEventStore()
	srv := server.NewServer(":6677", version, eventStore)
	l.L().Info().Msg("Server starting")

	if err := srv.Run(); err != nil {
		panic(err)
	}
}

//func executeClient(cmd *cobra.Command, args []string) {
//	cln := client.NewClient("client1", "127.0.0.1:6677")
//	cln.Connect()
//
//}

func executeClientShell(cmd *cobra.Command, args []string) {

	cln := client.NewClient("client1", "127.0.0.1:6677")
	//go cln.Connect()
	pr := cln.NewPrompt()
	pr.Run()
}

func defaultClientId() string {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "localhost"
	}
	return fmt.Sprintf("%s-%d", hostname, os.Getpid())
}
