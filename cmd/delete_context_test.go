package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteContextFound(t *testing.T) {
	path := writeConfigForTest(t, baseConfig())

	stdout, _, err := executeCommand(t, newDeleteCtxCmd(), "dev")
	require.NoError(t, err)

	assert.Contains(t, stdout, `deleted context "dev"`)

	got := readConfigForTest(t, path)
	_, found := got.ContextByName("dev")
	assert.False(t, found)
}

func TestDeleteContextNotFound(t *testing.T) {
	writeConfigForTest(t, baseConfig())

	_, _, err := executeCommand(t, newDeleteCtxCmd(), "missing")
	require.Error(t, err)
	assert.ErrorContains(t, err, `context "missing" not found`)
}

func TestDeleteActiveContextWarning(t *testing.T) {
	writeConfigForTest(t, baseConfig())

	_, stderr, err := executeCommand(t, newDeleteCtxCmd(), "dev")
	require.NoError(t, err)

	assert.Contains(t, stderr, "warning: this removed your active context")
}
