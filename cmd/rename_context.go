package cmd

import (
	"fmt"

	"github.com/lvlcn-t/azctx/config"
	"github.com/spf13/cobra"
)

// renameContextCmd renames an existing context entry in config.
var renameContextCmd = &cobra.Command{ //nolint:gochecknoglobals // Cobra command definition
	Use:               "rename-context OLD_NAME NEW_NAME",
	Short:             "Rename a context in azctx config",
	Long:              "Rename a context in azctx config.",
	Example:           "  azctx rename-context staging dev",
	RunE:              runRenameContext,
	DisableAutoGenTag: true,
	Args:              cobra.ExactArgs(2),
}

// runRenameContext executes the rename-context command.
func runRenameContext(cmd *cobra.Command, args []string) error {
	loaded, err := config.Load()
	if err != nil {
		return err
	}

	oldName := args[0]
	newName := args[1]

	if _, found := loaded.Config.ContextByName(oldName); !found {
		return fmt.Errorf("cannot rename context %q, it does not exist", oldName)
	}

	if _, found := loaded.Config.ContextByName(newName); found {
		return fmt.Errorf("cannot rename context %q, context %q already exists", oldName, newName)
	}

	writePath := loaded.PathForContext(oldName)
	fileConfig := loaded.FileConfig(writePath)

	if renamed := fileConfig.RenameContext(oldName, newName); !renamed {
		return fmt.Errorf("cannot rename context %q, it does not exist in %q", oldName, writePath)
	}

	if fileConfig.CurrentContext == oldName {
		fileConfig.CurrentContext = newName
	}

	if err := config.Write(writePath, &fileConfig); err != nil {
		return err
	}

	if loaded.Config.CurrentContext == oldName && writePath != loaded.PathForCurrentContext() {
		currentPath := loaded.PathForCurrentContext()
		currentConfig := loaded.FileConfig(currentPath)
		currentConfig.CurrentContext = newName

		if err := config.Write(currentPath, &currentConfig); err != nil {
			return err
		}
	}

	_, writeErr := fmt.Fprintf(cmd.OutOrStdout(), "Context %q renamed to %q.\n", oldName, newName)
	return writeErr
}
