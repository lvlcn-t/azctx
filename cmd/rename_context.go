package cmd

import (
	"fmt"

	"github.com/lvlcn-t/azctx/config"
	"github.com/lvlcn-t/azctx/contexts"
	"github.com/spf13/cobra"
)

type renameCtxCmd struct {
	manager *contexts.Manager
	loader  config.Loader
}

// newRenameCtxCmd renames an existing context entry in config.
func newRenameCtxCmd() *cobra.Command {
	command := &renameCtxCmd{
		loader:  config.NewLoader(),
		manager: contexts.New(),
	}

	renameContextCmd := &cobra.Command{ //nolint:exhaustruct // Cobra command definition
		Use:               "rename-context OLD_NAME NEW_NAME",
		Short:             "Rename a context in azctx config",
		Long:              "Rename a context in azctx config.",
		Example:           "  azctx rename-context staging dev",
		RunE:              command.run,
		DisableAutoGenTag: true,
		Args:              cobra.ExactArgs(2),
	}

	return renameContextCmd
}

// run executes the rename-context command.
func (c *renameCtxCmd) run(cmd *cobra.Command, args []string) error {
	store, err := c.loader.Load()
	if err != nil {
		return err
	}

	oldName := args[0]
	newName := args[1]

	if err := c.manager.RenameContext(&store, oldName, newName); err != nil {
		return err
	}

	_, writeErr := fmt.Fprintf(cmd.OutOrStdout(), "Context %q renamed to %q.\n", oldName, newName)
	return writeErr
}
