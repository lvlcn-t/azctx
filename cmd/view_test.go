package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestViewMerged(t *testing.T) {
	writeConfigForTest(t, baseConfig())

	stdout, _, err := executeCommand(t, newViewCmd())
	require.NoError(t, err)

	assert.Contains(t, stdout, "SECTION")
	assert.Contains(t, stdout, "current-context")
	assert.Contains(t, stdout, "dev")
	assert.Contains(t, stdout, "tenant")
	assert.Contains(t, stdout, "credential")
	assert.Contains(t, stdout, "context")
}
