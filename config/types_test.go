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

func TestParseCredentialType(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    CredentialType
		wantErr string
	}{
		{name: "service principal", raw: "service-principal", want: CredentialTypeServicePrincipal},
		{name: "user", raw: "user", want: CredentialTypeUser}, //nolint:goconst // test input
		{name: "managed identity", raw: "managed-identity", want: CredentialTypeManagedIdentity},
		{name: "oidc", raw: "oidc", want: CredentialTypeOIDC}, //nolint:goconst // test input
		{name: "empty", raw: "", wantErr: "credential type is required"},
		{name: "unsupported", raw: "something-else", wantErr: "unsupported credential type"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCredentialType(tt.raw)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCredentialValidate(t *testing.T) {
	tests := []struct {
		name    string
		input   *Credential
		wantErr string
	}{
		{
			name: "service principal with secret",
			input: &Credential{
				Name:         ciName,
				Type:         CredentialTypeServicePrincipal,
				ClientID:     clientIDVal,
				ClientSecret: "secret",
			},
		},
		{
			name: "service principal with certificate",
			input: &Credential{
				Name:                  ciName,
				Type:                  CredentialTypeServicePrincipal,
				ClientID:              clientIDVal,
				ClientCertificatePath: "/tmp/cert.pem",
			},
		},
		{
			name: "service principal missing id",
			input: &Credential{
				Name:         ciName,
				Type:         CredentialTypeServicePrincipal,
				ClientSecret: "secret",
			},
			wantErr: "requires client-id",
		},
		{
			name: "service principal missing auth material",
			input: &Credential{
				Name:     ciName,
				Type:     CredentialTypeServicePrincipal,
				ClientID: clientIDVal,
			},
			wantErr: "requires client-secret or client-certificate-path",
		},
		{
			name: "user",
			input: &Credential{
				Name: devContext.Name,
				Type: CredentialTypeUser,
			},
		},
		{
			name: "managed identity",
			input: &Credential{
				Name: "mi",
				Type: CredentialTypeManagedIdentity,
			},
		},
		{
			name: "oidc",
			input: &Credential{
				Name:               "oidc",
				Type:               CredentialTypeOIDC,
				ClientID:           clientIDVal,
				FederatedTokenFile: "/tmp/token",
			},
		},
		{
			name: "oidc missing token file",
			input: &Credential{
				Name:     "oidc",
				Type:     CredentialTypeOIDC,
				ClientID: clientIDVal,
			},
			wantErr: "requires federated-token-file",
		},
		{name: "nil credential", input: nil, wantErr: "credential is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestLookupHelpers(t *testing.T) {
	cfg := &Config{
		Tenants:     []Tenant{{Name: tenantCorp.Name, ID: "tenant-id"}},
		Credentials: []Credential{{Name: ciName, Type: CredentialTypeServicePrincipal}},
		Contexts:    []Context{{Name: devContext.Name, Tenant: tenantCorp.Name, Credential: ciName}},
	}

	tenant, ok := cfg.TenantByName(tenantCorp.Name)
	require.True(t, ok)
	assert.Equal(t, "tenant-id", tenant.ID)

	credential, ok := cfg.CredentialByName(ciName)
	require.True(t, ok)
	assert.Equal(t, CredentialTypeServicePrincipal, credential.Type)

	contextValue, ok := cfg.ContextByName(devContext.Name)
	require.True(t, ok)
	assert.Equal(t, tenantCorp.Name, contextValue.Tenant)

	_, ok = cfg.TenantByName("missing")
	assert.False(t, ok)
	_, ok = cfg.CredentialByName("missing")
	assert.False(t, ok)
	_, ok = cfg.ContextByName("missing")
	assert.False(t, ok)
}

func TestUpsertsAndDeletes(t *testing.T) {
	cfg := &Config{}
	cfg.UpsertTenant(Tenant{Name: tenantCorp.Name, ID: tenantCorp.ID})
	cfg.UpsertTenant(Tenant{Name: tenantCorp.Name, ID: tenantPlat.ID})

	credential := &Credential{Name: ciName, Type: CredentialTypeUser}
	cfg.UpsertCredential(credential)
	credential = &Credential{Name: ciName, Type: CredentialTypeManagedIdentity}
	cfg.UpsertCredential(credential)

	cfg.UpsertContext(Context{Name: devContext.Name, Tenant: tenantCorp.Name, Credential: ciName})
	cfg.UpsertContext(Context{Name: devContext.Name, Tenant: tenantCorp.Name, Credential: ciName, Subscription: "sub-1"})

	require.Len(t, cfg.Tenants, 1)
	assert.Equal(t, tenantPlat.ID, cfg.Tenants[0].ID)
	require.Len(t, cfg.Credentials, 1)
	assert.Equal(t, CredentialTypeManagedIdentity, cfg.Credentials[0].Type)
	require.Len(t, cfg.Contexts, 1)
	assert.Equal(t, "sub-1", cfg.Contexts[0].Subscription)

	assert.True(t, cfg.DeleteContext(devContext.Name))
	assert.False(t, cfg.DeleteContext(devContext.Name))
}

func TestRenameContext(t *testing.T) {
	cfg := &Config{
		Contexts: []Context{{Name: "old", Tenant: tenantCorp.Name, Credential: ciName}},
	}

	assert.True(t, cfg.RenameContext("old", "new"))
	assert.Equal(t, "new", cfg.Contexts[0].Name)
	assert.False(t, cfg.RenameContext("old", "other"))
}

func TestMerge(t *testing.T) {
	base := &Config{
		CurrentContext: devContext.Name,
		Tenants:        []Tenant{{Name: tenantCorp.Name, ID: tenantCorp.ID}},
		Credentials:    []Credential{{Name: ciName, Type: CredentialTypeUser}},
		Contexts:       []Context{{Name: devContext.Name, Tenant: tenantCorp.Name, Credential: ciName}},
	}

	next := &Config{
		CurrentContext: prodContext.Name,
		Tenants: []Tenant{
			{Name: tenantCorp.Name, ID: tenantPlat.ID},
			{Name: tenantPlat.Name, ID: "tenant-3"},
		},
		Credentials: []Credential{
			{Name: ciName, Type: CredentialTypeManagedIdentity},
			{Name: "ops", Type: CredentialTypeUser},
		},
		Contexts: []Context{
			{Name: devContext.Name, Tenant: tenantCorp.Name, Credential: ciName, Subscription: "sub-a"},
			{Name: prodContext.Name, Tenant: tenantPlat.Name, Credential: "ops"},
		},
	}

	base.Merge(next)

	assert.Equal(t, devContext.Name, base.CurrentContext)
	require.Len(t, base.Tenants, 2)
	require.Len(t, base.Credentials, 2)
	require.Len(t, base.Contexts, 2)

	tenant, ok := base.TenantByName(tenantCorp.Name)
	require.True(t, ok)
	assert.Equal(t, tenantCorp.ID, tenant.ID)

	contextValue, ok := base.ContextByName(devContext.Name)
	require.True(t, ok)
	assert.Empty(t, contextValue.Subscription)
}

func TestValidateContextReferences(t *testing.T) {
	cfg := &Config{
		Tenants:     []Tenant{{Name: tenantCorp.Name, ID: tenantCorp.ID}},
		Credentials: []Credential{{Name: ciName, Type: CredentialTypeUser}},
	}

	tests := []struct {
		name    string
		context Context
		wantErr string
	}{
		{
			name:    "valid",
			context: Context{Name: devContext.Name, Tenant: tenantCorp.Name, Credential: ciName},
		},
		{
			name:    "missing name",
			context: Context{Tenant: tenantCorp.Name, Credential: ciName},
			wantErr: "context name is required",
		},
		{
			name:    "missing tenant",
			context: Context{Name: devContext.Name, Credential: ciName},
			wantErr: "context tenant is required",
		},
		{
			name:    "missing credential",
			context: Context{Name: devContext.Name, Tenant: tenantCorp.Name},
			wantErr: "context credential is required",
		},
		{
			name:    "unknown tenant",
			context: Context{Name: devContext.Name, Tenant: "missing", Credential: ciName},
			wantErr: "tenant \"missing\" does not exist",
		},
		{
			name:    "unknown credential",
			context: Context{Name: devContext.Name, Tenant: tenantCorp.Name, Credential: "missing"},
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
