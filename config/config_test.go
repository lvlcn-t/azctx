package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	ciName      = "ci"
	clientIDVal = "client-id"
)

func TestLookupHelpers(t *testing.T) {
	cfg := &Config{
		Tenants:     []Tenant{{Name: tenantCorp.Name, Details: TenantDetails{ID: "tenant-id"}}},
		Credentials: []Credential{{Name: ciName, Details: CredentialDetails{Type: CredentialTypeServicePrincipal}}},
		Contexts:    []Context{{Name: devContext.Name, Details: ContextDetails{Tenant: tenantCorp.Name, Credential: ciName}}},
	}

	tenant, ok := cfg.TenantByName(tenantCorp.Name)
	require.True(t, ok)
	assert.Equal(t, "tenant-id", tenant.Details.ID)

	credential, ok := cfg.CredentialByName(ciName)
	require.True(t, ok)
	assert.Equal(t, CredentialTypeServicePrincipal, credential.Details.Type)

	contextValue, ok := cfg.ContextByName(devContext.Name)
	require.True(t, ok)
	assert.Equal(t, tenantCorp.Name, contextValue.Details.Tenant)

	_, ok = cfg.TenantByName("missing")
	assert.False(t, ok)
	_, ok = cfg.CredentialByName("missing")
	assert.False(t, ok)
	_, ok = cfg.ContextByName("missing")
	assert.False(t, ok)
}

func TestUpsertsAndDeletes(t *testing.T) {
	cfg := &Config{}
	cfg.UpsertTenant(Tenant{Name: tenantCorp.Name, Details: TenantDetails{ID: tenantCorp.Details.ID}})
	cfg.UpsertTenant(Tenant{Name: tenantCorp.Name, Details: TenantDetails{ID: tenantPlat.Details.ID}})

	credential := &Credential{Name: ciName, Details: CredentialDetails{Type: CredentialTypeUser}}
	cfg.UpsertCredential(credential)
	credential = &Credential{Name: ciName, Details: CredentialDetails{Type: CredentialTypeManagedIdentity}}
	cfg.UpsertCredential(credential)

	cfg.UpsertContext(Context{Name: devContext.Name, Details: ContextDetails{Tenant: tenantCorp.Name, Credential: ciName}})
	cfg.UpsertContext(Context{Name: devContext.Name, Details: ContextDetails{Tenant: tenantCorp.Name, Credential: ciName, Subscription: "sub-1"}})

	require.Len(t, cfg.Tenants, 1)
	assert.Equal(t, tenantPlat.Details.ID, cfg.Tenants[0].Details.ID)
	require.Len(t, cfg.Credentials, 1)
	assert.Equal(t, CredentialTypeManagedIdentity, cfg.Credentials[0].Details.Type)
	require.Len(t, cfg.Contexts, 1)
	assert.Equal(t, "sub-1", cfg.Contexts[0].Details.Subscription)

	assert.True(t, cfg.DeleteContext(devContext.Name))
	assert.False(t, cfg.DeleteContext(devContext.Name))
}

func TestRenameContext(t *testing.T) {
	cfg := &Config{
		Contexts: []Context{{Name: "old", Details: ContextDetails{Tenant: tenantCorp.Name, Credential: ciName}}},
	}

	assert.True(t, cfg.RenameContext("old", "new"))
	assert.Equal(t, "new", cfg.Contexts[0].Name)
	assert.False(t, cfg.RenameContext("old", "other"))
}

func TestMerge(t *testing.T) {
	base := &Config{
		APIVersion:     apiVersion,
		Kind:           kind,
		CurrentContext: devContext.Name,
		Tenants:        []Tenant{{Name: tenantCorp.Name, Details: TenantDetails{ID: tenantCorp.Details.ID}}},
		Credentials:    []Credential{{Name: ciName, Details: CredentialDetails{Type: CredentialTypeUser}}},
		Contexts:       []Context{{Name: devContext.Name, Details: ContextDetails{Tenant: tenantCorp.Name, Credential: ciName}}},
	}

	next := &Config{
		APIVersion:     apiVersion,
		Kind:           kind,
		CurrentContext: prodContext.Name,
		Tenants: []Tenant{
			{Name: tenantCorp.Name, Details: TenantDetails{ID: tenantPlat.Details.ID}},
			{Name: tenantPlat.Name, Details: TenantDetails{ID: "tenant-3"}},
		},
		Credentials: []Credential{
			{Name: ciName, Details: CredentialDetails{Type: CredentialTypeManagedIdentity}},
			{Name: "ops", Details: CredentialDetails{Type: CredentialTypeUser}},
		},
		Contexts: []Context{
			{Name: devContext.Name, Details: ContextDetails{Tenant: tenantCorp.Name, Credential: ciName, Subscription: "sub-a"}},
			{Name: prodContext.Name, Details: ContextDetails{Tenant: tenantPlat.Name, Credential: "ops"}},
		},
	}

	err := base.Merge(next)
	require.NoError(t, err)

	assert.Equal(t, devContext.Name, base.CurrentContext)
	require.Len(t, base.Tenants, 2)
	require.Len(t, base.Credentials, 2)
	require.Len(t, base.Contexts, 2)

	tenant, ok := base.TenantByName(tenantCorp.Name)
	require.True(t, ok)
	assert.Equal(t, tenantCorp.Details.ID, tenant.Details.ID)

	contextValue, ok := base.ContextByName(devContext.Name)
	require.True(t, ok)
	assert.Empty(t, contextValue.Details.Subscription)
}

func TestValidateContextReferences(t *testing.T) {
	cfg := &Config{
		Tenants:     []Tenant{{Name: tenantCorp.Name, Details: TenantDetails{ID: tenantCorp.Details.ID}}},
		Credentials: []Credential{{Name: ciName, Details: CredentialDetails{Type: CredentialTypeUser}}},
	}

	tests := []struct {
		name    string
		context Context
		wantErr string
	}{
		{
			name:    "valid",
			context: Context{Name: devContext.Name, Details: ContextDetails{Tenant: tenantCorp.Name, Credential: ciName}},
		},
		{
			name:    "missing name",
			context: Context{Details: ContextDetails{Tenant: tenantCorp.Name, Credential: ciName}},
			wantErr: "context name is required",
		},
		{
			name:    "missing tenant",
			context: Context{Name: devContext.Name, Details: ContextDetails{Credential: ciName}},
			wantErr: "context tenant is required",
		},
		{
			name:    "missing credential",
			context: Context{Name: devContext.Name, Details: ContextDetails{Tenant: tenantCorp.Name}},
			wantErr: "context credential is required",
		},
		{
			name:    "unknown tenant",
			context: Context{Name: devContext.Name, Details: ContextDetails{Tenant: "missing", Credential: ciName}},
			wantErr: "tenant \"missing\" does not exist",
		},
		{
			name:    "unknown credential",
			context: Context{Name: devContext.Name, Details: ContextDetails{Tenant: tenantCorp.Name, Credential: "missing"}},
			wantErr: "credential \"missing\" does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cfg.ValidateContextReferences(tt.context)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.EqualError(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}
