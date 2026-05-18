package cmd

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/lvlcn-t/azctx/az"
	"github.com/lvlcn-t/azctx/config"
	"github.com/lvlcn-t/azctx/wif"
	"github.com/lvlcn-t/azctx/wif/factory"
	"github.com/spf13/cobra"
)

type useCommand struct {
	az     func(ctx context.Context) (az.CLI, error)
	wif    func(ctx context.Context, cfg config.TokenDetails, cacheDir string) (wif.Provider, error)
	loader config.Loader
	writer config.Writer
}

// newUseCmd switches the active context and syncs Azure CLI state.
func newUseCmd() *cobra.Command {
	command := &useCommand{
		loader: config.NewLoader(),
		writer: config.NewWriter(),
		az:     az.NewClient,
		wif:    factory.NewProvider,
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

	resolved, err := store.Resolve(args[0])
	if err != nil {
		return err
	}

	var token string
	var cached bool
	var provider wif.Provider
	if resolved.Credential.Details.Type == config.CredentialTypeWorkloadIdentity {
		cacheDir := filepath.Dir(store.PathForCurrentContext())
		provider, err = c.wif(cmd.Context(), resolved.Credential.Details.Token, cacheDir)
		if err != nil {
			return fmt.Errorf("create token provider: %w", err)
		}

		token, cached, err = provider.AcquireToken(cmd.Context())
		if err != nil {
			return fmt.Errorf("acquire federated token: %w", err)
		}
	}

	loginErr := azcli.WithTenant(resolved.Tenant.Details.ID).
		WithCredential(&resolved.Credential).
		WithSubscription(resolved.Subscription).
		AllowNoSubscriptions(resolved.AllowNoSubscriptions).
		WithFederatedToken(token).
		Login(cmd.Context())

	// Retry once with a fresh token if login failed due to az login and we used a cached token.
	if errors.Is(loginErr, az.ErrLogin) && cached {
		token, _, err = provider.AcquireToken(cmd.Context(), wif.WithForceRefresh())
		if err != nil {
			return fmt.Errorf("az login: %w (refresh also failed: %w)", loginErr, err)
		}

		loginErr = azcli.WithTenant(resolved.Tenant.Details.ID).
			WithCredential(&resolved.Credential).
			WithSubscription(resolved.Subscription).
			AllowNoSubscriptions(resolved.AllowNoSubscriptions).
			WithFederatedToken(token).
			Login(cmd.Context())
	}

	if loginErr != nil {
		return fmt.Errorf("az login: %w", loginErr)
	}

	path := store.PathForCurrentContext()
	cfg := store.FileConfig(path)
	cfg.CurrentContext = resolved.Name

	if err = c.writer.Write(path, &cfg); err != nil {
		return err
	}

	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Switched to context %q.\n", resolved.Name)
	return err
}
