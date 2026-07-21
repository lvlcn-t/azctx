package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSetTenantWiring verifies the cobra glue for set-tenant. CRUD logic is
// covered in the contexts package.
func TestSetTenantWiring(t *testing.T) {
	writeConfig(t, baseConfig())

	stdout, _, err := execCmd(t, newSetTenantCmd(), "new-tenant", "--id", "tenant-99")
	require.NoError(t, err)
	assert.Contains(t, stdout, `Tenant "new-tenant" created.`)
}
