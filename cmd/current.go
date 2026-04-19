package cmd

import "github.com/spf13/cobra"

var currentCmd = &cobra.Command{
	Use:               "current",
	Short:             "Show the current active Azure subscription",
	RunE:              runCurrent,
	DisableAutoGenTag: true,
}

func runCurrent(cmd *cobra.Command, args []string) error {
	return nil
}
