package cmd

import (
	"fmt"

	"github.com/lvlcn-t/azctx/config"
	"github.com/lvlcn-t/azctx/contexts"
	"github.com/lvlcn-t/azctx/output"
	"github.com/spf13/cobra"
)

type currentCmd struct {
	loader config.Loader
}

// newCurrentCmd prints the currently active context.
func newCurrentCmd() *cobra.Command {
	command := &currentCmd{loader: config.NewLoader()}

	cmd := &cobra.Command{ //nolint:exhaustruct // Cobra command definition
		Use:               "current",
		Aliases:           []string{"current-context"},
		Short:             "Show the current active context",
		Long:              "Show the current active context from azctx config.",
		Example:           "  azctx current\n  azctx current -o json\n  azctx current --verbose -o table",
		RunE:              command.run,
		DisableAutoGenTag: true,
	}

	bindOutputFlag(cmd)
	cmd.Flags().BoolP("verbose", "v", false, "Include full context details")

	return cmd
}

// run executes the current command.
func (c *currentCmd) run(cmd *cobra.Command, _ []string) error {
	store, err := c.loader.Load()
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

	name, err := contexts.CurrentContextName(&store.Config)
	if err != nil {
		return err
	}

	current, found := store.Config.ContextByName(name)
	if !found {
		return fmt.Errorf("context %q not found", name)
	}

	view := buildContextView(&store.Config, current, store.Config.CurrentContext)

	switch format {
	case output.FormatText:
		return c.print(cmd, view.Name, contextViewText(&view))
	case output.FormatJSON:
		return c.print(cmd, map[string]string{"name": view.Name}, view)
	case output.FormatTable:
		if !verbose {
			return output.PrintTable(
				cmd.OutOrStdout(),
				[]string{"CURRENT", "NAME"}, //nolint:goconst // column headers
				[][]string{{"*", view.Name}},
			)
		}

		return output.PrintTable(
			cmd.OutOrStdout(),
			[]string{"CURRENT", "NAME", "TENANT", "TENANT ID", "CREDENTIAL", "TYPE", "SUBSCRIPTION"}, //nolint:goconst // column headers
			[][]string{contextTableRow(&view)},
		)
	default:
		return fmt.Errorf("unsupported output format %q", format)
	}
}

func (c *currentCmd) print(cmd *cobra.Command, simple, verbose any) error {
	format, err := outputFormat(cmd)
	if err != nil {
		return err
	}

	v, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return fmt.Errorf("read verbose flag: %w", err)
	}

	switch format {
	case output.FormatText:
		if !v {
			_, err = fmt.Fprintln(cmd.OutOrStdout(), simple)
			return err
		}

		_, err = fmt.Fprintln(cmd.OutOrStdout(), verbose)
		return err
	case output.FormatJSON:
		if !v {
			return output.PrintJSON(cmd.OutOrStdout(), simple)
		}

		return output.PrintJSON(cmd.OutOrStdout(), verbose)
	default:
		return fmt.Errorf("unsupported output format %q", format)
	}
}
