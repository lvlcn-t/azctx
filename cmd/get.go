package cmd

import "github.com/spf13/cobra"

var getCmd = &cobra.Command{
	Use:               "get",
	Short:             "Get information about a context",
	RunE:              runGet,
	DisableAutoGenTag: true,
}

func runGet(cmd *cobra.Command, args []string) error {
	return nil
}
