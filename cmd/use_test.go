package cmd

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/lvlcn-t/azctx/az"
	"github.com/lvlcn-t/azctx/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUseHappyPath(t *testing.T) {
	path := writeConfigForTest(t, baseConfig())

	mock := &az.CLIMock{
		LoginFunc: func(_ context.Context, credential *config.Credential, tenantID string) error {
			require.Equal(t, "sp", credential.Name)
			require.Equal(t, "tenant-2", tenantID)
			return nil
		},
		SetSubscriptionFunc: func(_ context.Context, subscriptionID string) error {
			require.Equal(t, "sub-prod", subscriptionID)
			return nil
		},
	}

	command := &useCommand{
		loader: config.NewLoader(),
		writer: config.NewWriter(),
		az: func() (az.CLI, error) {
			return mock, nil
		},
	}

	runCommand, stdout := newRunCommand()
	err := command.run(runCommand, []string{prodContext})
	require.NoError(t, err)

	assert.Contains(t, stdout.String(), fmt.Sprintf(`Switched to context %q.`, prodContext))
	assert.Len(t, mock.LoginCalls(), 1)
	assert.Len(t, mock.SetSubscriptionCalls(), 1)

	got := readConfigForTest(t, path)
	assert.Equal(t, prodContext, got.CurrentContext)
}

func TestUseContextNotFound(t *testing.T) {
	writeConfigForTest(t, baseConfig())

	command := &useCommand{
		loader: config.NewLoader(),
		writer: config.NewWriter(),
		az: func() (az.CLI, error) {
			return &az.CLIMock{
				LoginFunc: func(context.Context, *config.Credential, string) error {
					return nil
				},
				SetSubscriptionFunc: func(context.Context, string) error {
					return nil
				},
			}, nil
		},
	}

	runCommand, _ := newRunCommand()
	err := command.run(runCommand, []string{"missing"})
	require.Error(t, err)
	assert.ErrorContains(t, err, `context "missing" not found`)
}

func TestUseAZClientFactoryError(t *testing.T) {
	writeConfigForTest(t, baseConfig())

	sentinel := errors.New("new client failed")
	command := &useCommand{
		loader: config.NewLoader(),
		writer: config.NewWriter(),
		az: func() (az.CLI, error) {
			return nil, sentinel
		},
	}

	runCommand, _ := newRunCommand()
	err := command.run(runCommand, []string{devContext})
	require.Error(t, err)
	assert.ErrorContains(t, err, "create az client")
	assert.ErrorIs(t, err, sentinel)
}
