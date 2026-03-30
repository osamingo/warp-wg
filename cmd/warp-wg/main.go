package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/osamingo/warp-wg/internal/cmd"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := cmd.Run(ctx, os.Args[1:]); err != nil {
		slog.Error(err.Error()) //nolint:gosec // err is from internal cmd, not user input
		os.Exit(1)
	}
}
