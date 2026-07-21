package contexts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

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

// writeConfigs writes multiple config files to a temp dir and points azctx at
// all of them (colon-joined). It returns the paths in the given order.
func writeConfigs(t *testing.T, cfgs ...*config.Config) []string {
	t.Helper()

	dir := t.TempDir()
	writer := config.NewWriter()
	paths := make([]string, len(cfgs))
	for i, cfg := range cfgs {
		path := filepath.Join(dir, fmt.Sprintf("azctx-%d.yaml", i))
		require.NoError(t, writer.Write(path, cfg))
		paths[i] = path
	}

	t.Setenv(config.ConfigEnvVar, strings.Join(paths, string(os.PathListSeparator)))

	return paths
}

// readConfig reads a single config file from disk.
func readConfig(t *testing.T, path string) *config.Config {
	t.Helper()

	loader := config.NewLoader()
	cfg, err := loader.Read(path)
	require.NoError(t, err)

	return &cfg
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
