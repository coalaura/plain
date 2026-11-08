package internal

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func WaitForInterrupt(ctx context.Context) {
	sCtx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	<-sCtx.Done()
}
