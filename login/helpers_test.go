package login

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/lvlcn-t/azctx/az"
	"github.com/lvlcn-t/azctx/config"
	"github.com/stretchr/testify/require"
)

const (
	devContext  = "dev"
	prodContext = "prod"
)

// writeConfig writes cfg to a temp file and points azctx at it via the config
// env var. It returns the file path.
func writeConfig(t *testing.T, cfg *config.Config) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "azctx.yaml")
	writer := config.NewWriter()
	require.NoError(t, writer.Write(path, cfg))
	t.Setenv(config.ConfigEnvVar, path)

	return path
}

// loadStore loads the store from the config pointed to by writeConfig.
func loadStore(t *testing.T) *config.Store {
	t.Helper()

	loader := config.NewLoader()
	store, err := loader.Load()
	require.NoError(t, err)

	return &store
}

// readConfig reads a single config file from disk.
func readConfig(t *testing.T, path string) *config.Config {
	t.Helper()

	loader := config.NewLoader()
	cfg, err := loader.Read(path)
	require.NoError(t, err)

	return &cfg
}

// staticClient returns an az.CLIMock whose builder methods all return the mock
// itself and whose Login runs loginFn.
func staticClient(loginFn func() error) *az.CLIMock {
	var mock *az.CLIMock
	mock = &az.CLIMock{
		WithTenantFunc:           func(string) az.CLI { return mock },
		WithCredentialFunc:       func(*config.Credential) az.CLI { return mock },
		WithSubscriptionFunc:     func(string) az.CLI { return mock },
		AllowNoSubscriptionsFunc: func(bool) az.CLI { return mock },
		WithFederatedTokenFunc:   func(string) az.CLI { return mock },
		LoginFunc:                func(_ context.Context) error { return loginFn() },
	}

	return mock
}

// baseConfig returns a config with two tenants, two credentials, and two
// contexts, with dev as the current context.
func baseConfig() *config.Config {
	return &config.Config{
		CurrentContext: devContext,
		Tenants: []config.Tenant{
			{Name: "corp", Details: config.TenantDetails{ID: "tenant-1"}},
			{Name: "platform", Details: config.TenantDetails{ID: "tenant-2"}},
		},
		Credentials: []config.Credential{
			{Name: "user", Details: config.CredentialDetails{Type: config.CredentialTypeUser}},
			{
				Name: "sp",
				Details: config.CredentialDetails{
					Type: config.CredentialTypeServicePrincipal,
					Azure: config.AzureCredential{
						ClientID:     "client-1",
						ClientSecret: "secret-1",
					},
				},
			},
		},
		Contexts: []config.Context{
			{
				Name: devContext,
				Details: config.ContextDetails{
					Tenant:       "corp",
					Credential:   "user",
					Subscription: "sub-dev",
				},
			},
			{
				Name: prodContext,
				Details: config.ContextDetails{
					Tenant:       "platform",
					Credential:   "sp",
					Subscription: "sub-prod",
				},
			},
		},
	}
}
