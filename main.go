package main

import (
	"cmp"
	"fmt"
	"os"

	"github.com/lvlcn-t/azctx/cmd"
)

// version is set at build time using ldflags="-X main.version=$(VERSION)"
var version string

// main configures build metadata and executes the root command.
func main() {
	cmd.AzCtx.Version = cmp.Or(version, cmd.Version)

	if err := cmd.AzCtx.Execute(); err != nil {
		fmt.Fprintln(cmd.AzCtx.ErrOrStderr(), err)
		os.Exit(1)
	}
}
