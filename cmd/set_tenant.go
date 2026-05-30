package cmd

import (
	"fmt"

	"github.com/lvlcn-t/azctx/config"
	"github.com/spf13/cobra"
)

type setTenantCommand struct {
	writer config.Writer
	loader config.Loader
}

// newSetTenantCmd creates or updates a tenant entry in config.
func newSetTenantCmd() *cobra.Command {
	command := &setTenantCommand{
		loader: config.NewLoader(),
		writer: config.NewWriter(),
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
	if tenantName == "" {
		return fmt.Errorf("tenant name must not be empty")
	}

	tenantID, err := cmd.Flags().GetString("id")
	if err != nil {
		return fmt.Errorf("read id flag: %w", err)
	}

	if tenantID == "" {
		return fmt.Errorf("tenant id must not be empty")
	}

	wasExisting := false
	if _, found := store.Config.TenantByName(tenantName); found {
		wasExisting = true
	}

	path := store.PathForTenant(tenantName)
	cfg := store.FileConfig(path)
	nextTenant := config.Tenant{Name: tenantName, Details: config.TenantDetails{ID: tenantID}}
	cfg.UpsertTenant(nextTenant)

	if err = c.writer.Write(path, &cfg); err != nil {
		return err
	}

	if wasExisting {
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Tenant %q modified.\n", tenantName)
		return err
	}

	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Tenant %q created.\n", tenantName)
	return err
}
