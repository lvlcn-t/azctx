package main

import (
	"cmp"
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/lvlcn-t/azctx/cmd"
)

// version is set at build time using ldflags="-X main.version=$(VERSION)"
var version string

func main() {
	os.Exit(run())
}

// run is the entry point for the CLI and returns its exit code.
func run() int {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cmd.AzCtx.Version = cmp.Or(version, cmd.Version)

	if err := cmd.AzCtx.ExecuteContext(ctx); err != nil {
		// Error is already printed by Cobra (prefixed with "Error: ").
		return 1
	}

	return 0
}
