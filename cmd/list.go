package cmd

import "github.com/spf13/cobra"

var listCmd = &cobra.Command{
	Use:               "list",
	Short:             "List all available Azure contexts",
	RunE:              runList,
	DisableAutoGenTag: true,
}

func runList(cmd *cobra.Command, args []string) error {
	return nil
}
