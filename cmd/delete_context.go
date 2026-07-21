package cmd

import (
	"fmt"

	"github.com/lvlcn-t/azctx/config"
	"github.com/lvlcn-t/azctx/contexts"
	"github.com/spf13/cobra"
)

type deleteCtxCmd struct {
	manager *contexts.Manager
	loader  config.Loader
}

// newDeleteCtxCmd removes a context entry from config.
func newDeleteCtxCmd() *cobra.Command {
	command := &deleteCtxCmd{
		loader:  config.NewLoader(),
		manager: contexts.New(),
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

	res, err := c.manager.DeleteContext(&store, name)
	if err != nil {
		return err
	}

	if res.WasActive {
		if _, err = fmt.Fprintf(
			cmd.ErrOrStderr(),
			"warning: this removed your active context, use %q to select a different one\n",
			"azctx use",
		); err != nil {
			return err
		}
	}

	_, err = fmt.Fprintf(cmd.OutOrStdout(), "deleted context %q from %s\n", name, res.Path)
	return err
}
