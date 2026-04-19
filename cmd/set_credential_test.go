package cmd

import (
	"testing"

	"github.com/lvlcn-t/azctx/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetCredentialCreate(t *testing.T) {
	path := writeConfigForTest(t, baseConfig())

	stdout, _, err := executeCommand(
		t,
		newSetCredentialCmd(),
		"new-sp",
		"--type",
		"service-principal",
		"--client-id",
		"client-new",
		"--client-secret",
		"secret-new",
	)
	require.NoError(t, err)

	assert.Contains(t, stdout, `Credential "new-sp" created.`)

	got := readConfigForTest(t, path)
	credential, found := got.CredentialByName("new-sp")
	require.True(t, found)
	assert.Equal(t, config.CredentialTypeServicePrincipal, credential.Type)
	assert.Equal(t, "client-new", credential.ClientID)
}

func TestSetCredentialInvalidType(t *testing.T) {
	writeConfigForTest(t, baseConfig())

	_, _, err := executeCommand(t, newSetCredentialCmd(), "bad", "--type", "invalid")
	require.Error(t, err)
	assert.ErrorContains(t, err, `unsupported credential type "invalid"`)
}
