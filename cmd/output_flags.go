package cmd

import (
	"fmt"

	"github.com/lvlcn-t/azctx/output"
	"github.com/spf13/cobra"
)

// bindOutputFlag adds the shared output format flag to a command.
func bindOutputFlag(command *cobra.Command) {
	command.Flags().StringP(
		"output",
		"o",
		string(output.FormatText),
		"Output format. One of: text|table|json",
	)
}

// outputFormat resolves and validates the selected output format flag value.
func outputFormat(command *cobra.Command) (output.Format, error) {
	rawFormat, err := command.Flags().GetString("output")
	if err != nil {
		return "", fmt.Errorf("read output flag: %w", err)
	}

	return output.ParseFormat(rawFormat)
}
