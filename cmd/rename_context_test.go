package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRenameContextWiring verifies the cobra glue for rename-context. Rename and
// current-context semantics are covered in the contexts package.
func TestRenameContextWiring(t *testing.T) {
	writeConfig(t, baseConfig())

	stdout, _, err := execCmd(t, newRenameCtxCmd(), "dev", "local")
	require.NoError(t, err)
	assert.Contains(t, stdout, `Context "dev" renamed to "local".`)
}
