package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v4"
)

var (
	tenantCorp = Tenant{Name: "corp", Tenant: TenantDetails{ID: "tenant-1"}}
	tenantPlat = Tenant{Name: "platform", Tenant: TenantDetails{ID: "tenant-2"}}

	credUser = Credential{Name: "cred-a", Credential: CredentialDetails{Type: CredentialTypeUser}}
	credMI   = Credential{Name: "cred-b", Credential: CredentialDetails{Type: CredentialTypeManagedIdentity}}
	credOIDC = Credential{Name: "cred-c", Credential: CredentialDetails{Type: CredentialTypeWorkloadIdentity}}

	devContext  = Context{Name: "dev", Context: ContextDetails{Tenant: tenantCorp.Name, Credential: credUser.Name}}
	prodContext = Context{Name: "prod", Context: ContextDetails{Tenant: tenantPlat.Name, Credential: credMI.Name, Subscription: "sub-prod"}}
)

func TestExpandPath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr string
	}{
		{name: "tilde", input: "~", want: homeDir},
		{name: "tilde child", input: "~/config.yaml", want: filepath.Join(homeDir, "config.yaml")},
		{name: "relative path", input: "./config.yaml", want: filepath.Clean("./config.yaml")},
		{name: "unsupported user expansion", input: "~other/config.yaml", wantErr: "only current-user"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := expandPath(tt.input)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLoaderResolvePaths(t *testing.T) {
	homeDir, err := os.UserConfigDir()
	require.NoError(t, err)

	defaultPath := filepath.Join(homeDir, configDir, configFile)
	pathA := filepath.Clean("/tmp/a.yaml")
	pathB := filepath.Clean("/tmp/b.yaml")

	tests := []struct {
		name string
		env  string
		want []string
	}{
		{name: "empty falls back to default", env: "", want: []string{defaultPath}},
		{name: "single path", env: pathA, want: []string{pathA}},
		{
			name: "deduplicates",
			env:  pathA + string(os.PathListSeparator) + pathA + string(os.PathListSeparator) + pathB,
			want: []string{pathA, pathB},
		},
		{
			name: "whitespace and empty parts",
			env:  "  " + pathA + "  " + string(os.PathListSeparator) + "  " + string(os.PathListSeparator) + pathB,
			want: []string{pathA, pathB},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := Loader{fsys: afero.NewMemMapFs(), env: tt.env}

			got, resolveErr := loader.resolvePaths()
			require.NoError(t, resolveErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLoaderLoadMergesWithFirstWins(t *testing.T) {
	fs := afero.NewMemMapFs()
	pathOne := filepath.Clean("/tmp/one.yaml")
	pathTwo := filepath.Clean("/tmp/two.yaml")
	loader := Loader{fsys: fs, env: pathOne + string(os.PathListSeparator) + pathTwo}

	cfgOne := Config{
		CurrentContext: devContext.Name,
		Tenants:        []Tenant{tenantCorp},
		Credentials:    []Credential{credUser},
		Contexts:       []Context{devContext},
	}
	writeConfigYAML(t, fs, pathOne, &cfgOne)

	cfgTwoTenantCorp := Tenant{Name: "corp", Tenant: TenantDetails{ID: "tenant-2"}}
	cfgTwoTenantPlat := Tenant{Name: "platform", Tenant: TenantDetails{ID: "tenant-3"}}

	cfgTwo := Config{
		CurrentContext: prodContext.Name,
		Tenants:        []Tenant{cfgTwoTenantCorp, cfgTwoTenantPlat},
		Credentials:    []Credential{credMI, credOIDC},
		Contexts:       []Context{devContext, prodContext},
	}
	writeConfigYAML(t, fs, pathTwo, &cfgTwo)

	store, err := loader.Load()
	require.NoError(t, err)

	assert.Equal(t, []string{pathOne, pathTwo}, store.Paths)
	assert.Equal(t, pathOne, store.WritePath)
	assert.Equal(t, devContext.Name, store.Config.CurrentContext)

	tenant, ok := store.Config.TenantByName(tenantCorp.Name)
	require.True(t, ok)
	assert.Equal(t, "tenant-1", tenant.Tenant.ID)

	_, ok = store.Config.TenantByName(tenantPlat.Name)
	assert.True(t, ok)

	contextValue, ok := store.Config.ContextByName(devContext.Name)
	require.True(t, ok)
	assert.Empty(t, contextValue.Context.Subscription)

	assert.Equal(t, pathOne, store.PathForCurrentContext())
	assert.Equal(t, pathOne, store.PathForTenant(tenantCorp.Name))
	assert.Equal(t, pathTwo, store.PathForTenant(cfgTwoTenantPlat.Name))
	assert.Equal(t, pathOne, store.PathForContext(devContext.Name))
	assert.Equal(t, pathTwo, store.PathForContext(prodContext.Name))
}

func TestLoaderLoadMissingFileUsesFirstPathAsWritePath(t *testing.T) {
	fs := afero.NewMemMapFs()
	pathOne := filepath.Clean("/tmp/one.yaml")
	pathTwo := filepath.Clean("/tmp/two.yaml")
	loader := Loader{fsys: fs, env: pathOne + string(os.PathListSeparator) + pathTwo}

	store, err := loader.Load()
	require.NoError(t, err)

	assert.Equal(t, pathOne, store.WritePath)
	assert.Equal(t, pathOne, store.PathForContext("new"))
}

func TestLoaderLoadReturnsErrorOnBadYAML(t *testing.T) {
	fs := afero.NewMemMapFs()
	path := filepath.Clean("/tmp/one.yaml")
	loader := Loader{fsys: fs, env: path}

	err := afero.WriteFile(fs, path, []byte("::invalid"), 0o600)
	require.NoError(t, err)

	_, loadErr := loader.Load()
	require.Error(t, loadErr)
	assert.ErrorContains(t, loadErr, "parse config")
}

func TestLoaderRead(t *testing.T) {
	fs := afero.NewMemMapFs()
	loader := Loader{fsys: fs}

	path := filepath.Clean("/tmp/config.yaml")
	cfg := Config{
		CurrentContext: devContext.Name,
		Tenants:        []Tenant{tenantCorp},
	}
	writeConfigYAML(t, fs, path, &cfg)

	got, err := loader.Read(path)
	require.NoError(t, err)
	assert.Equal(t, cfg.CurrentContext, got.CurrentContext)
	assert.Equal(t, cfg.Tenants, got.Tenants)

	missing, err := loader.Read(filepath.Clean("/tmp/missing.yaml"))
	require.NoError(t, err)
	assert.Equal(t, Config{}, missing)
}

func TestWriterWriteAndLoaderReadRoundTrip(t *testing.T) {
	fs := afero.NewMemMapFs()
	writer := Writer{fsys: fs}
	loader := Loader{fsys: fs}

	path := filepath.Clean("/tmp/nested/config.yaml")
	input := Config{
		CurrentContext: devContext.Name,
		Tenants:        []Tenant{tenantCorp},
		Credentials:    []Credential{credMI},
		Contexts:       []Context{{Name: devContext.Name, Context: ContextDetails{Tenant: tenantCorp.Name, Credential: credMI.Name, Subscription: "sub-1"}}},
	}

	err := writer.Write(path, &input)
	require.NoError(t, err)

	info, err := fs.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())

	parentInfo, err := fs.Stat(filepath.Dir(path))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o700), parentInfo.Mode().Perm())

	got, err := loader.Read(path)
	require.NoError(t, err)
	assert.Equal(t, input, got)
}

func TestLoaderReadConfigFileNotExist(t *testing.T) {
	loader := Loader{fsys: afero.NewMemMapFs()}

	got, err := loader.readConfig(filepath.Clean("/tmp/missing.yaml"))
	require.NoError(t, err)
	assert.Equal(t, Config{}, got)
}

func TestLoaderReadConfigReturnsReadError(t *testing.T) {
	loader := Loader{fsys: afero.NewOsFs()}

	_, err := loader.readConfig(filepath.Clean("/dev/null/config.yaml"))
	require.Error(t, err)
	assert.False(t, errors.Is(err, os.ErrNotExist))
}

func writeConfigYAML(t *testing.T, fs afero.Fs, path string, cfg *Config) {
	t.Helper()

	const (
		dirMode  = 0o700
		fileMode = 0o600
	)
	encoded, err := yaml.Marshal(cfg)
	require.NoError(t, err)
	require.NoError(t, fs.MkdirAll(filepath.Dir(path), dirMode))
	require.NoError(t, afero.WriteFile(fs, path, encoded, fileMode))
}
