package cmd

import (
	"fmt"

	"github.com/lvlcn-t/azctx/config"
	"github.com/lvlcn-t/azctx/output"
	"github.com/spf13/cobra"
)

// listCmd lists all configured contexts.
var listCmd = &cobra.Command{
	Use:               "list",
	Aliases:           []string{"get-contexts"},
	Short:             "List all available Azure contexts",
	Long:              "List all available contexts from the merged azctx config.",
	Example:           "  azctx list\n  azctx list -o table\n  azctx list -o json",
	RunE:              runList,
	DisableAutoGenTag: true,
}

func init() { //nolint:gochecknoinits // Cobra command setup
	bindOutputFlag(listCmd)
}

// runList executes the list command.
func runList(cmd *cobra.Command, _ []string) error {
	loaded, err := config.Load()
	if err != nil {
		return err
	}

	format, err := outputFormat(cmd)
	if err != nil {
		return err
	}

	views := make([]contextView, 0, len(loaded.Config.Contexts))
	for _, context := range loaded.Config.Contexts {
		views = append(views, buildContextView(&loaded.Config, context, loaded.Config.CurrentContext))
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
			[]string{"CURRENT", "NAME", "TENANT", "TENANT ID", "CREDENTIAL", "TYPE", "SUBSCRIPTION"},
			rows,
		)
	default:
		return fmt.Errorf("unsupported output format %q", format)
	}
}
