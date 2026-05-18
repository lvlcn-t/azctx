package cmd

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/lvlcn-t/azctx/az"
	"github.com/lvlcn-t/azctx/config"
	"github.com/lvlcn-t/azctx/wif"
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
		WithFederatedTokenFunc: func(string) az.CLI {
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
		WithFederatedTokenFunc: func(string) az.CLI {
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

func TestUseOAuth2TokenAcquisition(t *testing.T) {
	cfg := baseConfig()
	cfg.Credentials = append(cfg.Credentials, config.Credential{
		Name: "wi",
		Credential: config.CredentialDetails{
			Type: config.CredentialTypeWorkloadIdentity,
			Azure: config.AzureCredential{
				ClientID: "wi-client",
			},
			Token: config.TokenDetails{
				Source: config.TokenSourceOAuth2,
				OAuth2: &config.OAuth2Source{
					Issuer:   "https://login.example.com",
					ClientID: "oauth-client",
					Scopes:   []string{"openid"},
				},
			},
		},
	})
	cfg.Contexts = append(cfg.Contexts, config.Context{
		Name: "wi-ctx",
		Context: config.ContextDetails{
			Tenant:       "corp",
			Credential:   "wi",
			Subscription: "sub-wi",
		},
	})
	writeConfigForTest(t, cfg)

	const fakeToken = "eyJ.fake-id-token.sig"

	var mock *az.CLIMock
	mock = &az.CLIMock{
		WithTenantFunc:           func(string) az.CLI { return mock },
		WithCredentialFunc:       func(*config.Credential) az.CLI { return mock },
		WithSubscriptionFunc:     func(string) az.CLI { return mock },
		AllowNoSubscriptionsFunc: func(bool) az.CLI { return mock },
		WithFederatedTokenFunc: func(token string) az.CLI {
			assert.Equal(t, fakeToken, token)
			return mock
		},
		LoginFunc: func(context.Context) error { return nil },
	}

	command := &useCommand{
		loader: config.NewLoader(),
		writer: config.NewWriter(),
		az:     func(context.Context) (az.CLI, error) { return mock, nil },
		wif: func(_ context.Context, cfg config.TokenDetails) (wif.Provider, error) {
			assert.Equal(t, config.TokenSourceOAuth2, cfg.Source)
			return &wif.ProviderMock{
				AcquireTokenFunc: func(context.Context) (string, error) {
					return fakeToken, nil
				},
			}, nil
		},
	}

	runCommand, stdout := newRunCommand()
	err := command.run(runCommand, []string{"wi-ctx"})
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), `Switched to context "wi-ctx".`)
	assert.Len(t, mock.WithFederatedTokenCalls(), 1)
	assert.Len(t, mock.LoginCalls(), 1)
}
