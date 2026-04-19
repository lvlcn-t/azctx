package cmd

import (
	"fmt"

	"github.com/lvlcn-t/azctx/config"
	"github.com/lvlcn-t/azctx/output"
	"github.com/spf13/cobra"
)

// currentCmd prints the currently active context.
var currentCmd = &cobra.Command{
	Use:               "current",
	Aliases:           []string{"current-context"},
	Short:             "Show the current active context",
	Long:              "Show the current active context from azctx config.",
	Example:           "  azctx current\n  azctx current -o json\n  azctx current --verbose -o table",
	RunE:              runCurrent,
	DisableAutoGenTag: true,
}

func init() { //nolint:gochecknoinits // Cobra command setup
	bindOutputFlag(currentCmd)
	currentCmd.Flags().BoolP("verbose", "v", false, "Include full context details")
}

// runCurrent executes the current command.
func runCurrent(cmd *cobra.Command, _ []string) error {
	loaded, err := config.Load()
	if err != nil {
		return err
	}

	format, err := outputFormat(cmd)
	if err != nil {
		return err
	}

	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return fmt.Errorf("read verbose flag: %w", err)
	}

	currentContextName, err := mustCurrentContextName(&loaded.Config)
	if err != nil {
		return err
	}

	currentContext, found := loaded.Config.ContextByName(currentContextName)
	if !found {
		if !verbose {
			_, writeErr := fmt.Fprintln(cmd.OutOrStdout(), currentContextName)
			return writeErr
		}

		return fmt.Errorf("context %q not found", currentContextName)
	}

	view := buildContextView(&loaded.Config, currentContext, loaded.Config.CurrentContext)

	switch format {
	case output.FormatText:
		if !verbose {
			_, writeErr := fmt.Fprintln(cmd.OutOrStdout(), view.Name)
			return writeErr
		}

		_, writeErr := fmt.Fprintln(cmd.OutOrStdout(), contextViewText(&view))
		return writeErr
	case output.FormatJSON:
		if !verbose {
			return output.PrintJSON(cmd.OutOrStdout(), map[string]string{"name": view.Name})
		}

		return output.PrintJSON(cmd.OutOrStdout(), view)
	case output.FormatTable:
		if !verbose {
			return output.PrintTable(
				cmd.OutOrStdout(),
				[]string{"CURRENT", "NAME"},
				[][]string{{"*", view.Name}},
			)
		}

		return output.PrintTable(
			cmd.OutOrStdout(),
			[]string{"CURRENT", "NAME", "TENANT", "TENANT ID", "CREDENTIAL", "TYPE", "SUBSCRIPTION"},
			[][]string{contextTableRow(&view)},
		)
	default:
		return fmt.Errorf("unsupported output format %q", format)
	}
}
