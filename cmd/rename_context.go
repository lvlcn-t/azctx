package cmd

import (
	"fmt"

	"github.com/lvlcn-t/azctx/config"
	"github.com/spf13/cobra"
)

type renameCtxCmd struct {
	writer config.Writer
	loader config.Loader
}

// newRenameCtxCmd renames an existing context entry in config.
func newRenameCtxCmd() *cobra.Command {
	command := &renameCtxCmd{
		loader: config.NewLoader(),
		writer: config.NewWriter(),
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

	if _, found := store.Config.ContextByName(oldName); !found {
		return fmt.Errorf("cannot rename context %q, it does not exist", oldName)
	}

	if _, found := store.Config.ContextByName(newName); found {
		return fmt.Errorf("cannot rename context %q, context %q already exists", oldName, newName)
	}

	path := store.PathForContext(oldName)
	cfg := store.FileConfig(path)

	if renamed := cfg.RenameContext(oldName, newName); !renamed {
		return fmt.Errorf("cannot rename context %q, it does not exist in %q", oldName, path)
	}

	if cfg.CurrentContext == oldName {
		cfg.CurrentContext = newName
	}

	if err := c.writer.Write(path, &cfg); err != nil {
		return err
	}

	if store.Config.CurrentContext == oldName && path != store.PathForCurrentContext() {
		currentPath := store.PathForCurrentContext()
		currentConfig := store.FileConfig(currentPath)
		currentConfig.CurrentContext = newName

		if err := c.writer.Write(currentPath, &currentConfig); err != nil {
			return err
		}
	}

	_, writeErr := fmt.Fprintf(cmd.OutOrStdout(), "Context %q renamed to %q.\n", oldName, newName)
	return writeErr
}
