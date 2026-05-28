package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	defaultPath := filepath.Join(homeDir, ".config", configDir, configFile)
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
	path1 := filepath.Clean("/tmp/one.yaml")
	path2 := filepath.Clean("/tmp/two.yaml")
	loader := Loader{fsys: fs, env: path1 + string(os.PathListSeparator) + path2}

	cfg1 := newTestConfig(t).
		withTenant("corp").
		withUserCredential("ci").
		withContext("dev", "corp", "ci").
		withCurrentContext("dev").
		write(fs, path1).
		build()

	newTestConfig(t).
		withTenant("corp").
		withTenant("platform").
		withUserCredential("mi").
		withWIFCredential("wif").
		withContext("dev", "corp", "ci").
		withContext("prod", "platform", "wif").
		write(fs, path2)

	store, err := loader.Load()
	require.NoError(t, err)

	assert.Equal(t, []string{path1, path2}, store.Paths)
	assert.Equal(t, path1, store.WritePath)
	assert.Equal(t, "dev", store.Config.CurrentContext)

	tenant, ok := store.Config.TenantByName("corp")
	require.True(t, ok)
	t1, _ := cfg1.TenantByName("corp")
	assert.Equal(t, t1.Details.ID, tenant.Details.ID)

	_, ok = store.Config.TenantByName("platform")
	assert.True(t, ok)

	contextValue, ok := store.Config.ContextByName("dev")
	require.True(t, ok)
	assert.Empty(t, contextValue.Details.Subscription)

	assert.Equal(t, path1, store.PathForCurrentContext())
	assert.Equal(t, path1, store.PathForTenant("corp"))
	assert.Equal(t, path2, store.PathForTenant("platform"))
	assert.Equal(t, path1, store.PathForContext("dev"))
	assert.Equal(t, path2, store.PathForContext("prod"))
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
	cfg := newTestConfig(t).
		withTenant("corp").
		withUserCredential("personal").
		withContext("dev", "corp", "personal").
		withCurrentContext("dev").
		write(fs, path).
		build()

	got, err := loader.Read(path)
	require.NoError(t, err)
	assert.Equal(t, cfg.CurrentContext, got.CurrentContext)
	assert.Equal(t, cfg.Tenants, got.Tenants)

	missing, err := loader.Read(filepath.Clean("/tmp/missing.yaml"))
	require.NoError(t, err)
	assert.Equal(t, Config{APIVersion: APIVersion, Kind: Kind}, missing)
}

func TestWriterWriteAndLoaderReadRoundTrip(t *testing.T) {
	fs := afero.NewMemMapFs()
	writer := Writer{fsys: fs}
	loader := Loader{fsys: fs}

	path := filepath.Clean("/tmp/nested/config.yaml")
	input := newTestConfig(t).
		withTenant("corp").
		withUserCredential("mi").
		withContext("dev", "corp", "mi").
		withCurrentContext("dev").
		build()

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
	require.ErrorIs(t, err, errNotFound)
	assert.Equal(t, Config{APIVersion: APIVersion, Kind: Kind}, got)
}

func TestLoaderReadConfigReturnsReadError(t *testing.T) {
	loader := Loader{fsys: afero.NewOsFs()}

	_, err := loader.readConfig(filepath.Clean("/dev/null/config.yaml"))
	require.Error(t, err)
	assert.False(t, errors.Is(err, os.ErrNotExist))
}
