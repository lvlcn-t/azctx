package cmd

import (
	"fmt"

	"github.com/lvlcn-t/azctx/config"
	"github.com/lvlcn-t/azctx/output"
	"github.com/spf13/cobra"
)

// getCmd prints details for one context.
var getCmd = &cobra.Command{
	Use:               "get [name]",
	Aliases:           []string{"get-context"},
	Short:             "Get information about a context",
	Long:              "Get information about one context. When NAME is omitted, it returns the current context.",
	Example:           "  azctx get\n  azctx get prod -o json\n  azctx get prod -o table",
	RunE:              runGet,
	DisableAutoGenTag: true,
	Args:              cobra.MaximumNArgs(1),
}

func init() { //nolint:gochecknoinits // Cobra command setup
	bindOutputFlag(getCmd)
}

// runGet executes the get command.
func runGet(cmd *cobra.Command, args []string) error {
	loaded, err := config.Load()
	if err != nil {
		return err
	}

	format, err := outputFormat(cmd)
	if err != nil {
		return err
	}

	var contextName string
	if len(args) == 1 {
		contextName = args[0]
	} else {
		currentContextName, currentErr := mustCurrentContextName(&loaded.Config)
		if currentErr != nil {
			return currentErr
		}

		contextName = currentContextName
	}

	context, found := loaded.Config.ContextByName(contextName)
	if !found {
		return fmt.Errorf("context %q not found", contextName)
	}

	view := buildContextView(&loaded.Config, context, loaded.Config.CurrentContext)

	switch format {
	case output.FormatText:
		_, writeErr := fmt.Fprintln(cmd.OutOrStdout(), contextViewText(&view))
		return writeErr
	case output.FormatJSON:
		return output.PrintJSON(cmd.OutOrStdout(), view)
	case output.FormatTable:
		return output.PrintTable(
			cmd.OutOrStdout(),
			[]string{"CURRENT", "NAME", "TENANT", "TENANT ID", "CREDENTIAL", "TYPE", "SUBSCRIPTION"},
			[][]string{contextTableRow(&view)},
		)
	default:
		return fmt.Errorf("unsupported output format %q", format)
	}
}
