package cmd

import (
	"fmt"

	"github.com/lvlcn-t/azctx/config"
	"github.com/lvlcn-t/azctx/tui"
	"github.com/lvlcn-t/azctx/tui/state"
	"github.com/spf13/cobra"
)

// AzCtx is the root command for the azctx CLI tool.
var AzCtx = &cobra.Command{
	Use:               "azctx",
	Short:             ShortDescription,
	Long:              Description,
	Example:           Example,
	Version:           Version,
	DisableAutoGenTag: true,
	SilenceUsage:      true,
	RunE: func(cmd *cobra.Command, _ []string) error {
		loader := config.NewLoader()

		// We purposely load the config before showing the TUI, so that we can
		// reduce the time between the user making a selection and the context actually
		// switching. The TUI should be snappy, and loading the config can take a
		// few milliseconds, especially if there are many contexts or if the config file
		// is on a slow disk.
		store, err := loader.Load()
		if err != nil {
			return err
		}

		choice, err := tui.Run(&store, state.ModeInteractive)
		if err != nil {
			return err
		}
		if choice == "" {
			return nil
		}

		switcher := newContextSwitcher()
		if err = switcher.switchContext(cmd.Context(), &store, choice); err != nil {
			return err
		}

		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Switched to context %q.\n", choice)
		return err
	},
}

func init() { //nolint:gochecknoinits // This is the standard way to set up Cobra commands
	AzCtx.AddCommand(newCurrentCmd())
	AzCtx.AddCommand(newDeleteCtxCmd())
	AzCtx.AddCommand(newGetCmd())
	AzCtx.AddCommand(newListCmd())
	AzCtx.AddCommand(newRenameCtxCmd())
	AzCtx.AddCommand(newSetCtxCmd())
	AzCtx.AddCommand(newSetCredentialCmd())
	AzCtx.AddCommand(newSetTenantCmd())
	AzCtx.AddCommand(newUseCmd())
	AzCtx.AddCommand(newViewCmd())
}
