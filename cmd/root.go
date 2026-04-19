package cmd

import "github.com/spf13/cobra"

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
	AzCtx.AddCommand(getCmd)
	AzCtx.AddCommand(listCmd)
	AzCtx.AddCommand(useCmd)
}
