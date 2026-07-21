// Package login activates an azctx context: it performs az login (acquiring
// and refreshing a federated token for workload-identity credentials) and
// persists the new current-context. It composes the config, az, and wif
// primitives so the CLI (cmd) and the interactive TUI (tui) behave identically.
// It is a leaf package: it must never import cmd or tui.
package login

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

// Manager performs az login and persists the new current-context. Its fields are
// exported so tests can build a Manager with fakes directly.
type Manager struct {
	// NewClient creates the az CLI client used to log in.
	NewClient func(ctx context.Context) (az.CLI, error)
	// Tokens creates the token provider for workload-identity credentials.
	Tokens factory.TokenProvider
	// Writer persists the updated current-context.
	Writer config.Writer
}

// New returns a [Manager] that uses the azure-cli.
func New() *Manager {
	return &Manager{
		NewClient: az.NewClient,
		Tokens:    factory.NewTokenProvider,
		Writer:    config.NewWriter(),
	}
}

// Login resolves the named context, performs az login (acquiring a federated
// token first for workload-identity credentials), and persists the new
// current-context. It retries login once with a force-refreshed token if a
// cached token was rejected.
func (c *Manager) Login(ctx context.Context, store *config.Store, name string) error {
	azcli, err := c.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("create az client: %w", err)
	}

	resolved, err := store.Resolve(name)
	if err != nil {
		return err
	}

	tok, err := c.acquireToken(ctx, store, &resolved)
	if err != nil {
		return err
	}

	loginErr := authenticate(ctx, azcli, &resolved, tok.value)
	// Retry once with a fresh token if login failed and we used a cached token.
	if errors.Is(loginErr, az.ErrLogin) && tok.cached {
		fresh, _, refreshErr := tok.provider.AcquireToken(ctx, wif.WithForceRefresh())
		if refreshErr != nil {
			return fmt.Errorf("az login: %w (refresh also failed: %w)", loginErr, refreshErr)
		}

		loginErr = authenticate(ctx, azcli, &resolved, fresh)
	}

	if loginErr != nil {
		return fmt.Errorf("az login: %w", loginErr)
	}

	path := store.PathForCurrentContext()
	cfg := store.FileConfig(path)
	cfg.CurrentContext = resolved.Name

	return c.Writer.Write(path, &cfg)
}

// federatedToken holds the workload-identity token state produced for login.
// provider is retained so a rejected cached token can be force-refreshed.
type federatedToken struct {
	provider wif.Provider
	value    string
	cached   bool
}

// acquireToken fetches a federated token for workload-identity credentials. For
// other credential types it returns the zero federatedToken.
func (c *Manager) acquireToken(ctx context.Context, store *config.Store, resolved *config.ResolvedContext) (federatedToken, error) {
	if resolved.Credential.Details.Type != config.CredentialTypeWorkloadIdentity {
		return federatedToken{}, nil
	}

	cacheDir := filepath.Dir(store.PathForCurrentContext())
	provider, err := c.Tokens(ctx, resolved.Credential.Details.Token, cacheDir)
	if err != nil {
		return federatedToken{}, fmt.Errorf("create token provider: %w", err)
	}

	value, cached, err := provider.AcquireToken(ctx)
	if err != nil {
		return federatedToken{}, fmt.Errorf("acquire federated token: %w", err)
	}

	return federatedToken{value: value, cached: cached, provider: provider}, nil
}

// authenticate applies the resolved context to the client and logs in.
func authenticate(ctx context.Context, azcli az.CLI, resolved *config.ResolvedContext, token string) error {
	return azcli.WithTenant(resolved.Tenant.Details.ID).
		WithCredential(&resolved.Credential).
		WithSubscription(resolved.Subscription).
		AllowNoSubscriptions(resolved.AllowNoSubscriptions).
		WithFederatedToken(token).
		Login(ctx)
}
