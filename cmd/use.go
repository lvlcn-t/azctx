package cmd

import "github.com/spf13/cobra"

var useCmd = &cobra.Command{
	Use:   "use",
	Short: "Set the active Azure context",
	RunE:  runUse,
}

func runUse(cmd *cobra.Command, args []string) error {
	return nil
}
