package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetTenantCreate(t *testing.T) {
	path := writeConfigForTest(t, baseConfig())

	stdout, _, err := executeCommand(
		t,
		newSetTenantCmd(),
		"new-tenant",
		"--id",
		"tenant-99",
	)
	require.NoError(t, err)

	assert.Contains(t, stdout, `Tenant "new-tenant" created.`)

	got := readConfigForTest(t, path)
	tenant, found := got.TenantByName("new-tenant")
	require.True(t, found)
	assert.Equal(t, "tenant-99", tenant.ID)
}

func TestSetTenantUpdate(t *testing.T) {
	path := writeConfigForTest(t, baseConfig())

	stdout, _, err := executeCommand(t, newSetTenantCmd(), "corp", "--id", "tenant-updated")
	require.NoError(t, err)

	assert.Contains(t, stdout, `Tenant "corp" modified.`)

	got := readConfigForTest(t, path)
	tenant, found := got.TenantByName("corp")
	require.True(t, found)
	assert.Equal(t, "tenant-updated", tenant.ID)
}
