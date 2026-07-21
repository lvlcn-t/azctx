package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSetCredentialWiring verifies the cobra glue for set-credential. Validation
// and CRUD logic is covered in the contexts package.
func TestSetCredentialWiring(t *testing.T) {
	writeConfig(t, baseConfig())

	stdout, _, err := execCmd(
		t,
		newSetCredentialCmd(),
		"new-sp",
		"--type", "service-principal",
		"--client-id", "client-new",
		"--client-secret", "secret-new",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, `Credential "new-sp" created.`)
}
