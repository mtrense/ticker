package main

import (
	"strconv"
	"strings"

	"github.com/spf13/viper"
	"google.golang.org/grpc"

	"github.com/mtrense/ticker/client"
	"github.com/mtrense/ticker/eventstream/base"
	"github.com/spf13/cobra"
)

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

func clientConnect() *grpc.ClientConn {
	connect := viper.GetString("connect")
	if conn, err := grpc.Dial(connect, grpc.WithInsecure()); err != nil {
		panic(err)
	} else {
		return conn
	}
	return nil
}
