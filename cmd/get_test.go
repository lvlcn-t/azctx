package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetByName(t *testing.T) {
	writeConfig(t, baseConfig())

	stdout, _, err := execCmd(t, newGetCmd(), "prod")
	require.NoError(t, err)

	assert.Contains(t, stdout, "name: prod")
	assert.Contains(t, stdout, "tenant: platform")
}

func TestGetCurrentContext(t *testing.T) {
	writeConfig(t, baseConfig())

	stdout, _, err := execCmd(t, newGetCmd())
	require.NoError(t, err)

	assert.Contains(t, stdout, "name: dev")
}

func TestGetNotFound(t *testing.T) {
	writeConfig(t, baseConfig())

	_, _, err := execCmd(t, newGetCmd(), "unknown")
	require.Error(t, err)
	assert.ErrorContains(t, err, `context "unknown" not found`)
}
