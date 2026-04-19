package cmd

import (
	"fmt"

	"github.com/lvlcn-t/azctx/az"
	"github.com/lvlcn-t/azctx/config"
	"github.com/spf13/cobra"
)

// useCmd switches the active context and syncs Azure CLI state.
var useCmd = &cobra.Command{
	Use:     "use NAME",
	Aliases: []string{"use-context"},
	Short:   "Set the active Azure context",
	Long:    "Set the active Azure context, then sync Azure CLI state by calling az login and az account set.",
	Example: `  azctx use dev
  azctx use prod`,
	RunE: runUse,
	Args: cobra.ExactArgs(1),
}

// runUse executes the use command.
func runUse(cmd *cobra.Command, args []string) error {
	azcli, err := az.NewClient()
	if err != nil {
		return fmt.Errorf("create az client: %w", err)
	}

	loaded, err := config.Load()
	if err != nil {
		return err
	}

	ctxName := args[0]
	ctx, found := loaded.Config.ContextByName(ctxName)
	if !found {
		return fmt.Errorf("context %q not found", ctxName)
	}

	tenant, found := loaded.Config.TenantByName(ctx.Tenant)
	if !found {
		return fmt.Errorf("tenant %q not found for context %q", ctx.Tenant, ctx.Name)
	}

	credential, found := loaded.Config.CredentialByName(ctx.Credential)
	if !found {
		return fmt.Errorf("credential %q not found for context %q", ctx.Credential, ctx.Name)
	}

	if tenant.ID == "" {
		return fmt.Errorf("tenant %q is missing id", tenant.Name)
	}

	if err := azcli.Login(cmd.Context(), &credential, tenant.ID); err != nil {
		return err
	}

	if err := azcli.SetSubscription(cmd.Context(), ctx.Subscription); err != nil {
		return err
	}

	writePath := loaded.PathForCurrentContext()
	fileConfig := loaded.FileConfig(writePath)
	fileConfig.CurrentContext = ctx.Name

	if err := config.Write(writePath, &fileConfig); err != nil {
		return err
	}

	_, writeErr := fmt.Fprintf(cmd.OutOrStdout(), "Switched to context %q.\n", ctx.Name)
	return writeErr
}
