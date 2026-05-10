package cmd

import (
	"context"
	"fmt"

	"github.com/lvlcn-t/azctx/az"
	"github.com/lvlcn-t/azctx/config"
	"github.com/lvlcn-t/azctx/output"
	"github.com/spf13/cobra"
)

type getCmd struct {
	loader config.Loader
	az     func(ctx context.Context) (az.CLI, error)
}

// newGetCmd prints details for one context.
func newGetCmd() *cobra.Command {
	command := &getCmd{
		loader: config.NewLoader(),
		az:     az.NewClient,
	}

	cmd := &cobra.Command{ //nolint:exhaustruct // Cobra command definition
		Use:               "get [name]",
		Aliases:           []string{"get-context"},
		Short:             "Get information about a context",
		Long:              "Get information about one context. When NAME is omitted, it returns the current context.",
		Example:           "  azctx get\n  azctx get prod -o json\n  azctx get prod -o table",
		RunE:              command.run,
		DisableAutoGenTag: true,
		Args:              cobra.MaximumNArgs(1),
	}

	bindOutputFlag(cmd)

	return cmd
}

// run executes the get command.
func (c *getCmd) run(cmd *cobra.Command, args []string) error {
	store, err := c.loader.Load()
	if err != nil {
		return err
	}

	format, err := outputFormat(cmd)
	if err != nil {
		return err
	}

	var name string
	if len(args) == 1 {
		name = args[0]
	} else {
		name, err = mustCurrentContextName(&store.Config)
		if err != nil {
			return err
		}
	}

	ctx, found := store.Config.ContextByName(name)
	if !found {
		return fmt.Errorf("context %q not found", name)
	}

	view := buildContextView(&store.Config, ctx, store.Config.CurrentContext)

	switch format {
	case output.FormatText:
		_, err = fmt.Fprintln(cmd.OutOrStdout(), contextViewText(&view))
		return err
	case output.FormatJSON:
		return output.PrintJSON(cmd.OutOrStdout(), view)
	case output.FormatTable:
		return output.PrintTable(
			cmd.OutOrStdout(),
			[]string{"CURRENT", "NAME", "TENANT", "TENANT ID", "CREDENTIAL", "TYPE", "SUBSCRIPTION"}, //nolint:goconst // column headers
			[][]string{contextTableRow(&view)},
		)
	default:
		return fmt.Errorf("unsupported output format %q", format)
	}
}
