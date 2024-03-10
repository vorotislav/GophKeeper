package signals

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

type OSSignals struct {
	ctx context.Context
	ch  chan os.Signal
}

func NewOSSignals(ctx context.Context) OSSignals {
	return OSSignals{
		ctx: ctx,
		ch:  make(chan os.Signal, 1),
	}
}

func (oss *OSSignals) Subscribe(onSignal func(signal os.Signal)) {
	signals := []os.Signal{
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
	}

	signal.Notify(oss.ch, signals...)

	go func(ch <-chan os.Signal) {
		select {
		case <-oss.ctx.Done():
			break
		case sig, opened := <-ch:
			if oss.ctx.Err() != nil {
				break
			}

			if opened && sig != nil {
				onSignal(sig)
			}
		}
	}(oss.ch)
}

func (oss *OSSignals) Stop() {
	signal.Stop(oss.ch)
	close(oss.ch)
}
