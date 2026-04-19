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
}

func init() { //nolint:gochecknoinits // This is the standard way to set up Cobra commands
	AzCtx.AddCommand(currentCmd)
	AzCtx.AddCommand(deleteContextCmd)
	AzCtx.AddCommand(getCmd)
	AzCtx.AddCommand(listCmd)
	AzCtx.AddCommand(renameContextCmd)
	AzCtx.AddCommand(setContextCmd)
	AzCtx.AddCommand(setCredentialCmd)
	AzCtx.AddCommand(setTenantCmd)
	AzCtx.AddCommand(useCmd)
	AzCtx.AddCommand(viewCmd)
}
