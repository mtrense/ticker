package main

import (
	"github.com/mtrense/soil"
	. "github.com/mtrense/soil/config"
	"github.com/mtrense/ticker/connection"
	"github.com/spf13/cobra"
)

var (
	version = "none"
	commit  = "none"
	app     = NewCommandline("ticker",
		SubCommand("server",
			Short("Run the ticker server"),
			Run(executeServer),
		),
		Version(version, commit),
		Completion(),
	).GenerateCobra()
)

func init() {
	EnvironmentConfig("TICKER")
	ApplyLogFlags(app)
	//soil.DefaultCLI(app, version, commit, "TICKER")
	soil.ConfigureDefaultLogging()
}

func main() {
	if err := app.Execute(); err != nil {
		panic(err)
	}
}

func executeServer(cmd *cobra.Command, args []string) {
	server := connection.NewServer(":6677")
	server.Start()
}
