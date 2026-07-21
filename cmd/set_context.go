package cmd

import (
	"fmt"

	"github.com/lvlcn-t/azctx/config"
	"github.com/lvlcn-t/azctx/contexts"
	"github.com/spf13/cobra"
)

type setCtxCmd struct {
	manager *contexts.Manager
	loader  config.Loader
}

// newSetCtxCmd creates or updates a context entry in config.
func newSetCtxCmd() *cobra.Command {
	command := &setCtxCmd{
		loader:  config.NewLoader(),
		manager: contexts.New(),
	}

	cmd := &cobra.Command{
		Use:               "set-context NAME",
		Short:             "Set a context entry in azctx config",
		Long:              "Set a context entry in azctx config. The context points to tenant and credential entries in the same merged azctx config.",
		Example:           "  azctx set-context prod --tenant corp --credential ci-sp --subscription 00000000-0000-0000-0000-000000000000",
		RunE:              command.run,
		DisableAutoGenTag: true,
		Args:              cobra.ExactArgs(1),
	}

	cmd.Flags().String("tenant", "", "Tenant name for the context")
	cmd.Flags().String("credential", "", "Credential name for the context")
	cmd.Flags().String("subscription", "", "Optional subscription ID for the context")

	return cmd
}

// run executes the set-context command.
func (c *setCtxCmd) run(cmd *cobra.Command, args []string) error {
	store, err := c.loader.Load()
	if err != nil {
		return err
	}

	ctx := args[0]

	tenant, err := cmd.Flags().GetString("tenant")
	if err != nil {
		return fmt.Errorf("read tenant flag: %w", err)
	}

	credential, err := cmd.Flags().GetString("credential")
	if err != nil {
		return fmt.Errorf("read credential flag: %w", err)
	}

	subscriptionID, err := cmd.Flags().GetString("subscription")
	if err != nil {
		return fmt.Errorf("read subscription flag: %w", err)
	}

	next := config.Context{
		Name: ctx,
		Details: config.ContextDetails{
			Tenant:       tenant,
			Credential:   credential,
			Subscription: subscriptionID,
		},
	}

	hasExistingInMerged, err := c.manager.SetContext(
		&store,
		next,
		cmd.Flags().Changed("subscription"),
	)
	if err != nil {
		return err
	}

	if hasExistingInMerged {
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Context %q modified.\n", ctx)
		return err
	}

	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Context %q created.\n", ctx)
	return err
}
