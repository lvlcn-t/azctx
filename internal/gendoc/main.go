// Package main generates CLI reference documentation into docs/reference/.
//
//go:generate go run . -out ../../docs/reference
package main

import (
	"flag"
	"log"
	"os"

	"github.com/lvlcn-t/azctx/cmd"
	"github.com/spf13/cobra/doc"
)

func main() {
	out := flag.String("out", "docs/reference", "output directory for generated docs")
	flag.Parse()

	if err := os.MkdirAll(*out, 0o755); err != nil { //nolint:mnd // standard directory permission
		log.Fatal(err)
	}
	if err := doc.GenMarkdownTree(cmd.AzCtx, *out); err != nil {
		log.Fatal(err)
	}
}
