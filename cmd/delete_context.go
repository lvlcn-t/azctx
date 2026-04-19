package cmd

import (
	"fmt"

	"github.com/lvlcn-t/azctx/config"
	"github.com/spf13/cobra"
)

// deleteContextCmd removes a context entry from config.
var deleteContextCmd = &cobra.Command{ //nolint:gochecknoglobals // Cobra command definition
	Use:               "delete-context NAME",
	Aliases:           []string{"unset-context"},
	Short:             "Delete a context from azctx config",
	Long:              "Delete a context from azctx config.",
	Example:           "  azctx delete-context prod",
	RunE:              runDeleteContext,
	DisableAutoGenTag: true,
	Args:              cobra.ExactArgs(1),
}

// runDeleteContext executes the delete-context command.
func runDeleteContext(cmd *cobra.Command, args []string) error {
	loaded, err := config.Load()
	if err != nil {
		return err
	}

	contextName := args[0]
	if _, found := loaded.Config.ContextByName(contextName); !found {
		return fmt.Errorf("context %q not found", contextName)
	}

	writePath := loaded.PathForContext(contextName)
	fileConfig := loaded.FileConfig(writePath)

	if deleted := fileConfig.DeleteContext(contextName); !deleted {
		return fmt.Errorf("context %q not found in %q", contextName, writePath)
	}

	if err := config.Write(writePath, &fileConfig); err != nil {
		return err
	}

	if loaded.Config.CurrentContext == contextName {
		if _, warnErr := fmt.Fprintf(
			cmd.ErrOrStderr(),
			"warning: this removed your active context, use %q to select a different one\n",
			"azctx use",
		); warnErr != nil {
			return warnErr
		}
	}

	_, writeErr := fmt.Fprintf(cmd.OutOrStdout(), "deleted context %q from %s\n", contextName, writePath)
	return writeErr
}
