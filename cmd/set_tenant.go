package cmd

import (
	"fmt"

	"github.com/lvlcn-t/azctx/config"
	"github.com/spf13/cobra"
)

// setTenantCmd creates or updates a tenant entry in config.
var setTenantCmd = &cobra.Command{ //nolint:gochecknoglobals // Cobra command definition
	Use:               "set-tenant NAME",
	Short:             "Set a tenant entry in azctx config",
	Long:              "Set a tenant entry in azctx config.",
	Example:           "  azctx set-tenant corp --id 00000000-0000-0000-0000-000000000000",
	RunE:              runSetTenant,
	DisableAutoGenTag: true,
	Args:              cobra.ExactArgs(1),
}

func init() { //nolint:gochecknoinits // Cobra command setup
	setTenantCmd.Flags().String("id", "", "Tenant ID")

	if err := setTenantCmd.MarkFlagRequired("id"); err != nil {
		panic(fmt.Errorf("mark id flag required: %w", err))
	}
}

// runSetTenant executes the set-tenant command.
func runSetTenant(cmd *cobra.Command, args []string) error {
	loaded, err := config.Load()
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
	if _, found := loaded.Config.TenantByName(tenantName); found {
		wasExisting = true
	}

	writePath := loaded.PathForTenant(tenantName)
	fileConfig := loaded.FileConfig(writePath)
	nextTenant := config.Tenant{Name: tenantName, ID: tenantID}
	fileConfig.UpsertTenant(nextTenant)

	if err := config.Write(writePath, &fileConfig); err != nil {
		return err
	}

	if wasExisting {
		_, writeErr := fmt.Fprintf(cmd.OutOrStdout(), "Tenant %q modified.\n", tenantName)
		return writeErr
	}

	_, writeErr := fmt.Fprintf(cmd.OutOrStdout(), "Tenant %q created.\n", tenantName)
	return writeErr
}
