package cmd

import (
	"fmt"

	"github.com/lvlcn-t/azctx/config"
	"github.com/lvlcn-t/azctx/output"
	"github.com/lvlcn-t/azctx/tui"
	"github.com/spf13/cobra"
)

type listCmd struct {
	loader config.Loader
}

// newListCmd lists all configured contexts.
func newListCmd() *cobra.Command {
	c := &listCmd{loader: config.NewLoader()}
	cmd := &cobra.Command{ //nolint:exhaustruct // Cobra command definition
		Use:               "list",
		Aliases:           []string{"get-contexts"},
		Short:             "List all available Azure contexts",
		Long:              "List all available contexts from the merged azctx config.",
		Example:           "  azctx list\n  azctx list -o table\n  azctx list -o json",
		RunE:              c.run,
		DisableAutoGenTag: true,
	}

	bindOutputFlag(cmd)

	return cmd
}

// run executes the list command.
func (c *listCmd) run(cmd *cobra.Command, _ []string) error {
	// If no explicit -o flag, launch interactive TUI.
	if !cmd.Flags().Changed("output") {
		_, tuiErr := tui.Run(c.loader, tui.ModeBrowse)
		return tuiErr
	}

	store, err := c.loader.Load()
	if err != nil {
		return err
	}

	format, err := outputFormat(cmd)
	if err != nil {
		return err
	}

	views := make([]contextView, 0, len(store.Config.Contexts))
	for _, context := range store.Config.Contexts {
		views = append(views, buildContextView(&store.Config, context, store.Config.CurrentContext))
	}

	switch format {
	case output.FormatText:
		for _, view := range views {
			prefix := " "
			if view.Current {
				prefix = "*"
			}

			if _, writeErr := fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", prefix, view.Name); writeErr != nil {
				return writeErr
			}
		}

		return nil
	case output.FormatJSON:
		return output.PrintJSON(cmd.OutOrStdout(), views)
	case output.FormatTable:
		rows := make([][]string, 0, len(views))
		for index := range views {
			rows = append(rows, contextTableRow(&views[index]))
		}

		return output.PrintTable(
			cmd.OutOrStdout(),
			[]string{"CURRENT", "NAME", "TENANT", "TENANT ID", "CREDENTIAL", "TYPE", "SUBSCRIPTION"}, //nolint:goconst // column headers
			rows,
		)
	default:
		return fmt.Errorf("unsupported output format %q", format)
	}
}
