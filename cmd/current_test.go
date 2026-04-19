package cmd

import (
	"testing"

	"github.com/lvlcn-t/azctx/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCurrentText(t *testing.T) {
	writeConfigForTest(t, baseConfig())

	stdout, _, err := executeCommand(t, newCurrentCmd())
	require.NoError(t, err)

	assert.Equal(t, "dev\n", stdout)
}

func TestCurrentNoCurrentContext(t *testing.T) {
	cfg := baseConfig()
	cfg.CurrentContext = ""
	writeConfigForTest(t, cfg)

	_, _, err := executeCommand(t, newCurrentCmd())
	require.Error(t, err)
	assert.ErrorIs(t, err, errCurrentContextUnset)
}

func TestCurrentVerboseText(t *testing.T) {
	writeConfigForTest(t, baseConfig())

	stdout, _, err := executeCommand(t, newCurrentCmd(), "--verbose")
	require.NoError(t, err)

	assert.Contains(t, stdout, "name: dev")
	assert.Contains(t, stdout, "tenant: corp")
	assert.Contains(t, stdout, "subscription: sub-dev")
}

func TestCurrentVerboseTable(t *testing.T) {
	writeConfigForTest(t, baseConfig())

	stdout, _, err := executeCommand(t, newCurrentCmd(), "--verbose", "-o", "table")
	require.NoError(t, err)

	assert.Contains(t, stdout, "CURRENT")
	assert.Contains(t, stdout, "TENANT")
	assert.Contains(t, stdout, "dev")
	assert.Contains(t, stdout, "corp")
}

func TestCurrentJSON(t *testing.T) {
	writeConfigForTest(t, baseConfig())

	stdout, _, err := executeCommand(t, newCurrentCmd(), "-o", "json")
	require.NoError(t, err)

	assert.Contains(t, stdout, `"name": "dev"`)
	assert.NotContains(t, stdout, `"current"`)
}

func TestCurrentTable(t *testing.T) {
	writeConfigForTest(t, baseConfig())

	stdout, _, err := executeCommand(t, newCurrentCmd(), "-o", "table")
	require.NoError(t, err)

	assert.Contains(t, stdout, "CURRENT")
	assert.Contains(t, stdout, "NAME")
	assert.Contains(t, stdout, "dev")
	assert.NotContains(t, stdout, "TENANT ID")
}

func TestCurrentContextMissingFromList(t *testing.T) {
	cfg := baseConfig()
	cfg.CurrentContext = "ghost"
	writeConfigForTest(t, cfg)

	_, _, err := executeCommand(t, newCurrentCmd())
	require.Error(t, err)
	assert.ErrorContains(t, err, `context "ghost" not found`)
}

func TestCurrentLoadError(t *testing.T) {
	t.Setenv(config.ConfigEnvVar, "~other/config.yaml")
	command := &currentCmd{loader: config.NewLoader()}
	runCommand, _ := newRunCommand()

	err := command.run(runCommand, nil)
	require.Error(t, err)
	assert.ErrorContains(t, err, "invalid azctx path")
}
