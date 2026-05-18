package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenameContextHappyPath(t *testing.T) {
	path := writeConfigForTest(t, baseConfig())

	stdout, _, err := executeCommand(t, newRenameCtxCmd(), "dev", "local")
	require.NoError(t, err)

	assert.Contains(t, stdout, `Context "dev" renamed to "local".`)

	got := readConfigForTest(t, path)
	_, oldFound := got.ContextByName("dev")
	assert.False(t, oldFound)

	renamed, newFound := got.ContextByName("local")
	require.True(t, newFound)
	assert.Equal(t, "corp", renamed.Details.Tenant)
	assert.Equal(t, "local", got.CurrentContext)
}

func TestRenameContextOldNotFound(t *testing.T) {
	writeConfigForTest(t, baseConfig())

	_, _, err := executeCommand(t, newRenameCtxCmd(), "missing", "next")
	require.Error(t, err)
	assert.ErrorContains(t, err, `cannot rename context "missing", it does not exist`)
}

func TestRenameContextNewAlreadyExists(t *testing.T) {
	writeConfigForTest(t, baseConfig())

	_, _, err := executeCommand(t, newRenameCtxCmd(), "dev", "prod")
	require.Error(t, err)
	assert.ErrorContains(t, err, `context "prod" already exists`)
}
