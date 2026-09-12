package internal

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func WaitForInterrupt() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	<-ctx.Done()
}
