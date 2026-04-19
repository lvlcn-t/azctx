package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetContextCreate(t *testing.T) {
	path := writeConfigForTest(t, baseConfig())

	stdout, _, err := executeCommand(
		t,
		newSetCtxCmd(),
		"stage",
		"--tenant",
		"corp",
		"--credential",
		"user",
		"--subscription",
		"sub-stage",
	)
	require.NoError(t, err)

	assert.Contains(t, stdout, `Context "stage" created.`)

	got := readConfigForTest(t, path)
	contextValue, found := got.ContextByName("stage")
	require.True(t, found)
	assert.Equal(t, "corp", contextValue.Tenant)
	assert.Equal(t, "user", contextValue.Credential)
	assert.Equal(t, "sub-stage", contextValue.Subscription)
}

func TestSetContextUpdate(t *testing.T) {
	path := writeConfigForTest(t, baseConfig())

	stdout, _, err := executeCommand(
		t,
		newSetCtxCmd(),
		"dev",
		"--tenant",
		"platform",
		"--credential",
		"sp",
	)
	require.NoError(t, err)

	assert.Contains(t, stdout, `Context "dev" modified.`)

	got := readConfigForTest(t, path)
	contextValue, found := got.ContextByName("dev")
	require.True(t, found)
	assert.Equal(t, "platform", contextValue.Tenant)
	assert.Equal(t, "sp", contextValue.Credential)
}

func TestSetContextMissingTenant(t *testing.T) {
	writeConfigForTest(t, baseConfig())

	_, _, err := executeCommand(
		t,
		newSetCtxCmd(),
		"bad",
		"--tenant",
		"missing",
		"--credential",
		"user",
	)
	require.Error(t, err)
	assert.ErrorContains(t, err, `tenant "missing" does not exist`)
}
