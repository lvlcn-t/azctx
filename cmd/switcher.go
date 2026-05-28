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
)

// contextSwitcher performs az login and writes the new current-context.
type contextSwitcher interface {
	switchContext(ctx context.Context, store *config.Store, name string) error
}

// azContextSwitcher is the production implementation of contextSwitcher.
type azContextSwitcher struct {
	az     func(ctx context.Context) (az.CLI, error)
	wif    factory.TokenProvider
	writer config.Writer
}

// newContextSwitcher creates a contextSwitcher with default production deps.
func newContextSwitcher() *azContextSwitcher {
	return &azContextSwitcher{
		az:     az.NewClient,
		wif:    factory.NewTokenProvider,
		writer: config.NewWriter(),
	}
}

func (s *azContextSwitcher) switchContext(ctx context.Context, store *config.Store, name string) error {
	azcli, err := s.az(ctx)
	if err != nil {
		return fmt.Errorf("create az client: %w", err)
	}

	resolved, err := store.Resolve(name)
	if err != nil {
		return err
	}

	var token string
	var cached bool
	var provider wif.Provider
	if resolved.Credential.Details.Type == config.CredentialTypeWorkloadIdentity {
		cacheDir := filepath.Dir(store.PathForCurrentContext())
		provider, err = s.wif(ctx, resolved.Credential.Details.Token, cacheDir)
		if err != nil {
			return fmt.Errorf("create token provider: %w", err)
		}

		token, cached, err = provider.AcquireToken(ctx)
		if err != nil {
			return fmt.Errorf("acquire federated token: %w", err)
		}
	}

	loginErr := azcli.WithTenant(resolved.Tenant.Details.ID).
		WithCredential(&resolved.Credential).
		WithSubscription(resolved.Subscription).
		AllowNoSubscriptions(resolved.AllowNoSubscriptions).
		WithFederatedToken(token).
		Login(ctx)

	// Retry once with a fresh token if login failed and we used a cached token.
	if errors.Is(loginErr, az.ErrLogin) && cached {
		token, _, err = provider.AcquireToken(ctx, wif.WithForceRefresh())
		if err != nil {
			return fmt.Errorf("az login: %w (refresh also failed: %w)", loginErr, err)
		}

		loginErr = azcli.WithTenant(resolved.Tenant.Details.ID).
			WithCredential(&resolved.Credential).
			WithSubscription(resolved.Subscription).
			AllowNoSubscriptions(resolved.AllowNoSubscriptions).
			WithFederatedToken(token).
			Login(ctx)
	}

	if loginErr != nil {
		return fmt.Errorf("az login: %w", loginErr)
	}

	path := store.PathForCurrentContext()
	cfg := store.FileConfig(path)
	cfg.CurrentContext = resolved.Name

	return s.writer.Write(path, &cfg)
}
