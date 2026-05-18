package cmd

import (
	"context"
	"fmt"

	"github.com/lvlcn-t/azctx/az"
	"github.com/lvlcn-t/azctx/config"
	"github.com/lvlcn-t/azctx/wif"
	"github.com/spf13/cobra"
)

type useCommand struct {
	az     func(ctx context.Context) (az.CLI, error)
	wif    func(ctx context.Context, cfg config.TokenDetails) (wif.Provider, error)
	loader config.Loader
	writer config.Writer
}

// newUseCmd switches the active context and syncs Azure CLI state.
func newUseCmd() *cobra.Command {
	command := &useCommand{
		loader: config.NewLoader(),
		writer: config.NewWriter(),
		az:     az.NewClient,
		wif:    wif.NewProvider,
	}

	useCmd := &cobra.Command{ //nolint:exhaustruct // Cobra command definition
		Use:     "use NAME",
		Aliases: []string{"use-context"},
		Short:   "Set the active Azure context",
		Long:    "Set the active Azure context, then sync Azure CLI state by calling az login and az account set.",
		Example: `  azctx use dev
  azctx use prod`,
		RunE: command.run,
		Args: cobra.ExactArgs(1),
	}

	return useCmd
}

// run executes the use command.
func (c *useCommand) run(cmd *cobra.Command, args []string) error {
	azcli, err := c.az(cmd.Context())
	if err != nil {
		return fmt.Errorf("create az client: %w", err)
	}

	store, err := c.loader.Load()
	if err != nil {
		return err
	}

	ctxName := args[0]
	ctx, found := store.Config.ContextByName(ctxName)
	if !found {
		return fmt.Errorf("context %q not found", ctxName)
	}

	tenant, found := store.Config.TenantByName(ctx.Context.Tenant)
	if !found {
		return fmt.Errorf("tenant %q not found for context %q", ctx.Context.Tenant, ctx.Name)
	}

	credential, found := store.Config.CredentialByName(ctx.Context.Credential)
	if !found {
		return fmt.Errorf("credential %q not found for context %q", ctx.Context.Credential, ctx.Name)
	}

	if tenant.Tenant.ID == "" {
		return fmt.Errorf("tenant %q is missing id", tenant.Name)
	}

	var token string
	if credential.Credential.Type == config.CredentialTypeWorkloadIdentity {
		var provider wif.Provider
		provider, err = c.wif(cmd.Context(), credential.Credential.Token)
		if err != nil {
			return fmt.Errorf("create token provider: %w", err)
		}

		token, err = provider.AcquireToken(cmd.Context())
		if err != nil {
			return fmt.Errorf("acquire federated token: %w", err)
		}
	}

	err = azcli.WithTenant(tenant.Tenant.ID).
		WithCredential(&credential).
		WithSubscription(ctx.Context.Subscription).
		AllowNoSubscriptions(ctx.Context.AllowNoSubscriptions).
		WithFederatedToken(token).
		Login(cmd.Context())
	if err != nil {
		return fmt.Errorf("az login: %w", err)
	}

	path := store.PathForCurrentContext()
	cfg := store.FileConfig(path)
	cfg.CurrentContext = ctx.Name

	if err = c.writer.Write(path, &cfg); err != nil {
		return err
	}

	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Switched to context %q.\n", ctx.Name)
	return err
}
