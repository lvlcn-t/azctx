package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetByName(t *testing.T) {
	writeConfigForTest(t, baseConfig())

	stdout, _, err := executeCommand(t, newGetCmd(), "prod")
	require.NoError(t, err)

	assert.Contains(t, stdout, "name: prod")
	assert.Contains(t, stdout, "tenant: platform")
}

func TestGetCurrentContext(t *testing.T) {
	writeConfigForTest(t, baseConfig())

	stdout, _, err := executeCommand(t, newGetCmd())
	require.NoError(t, err)

	assert.Contains(t, stdout, "name: dev")
}

func TestGetNotFound(t *testing.T) {
	writeConfigForTest(t, baseConfig())

	_, _, err := executeCommand(t, newGetCmd(), "unknown")
	require.Error(t, err)
	assert.ErrorContains(t, err, `context "unknown" not found`)
}
