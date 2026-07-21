package login

import (
	"context"
	"errors"
	"testing"

	"github.com/lvlcn-t/azctx/az"
	"github.com/lvlcn-t/azctx/config"
	"github.com/lvlcn-t/azctx/wif"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Switch(t *testing.T) {
	path := writeConfig(t, baseConfig())
	store := loadStore(t)

	mock := staticClient(func() error { return nil })
	mock.WithTenantFunc = func(id string) az.CLI {
		assert.Equal(t, "tenant-2", id)
		return mock
	}
	mock.WithCredentialFunc = func(cred *config.Credential) az.CLI {
		assert.Equal(t, "sp", cred.Name)
		return mock
	}
	mock.WithSubscriptionFunc = func(id string) az.CLI {
		assert.Equal(t, "sub-prod", id)
		return mock
	}

	client := &Manager{
		NewClient: func(_ context.Context) (az.CLI, error) { return mock, nil },
		Writer:    config.NewWriter(),
	}

	require.NoError(t, client.Login(t.Context(), store, prodContext))
	assert.Len(t, mock.LoginCalls(), 1)
	assert.Equal(t, prodContext, readConfig(t, path).CurrentContext)
}

func TestClient_Switch_ContextNotFound(t *testing.T) {
	writeConfig(t, baseConfig())
	store := loadStore(t)

	client := &Manager{
		NewClient: func(_ context.Context) (az.CLI, error) {
			return staticClient(func() error { return nil }), nil
		},
		Writer: config.NewWriter(),
	}

	require.Error(t, client.Login(t.Context(), store, "missing"))
}

func TestClient_Switch_ClientFactoryError(t *testing.T) {
	writeConfig(t, baseConfig())
	store := loadStore(t)

	sentinel := errors.New("new client failed")
	client := &Manager{
		NewClient: func(_ context.Context) (az.CLI, error) { return nil, sentinel },
		Writer:    config.NewWriter(),
	}

	err := client.Login(t.Context(), store, devContext)
	require.ErrorIs(t, err, sentinel)
}

// withWorkloadIdentity appends a workload-identity credential and a context
// that uses it to cfg, returning the context name.
func withWorkloadIdentity(cfg *config.Config) string {
	cfg.Credentials = append(cfg.Credentials, config.Credential{
		Name: "wi",
		Details: config.CredentialDetails{
			Type:  config.CredentialTypeWorkloadIdentity,
			Azure: config.AzureCredential{ClientID: "wi-client"},
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
		Name:    "wi-ctx",
		Details: config.ContextDetails{Tenant: "corp", Credential: "wi"},
	})

	return "wi-ctx"
}

func TestClient_Switch_AcquiresFederatedToken(t *testing.T) {
	cfg := baseConfig()
	name := withWorkloadIdentity(cfg)
	writeConfig(t, cfg)
	store := loadStore(t)

	const fakeToken = "eyJ.fake-id-token.sig"

	mock := staticClient(func() error { return nil })
	mock.WithFederatedTokenFunc = func(token string) az.CLI {
		assert.Equal(t, fakeToken, token)
		return mock
	}

	client := &Manager{
		NewClient: func(_ context.Context) (az.CLI, error) { return mock, nil },
		Tokens: func(_ context.Context, cfg config.TokenDetails, _ string) (wif.Provider, error) {
			assert.Equal(t, config.TokenSourceOAuth2, cfg.Source)
			return &wif.ProviderMock{
				AcquireTokenFunc: func(context.Context, ...wif.AcquireOption) (string, bool, error) {
					return fakeToken, false, nil
				},
			}, nil
		},
		Writer: config.NewWriter(),
	}

	require.NoError(t, client.Login(t.Context(), store, name))
	assert.Len(t, mock.WithFederatedTokenCalls(), 1)
	assert.Len(t, mock.LoginCalls(), 1)
}

func TestClient_Switch_RetriesWithFreshTokenOnCachedLoginFailure(t *testing.T) {
	cfg := baseConfig()
	name := withWorkloadIdentity(cfg)
	writeConfig(t, cfg)
	store := loadStore(t)

	logins := 0
	mock := staticClient(func() error {
		logins++
		if logins == 1 {
			return az.ErrLogin
		}
		return nil
	})

	acquisitions := 0
	forceRefreshed := false
	client := &Manager{
		NewClient: func(_ context.Context) (az.CLI, error) { return mock, nil },
		Tokens: func(context.Context, config.TokenDetails, string) (wif.Provider, error) {
			return &wif.ProviderMock{
				AcquireTokenFunc: func(_ context.Context, opts ...wif.AcquireOption) (string, bool, error) {
					acquisitions++
					if len(opts) > 0 {
						forceRefreshed = true
						return "fresh-token", false, nil
					}
					return "cached-token", true, nil
				},
			}, nil
		},
		Writer: config.NewWriter(),
	}

	require.NoError(t, client.Login(t.Context(), store, name))
	assert.Equal(t, 2, logins)
	assert.Equal(t, 2, acquisitions)
	assert.True(t, forceRefreshed, "expected the retry to force-refresh the token")
}
