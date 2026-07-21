package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDeleteContextWiring verifies the cobra glue for delete-context. Deletion
// logic is covered in the contexts package.
func TestDeleteContextWiring(t *testing.T) {
	writeConfig(t, baseConfig())

	stdout, _, err := execCmd(t, newDeleteCtxCmd(), "prod")
	require.NoError(t, err)
	assert.Contains(t, stdout, `deleted context "prod"`)
}

// TestDeleteActiveContextWarning verifies the stderr warning emitted by the
// command when the deleted context was the active one.
func TestDeleteActiveContextWarning(t *testing.T) {
	writeConfig(t, baseConfig())

	_, stderr, err := execCmd(t, newDeleteCtxCmd(), "dev")
	require.NoError(t, err)
	assert.Contains(t, stderr, "warning: this removed your active context")
}
