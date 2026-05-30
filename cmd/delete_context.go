package cmd

import (
	"fmt"

	"github.com/lvlcn-t/azctx/config"
	"github.com/spf13/cobra"
)

type deleteCtxCmd struct {
	writer config.Writer
	loader config.Loader
}

// newDeleteCtxCmd removes a context entry from config.
func newDeleteCtxCmd() *cobra.Command {
	command := &deleteCtxCmd{
		loader: config.NewLoader(),
		writer: config.NewWriter(),
	}

	cmd := &cobra.Command{
		Use:               "delete-context NAME",
		Aliases:           []string{"unset-context"},
		Short:             "Delete a context from azctx config",
		Long:              "Delete a context from azctx config.",
		Example:           "  azctx delete-context prod",
		RunE:              command.run,
		DisableAutoGenTag: true,
		Args:              cobra.ExactArgs(1),
	}

	return cmd
}

// run executes the delete-context command.
func (c *deleteCtxCmd) run(cmd *cobra.Command, args []string) error {
	store, err := c.loader.Load()
	if err != nil {
		return err
	}

	name := args[0]
	if _, found := store.Config.ContextByName(name); !found {
		return fmt.Errorf("context %q not found", name)
	}

	path := store.PathForContext(name)
	cfg := store.FileConfig(path)
	if deleted := cfg.DeleteContext(name); !deleted {
		return fmt.Errorf("context %q not found in %q", name, path)
	}

	if err := c.writer.Write(path, &cfg); err != nil {
		return err
	}

	if store.Config.CurrentContext == name {
		if _, warnErr := fmt.Fprintf(
			cmd.ErrOrStderr(),
			"warning: this removed your active context, use %q to select a different one\n",
			"azctx use",
		); warnErr != nil {
			return warnErr
		}
	}

	_, writeErr := fmt.Fprintf(cmd.OutOrStdout(), "deleted context %q from %s\n", name, path)
	return writeErr
}
