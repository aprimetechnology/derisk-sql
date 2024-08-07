package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/aprimetechnology/derisk-sql/internal/cmd"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := cmd.RootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
