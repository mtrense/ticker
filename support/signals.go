package support

import (
	"context"
	"os"
	"os/signal"
)

func CancelContextOnSignals(parent context.Context, sigs ...os.Signal) context.Context {
	ctx, cancel := context.WithCancel(parent)
	signals := make(chan os.Signal)
	signal.Notify(signals, sigs...)
	go func() {
		<-signals
		cancel()
	}()
	return ctx
}
