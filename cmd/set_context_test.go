package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSetContextWiring verifies the cobra glue: flags are parsed and the
// created message is emitted. Merge and validation logic is covered in the contexts package.
func TestSetContextWiring(t *testing.T) {
	writeConfig(t, baseConfig())

	stdout, _, err := execCmd(
		t,
		newSetCtxCmd(),
		"stage",
		"--tenant", "corp",
		"--credential", "user",
		"--subscription", "sub-stage",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, `Context "stage" created.`)
}
