package cmd

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	"github.com/lvlcn-t/azctx/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

const (
	devContext  = "dev"
	prodContext = "prod"
)

func writeConfigForTest(t *testing.T, cfg *config.Config) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "azctx.yaml")
	writer := config.NewWriter()
	require.NoError(t, writer.Write(path, cfg))

	t.Setenv(config.ConfigEnvVar, path)

	return path
}

func readConfigForTest(t *testing.T, path string) config.Config {
	t.Helper()

	loader := config.NewLoader()
	loaded, err := loader.Read(path)
	require.NoError(t, err)

	return loaded
}

func executeCommand(
	t *testing.T,
	command *cobra.Command,
	args ...string,
) (stdout, stderr string, err error) {
	t.Helper()

	var stdoutBuffer bytes.Buffer
	var stderrBuffer bytes.Buffer

	command.SetOut(&stdoutBuffer)
	command.SetErr(&stderrBuffer)
	command.SilenceErrors = true
	command.SilenceUsage = true
	command.SetArgs(args)

	err = command.Execute()

	return stdoutBuffer.String(), stderrBuffer.String(), err
}

func newRunCommand() (command *cobra.Command, stdout *bytes.Buffer) {
	stdout = &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	command = &cobra.Command{} //nolint:exhaustruct // minimal command for unit execution
	command.SetOut(stdout)
	command.SetErr(stderr)
	command.SetContext(context.Background())

	return command, stdout
}

func baseConfig() *config.Config {
	return &config.Config{
		CurrentContext: devContext,
		Tenants: []config.Tenant{
			{Name: "corp", Tenant: config.TenantDetails{ID: "tenant-1"}},
			{Name: "platform", Tenant: config.TenantDetails{ID: "tenant-2"}},
		},
		Credentials: []config.Credential{
			{Name: "user", Credential: config.CredentialDetails{Type: config.CredentialTypeUser}},
			{
				Name: "sp",
				Credential: config.CredentialDetails{
					Type: config.CredentialTypeServicePrincipal,
					Azure: config.AzureCredential{
						ClientID:     "client-1",
						ClientSecret: "secret-1",
					},
				},
			},
		},
		Contexts: []config.Context{
			{
				Name: devContext,
				Context: config.ContextDetails{
					Tenant:       "corp",
					Credential:   "user",
					Subscription: "sub-dev",
				},
			},
			{
				Name: prodContext,
				Context: config.ContextDetails{
					Tenant:       "platform",
					Credential:   "sp",
					Subscription: "sub-prod",
				},
			},
		},
	}
}
