package cmd

import "github.com/spf13/cobra"

// AzCtx is the root command for the azctx CLI tool.
var AzCtx = &cobra.Command{
	Use:               "azctx",
	Short:             ShortDescription,
	Long:              Description,
	Example:           Example,
	Version:           Version,
	DisableAutoGenTag: true,
	SilenceUsage:      true,
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
