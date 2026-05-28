package cmd

import (
	"fmt"

	"github.com/lvlcn-t/azctx/config"
	"github.com/lvlcn-t/azctx/tui"
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

		choice, err := tui.Run(loader, tui.ModeInteractive)
		if err != nil {
			return err
		}
		if choice == "" {
			return nil
		}

		store, err := loader.Load()
		if err != nil {
			return err
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
