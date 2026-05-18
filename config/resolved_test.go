package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_Resolve(t *testing.T) {
	store := &Store{
		Config: Config{
			Tenants: []Tenant{
				{Name: "my-tenant", Details: TenantDetails{ID: "tid-123"}},
			},
			Credentials: []Credential{
				{Name: "my-cred", Details: CredentialDetails{Type: CredentialTypeUser}},
			},
			Contexts: []Context{
				{
					Name: "dev",
					Details: ContextDetails{
						Tenant:               "my-tenant",
						Credential:           "my-cred",
						Subscription:         "sub-456",
						AllowNoSubscriptions: true,
					},
				},
			},
		},
	}

	resolved, err := store.Resolve("dev")
	require.NoError(t, err)
	assert.Equal(t, "dev", resolved.Name)
	assert.Equal(t, "tid-123", resolved.Tenant.Details.ID)
	assert.Equal(t, "my-cred", resolved.Credential.Name)
	assert.Equal(t, "sub-456", resolved.Subscription)
	assert.True(t, resolved.AllowNoSubscriptions)
}

func TestStore_Resolve_ContextNotFound(t *testing.T) {
	store := &Store{Config: Config{}}
	_, err := store.Resolve("nope")
	require.ErrorContains(t, err, `context "nope" not found`)
}

func TestStore_Resolve_TenantNotFound(t *testing.T) {
	store := &Store{
		Config: Config{
			Contexts: []Context{
				{Name: "ctx", Details: ContextDetails{Tenant: "missing", Credential: "c"}},
			},
		},
	}
	_, err := store.Resolve("ctx")
	require.ErrorContains(t, err, `tenant "missing" not found`)
}

func TestStore_Resolve_CredentialNotFound(t *testing.T) {
	store := &Store{
		Config: Config{
			Tenants:  []Tenant{{Name: "t", Details: TenantDetails{ID: "id"}}},
			Contexts: []Context{{Name: "ctx", Details: ContextDetails{Tenant: "t", Credential: "missing"}}},
		},
	}
	_, err := store.Resolve("ctx")
	require.ErrorContains(t, err, `credential "missing" not found`)
}

func TestStore_Resolve_TenantMissingID(t *testing.T) {
	store := &Store{
		Config: Config{
			Tenants:     []Tenant{{Name: "t", Details: TenantDetails{}}},
			Credentials: []Credential{{Name: "c"}},
			Contexts:    []Context{{Name: "ctx", Details: ContextDetails{Tenant: "t", Credential: "c"}}},
		},
	}
	_, err := store.Resolve("ctx")
	require.ErrorContains(t, err, `tenant "t" is missing id`)
}
