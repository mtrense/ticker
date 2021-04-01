package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/mtrense/ticker/eventstream/memory"

	"github.com/mtrense/ticker/eventstream/base"

	"github.com/mtrense/ticker/client"

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
				Flag("topic", Str(""), Abbr("t"), Description("Select Topic and Type of the emitted event"), Persistent()),
				Flag("payload", Str("{}"), Abbr("p"), Description("The payload of the emitted event (- for stdin)"), Persistent()),
				Run(executeClientEmit),
			),
			SubCommand("sample",
				Short("Emit sample events"),
				Run(executeClientSample),
			),
			SubCommand("stream",
				Short("Stream a portion of the event stream"),
				Flag("format", Str("text"), Description("Format for Event output (text, json)"), Persistent()),
				Flag("omit-payload", Bool(), Description("Omit Payload in Event output"), Persistent()),
				Flag("pretty", Bool(), Description("Use pretty-mode in Event output"), Persistent()),
				Flag("selector", Str(">/*"), Abbr("s"), Description("Select which events to stream"), Persistent()),
				Flag("range", Str("1:"), Abbr("r"), Description("Select which events to stream"), Persistent()),
				Run(executeClientStream),
			),
			SubCommand("subscribe",
				Short("Subscribe to a specific event stream"),
				Flag("format", Str("text"), Description("Format for Event output (text, json)"), Persistent()),
				Flag("omit-payload", Bool(), Description("Omit Payload in Event output"), Persistent()),
				Flag("pretty", Bool(), Description("Use pretty-mode in Event output"), Persistent()),
				Flag("selector", Str(">/*"), Description("Select which events to subscribe to"), Persistent()),
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
	stream := memory.NewMemoryEventStream(memory.NewMemorySequenceStore())
	srv := server.NewServer(listen, version, stream)
	if err := srv.Start(); err != nil {
		panic(err)
	}
}

func executeClientEmit(cmd *cobra.Command, args []string) {
	payloadString, _ := cmd.Flags().GetString("payload")
	topicAndType, _ := cmd.Flags().GetString("topic")
	if selector, err := base.ParseSelector(topicAndType); err == nil {
		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(payloadString), &payload); err != nil {
			panic(err)
		}
		cl := client.NewClient(clientConnect())
		ctx := support.CancelContextOnSignals(context.Background(), syscall.SIGINT)
		event := base.Event{
			Aggregate:  selector.Aggregate,
			Type:       selector.Type,
			OccurredAt: time.Now(),
			Payload:    payload,
		}
		if _, err := cl.Emit(ctx, event); err != nil {
			panic(err)
		} else {

		}
	} else {
		panic(err)
	}
}

func executeClientSample(cmd *cobra.Command, args []string) {

}

func executeClientStream(cmd *cobra.Command, args []string) {
	formatter := createFormatter(cmd)
	cl := client.NewClient(clientConnect())
	ctx := support.CancelContextOnSignals(context.Background(), syscall.SIGINT)
	err := cl.Stream(ctx, selectorFromFlags(cmd), bracketFromFlags(cmd), func(e *base.Event) {
		if err := formatter(os.Stdout, e); err != nil {
			panic(err)
		}
	})
	if err != nil {
		panic(err)
	}
}

func executeClientSubscribe(cmd *cobra.Command, args []string) {
	//formatter := createFormatter(cmd)
	cl := client.NewClient(clientConnect())
	ctx := support.CancelContextOnSignals(context.Background(), syscall.SIGINT)
	cl.Subscribe(ctx)
}

func selectorFromFlags(cmd *cobra.Command) *base.Selector {
	sel, _ := cmd.Flags().GetString("selector")
	if selector, err := base.ParseSelector(sel); err == nil {
		return selector
	}
	return nil
}

func bracketFromFlags(cmd *cobra.Command) *base.Bracket {
	rang, err := cmd.Flags().GetString("range")
	if err != nil {
		panic(err)
	}
	ranges := strings.Split(rang, ":")
	first, err := strconv.ParseInt(ranges[0], 10, 64)
	if err != nil {
		first = 0
	}
	last, err := strconv.ParseInt(ranges[1], 10, 64)
	if err != nil {
		last = -1
	}
	return &base.Bracket{
		NextSequence: first,
		LastSequence: last,
	}
}

func createFormatter(cmd *cobra.Command) client.Formatter {
	format, _ := cmd.Flags().GetString("format")
	omitPayload, _ := cmd.Flags().GetBool("omit-payload")
	pretty, _ := cmd.Flags().GetBool("pretty")
	var formatter client.Formatter
	switch strings.ToLower(format) {
	case "json":
		formatter = client.JsonFormatter(pretty)
	default:
		formatter = client.TextFormatter(pretty)
	}
	if omitPayload {
		formatter = client.OmitPayload(formatter)
	}
	return formatter
}

func executeClientMetrics(cmd *cobra.Command, args []string) {
	conn := clientConnect()
	admin := rpc.NewMaintenanceClient(conn)
	ctx := support.CancelContextOnSignals(context.Background(), syscall.SIGINT)
	for {
		if state, err := admin.GetServerState(ctx, &emptypb.Empty{}); err == nil {
			fmt.Printf("uptime: %5ds   |   active connections: %3d   |   events stored: %8d\n", state.Uptime, state.ConnectionCount, state.EventCount)
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
