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

	var mock *az.CLIMock
	mock = &az.CLIMock{
		WithTenantFunc: func(tenantID string) az.CLI {
			assert.Equal(t, "tenant-2", tenantID)
			return mock
		},
		WithCredentialFunc: func(credential *config.Credential) az.CLI {
			assert.Equal(t, "sp", credential.Name)
			return mock
		},
		WithSubscriptionFunc: func(subscriptionID string) az.CLI {
			assert.Equal(t, "sub-prod", subscriptionID)
			return mock
		},
		AllowNoSubscriptionsFunc: func(allow bool) az.CLI {
			return mock
		},
		LoginFunc: func(ctx context.Context) error {
			return nil
		},
	}

	command := &useCommand{
		loader: config.NewLoader(),
		writer: config.NewWriter(),
		az: func(ctx context.Context) (az.CLI, error) {
			return mock, nil
		},
	}

	runCommand, stdout := newRunCommand()
	err := command.run(runCommand, []string{prodContext})
	require.NoError(t, err)

	assert.Contains(t, stdout.String(), fmt.Sprintf(`Switched to context %q.`, prodContext))
	assert.Len(t, mock.WithTenantCalls(), 1)
	assert.Len(t, mock.WithCredentialCalls(), 1)
	assert.Len(t, mock.WithSubscriptionCalls(), 1)
	assert.Len(t, mock.LoginCalls(), 1)

	got := readConfigForTest(t, path)
	assert.Equal(t, prodContext, got.CurrentContext)
}

func TestUseContextNotFound(t *testing.T) {
	writeConfigForTest(t, baseConfig())

	var mock *az.CLIMock
	mock = &az.CLIMock{
		WithTenantFunc: func(tenantID string) az.CLI {
			return mock
		},
		WithCredentialFunc: func(credential *config.Credential) az.CLI {
			return mock
		},
		WithSubscriptionFunc: func(subscriptionID string) az.CLI {
			return mock
		},
		AllowNoSubscriptionsFunc: func(allow bool) az.CLI {
			return mock
		},
		LoginFunc: func(ctx context.Context) error {
			return nil
		},
	}

	command := &useCommand{
		loader: config.NewLoader(),
		writer: config.NewWriter(),
		az: func(ctx context.Context) (az.CLI, error) {
			return mock, nil
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
		az: func(ctx context.Context) (az.CLI, error) {
			return nil, sentinel
		},
	}

	runCommand, _ := newRunCommand()
	err := command.run(runCommand, []string{devContext})
	require.Error(t, err)
	assert.ErrorContains(t, err, "create az client")
	assert.ErrorIs(t, err, sentinel)
}
