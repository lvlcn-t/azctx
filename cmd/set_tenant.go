package cmd

import (
	"fmt"

	"github.com/lvlcn-t/azctx/config"
	"github.com/lvlcn-t/azctx/contexts"
	"github.com/spf13/cobra"
)

type setTenantCommand struct {
	manager *contexts.Manager
	loader  config.Loader
}

// newSetTenantCmd creates or updates a tenant entry in config.
func newSetTenantCmd() *cobra.Command {
	command := &setTenantCommand{
		loader:  config.NewLoader(),
		manager: contexts.New(),
	}

	cmd := &cobra.Command{
		Use:               "set-tenant NAME",
		Short:             "Set a tenant entry in azctx config",
		Long:              "Set a tenant entry in azctx config.",
		Example:           "  azctx set-tenant corp --id 00000000-0000-0000-0000-000000000000",
		RunE:              command.run,
		DisableAutoGenTag: true,
		Args:              cobra.ExactArgs(1),
	}

	cmd.Flags().String("id", "", "Tenant ID")
	if err := cmd.MarkFlagRequired("id"); err != nil {
		panic(fmt.Errorf("mark id flag required: %w", err))
	}

	return cmd
}

// run executes the set-tenant command.
func (c *setTenantCommand) run(cmd *cobra.Command, args []string) error {
	store, err := c.loader.Load()
	if err != nil {
		return err
	}

	tenantName := args[0]

	tenantID, err := cmd.Flags().GetString("id")
	if err != nil {
		return fmt.Errorf("read id flag: %w", err)
	}

	wasExisting, err := c.manager.SetTenant(&store, tenantName, tenantID)
	if err != nil {
		return err
	}

	if wasExisting {
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Tenant %q modified.\n", tenantName)
		return err
	}

	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Tenant %q created.\n", tenantName)
	return err
}
