package cmd

import (
	"fmt"

	"github.com/lvlcn-t/azctx/config"
	"github.com/lvlcn-t/azctx/tui"
	"github.com/lvlcn-t/azctx/tui/state"
	"github.com/spf13/cobra"
)

type useCommand struct {
	switcher contextSwitcher
	loader   config.Loader
}

// newUseCmd switches the active context and syncs Azure CLI state.
func newUseCmd() *cobra.Command {
	command := &useCommand{
		switcher: newContextSwitcher(),
		loader:   config.NewLoader(),
	}

	useCmd := &cobra.Command{ //nolint:exhaustruct // Cobra command definition
		Use:     "use NAME",
		Aliases: []string{"use-context"},
		Short:   "Set the active Azure context",
		Long:    "Set the active Azure context, then sync Azure CLI state by calling az login and az account set.",
		Example: `  azctx use dev
  azctx use prod`,
		RunE: command.run,
		Args: cobra.MaximumNArgs(1),
	}

	return useCmd
}

// run executes the use command.
func (c *useCommand) run(cmd *cobra.Command, args []string) error {
	var name string

	store, err := c.loader.Load()
	if err != nil {
		return err
	}

	if len(args) > 0 {
		name = args[0]
	} else {
		picked, perr := tui.RunV2(&store, state.ModeInteractive)
		if perr != nil {
			return perr
		}
		if picked == "" {
			return nil
		}
		name = picked
	}

	if err = c.switcher.switchContext(cmd.Context(), &store, name); err != nil {
		return err
	}

	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Switched to context %q.\n", name)
	return err
}
