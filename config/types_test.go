package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCredentialType(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    CredentialType
		wantErr string
	}{
		{name: "service principal", raw: "service-principal", want: CredentialTypeServicePrincipal},
		{name: "user", raw: "user", want: CredentialTypeUser},
		{name: "managed identity", raw: "managed-identity", want: CredentialTypeManagedIdentity},
		{name: "oidc", raw: "oidc", want: CredentialTypeOIDC},
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
				Name:         "ci",
				Type:         CredentialTypeServicePrincipal,
				ClientID:     "client-id",
				ClientSecret: "secret",
			},
		},
		{
			name: "service principal with certificate",
			input: &Credential{
				Name:                  "ci",
				Type:                  CredentialTypeServicePrincipal,
				ClientID:              "client-id",
				ClientCertificatePath: "/tmp/cert.pem",
			},
		},
		{
			name: "service principal missing id",
			input: &Credential{
				Name:         "ci",
				Type:         CredentialTypeServicePrincipal,
				ClientSecret: "secret",
			},
			wantErr: "requires client-id",
		},
		{
			name: "service principal missing auth material",
			input: &Credential{
				Name:     "ci",
				Type:     CredentialTypeServicePrincipal,
				ClientID: "client-id",
			},
			wantErr: "requires client-secret or client-certificate-path",
		},
		{
			name: "user",
			input: &Credential{
				Name: "dev",
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
				ClientID:           "client-id",
				FederatedTokenFile: "/tmp/token",
			},
		},
		{
			name: "oidc missing token file",
			input: &Credential{
				Name:     "oidc",
				Type:     CredentialTypeOIDC,
				ClientID: "client-id",
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
		Tenants:     []Tenant{{Name: "corp", ID: "tenant-id"}},
		Credentials: []Credential{{Name: "ci", Type: CredentialTypeServicePrincipal}},
		Contexts:    []Context{{Name: "dev", Tenant: "corp", Credential: "ci"}},
	}

	tenant, ok := cfg.TenantByName("corp")
	require.True(t, ok)
	assert.Equal(t, "tenant-id", tenant.ID)

	credential, ok := cfg.CredentialByName("ci")
	require.True(t, ok)
	assert.Equal(t, CredentialTypeServicePrincipal, credential.Type)

	contextValue, ok := cfg.ContextByName("dev")
	require.True(t, ok)
	assert.Equal(t, "corp", contextValue.Tenant)

	_, ok = cfg.TenantByName("missing")
	assert.False(t, ok)
	_, ok = cfg.CredentialByName("missing")
	assert.False(t, ok)
	_, ok = cfg.ContextByName("missing")
	assert.False(t, ok)
}

func TestUpsertsAndDeletes(t *testing.T) {
	cfg := &Config{}
	cfg.UpsertTenant(Tenant{Name: "corp", ID: "tenant-1"})
	cfg.UpsertTenant(Tenant{Name: "corp", ID: "tenant-2"})

	credential := &Credential{Name: "ci", Type: CredentialTypeUser}
	cfg.UpsertCredential(credential)
	credential = &Credential{Name: "ci", Type: CredentialTypeManagedIdentity}
	cfg.UpsertCredential(credential)

	cfg.UpsertContext(Context{Name: "dev", Tenant: "corp", Credential: "ci"})
	cfg.UpsertContext(Context{Name: "dev", Tenant: "corp", Credential: "ci", Subscription: "sub-1"})

	require.Len(t, cfg.Tenants, 1)
	assert.Equal(t, "tenant-2", cfg.Tenants[0].ID)
	require.Len(t, cfg.Credentials, 1)
	assert.Equal(t, CredentialTypeManagedIdentity, cfg.Credentials[0].Type)
	require.Len(t, cfg.Contexts, 1)
	assert.Equal(t, "sub-1", cfg.Contexts[0].Subscription)

	assert.True(t, cfg.DeleteContext("dev"))
	assert.False(t, cfg.DeleteContext("dev"))
}

func TestRenameContext(t *testing.T) {
	cfg := &Config{
		Contexts: []Context{{Name: "old", Tenant: "corp", Credential: "ci"}},
	}

	assert.True(t, cfg.RenameContext("old", "new"))
	assert.Equal(t, "new", cfg.Contexts[0].Name)
	assert.False(t, cfg.RenameContext("old", "other"))
}

func TestMerge(t *testing.T) {
	base := &Config{
		CurrentContext: "dev",
		Tenants:        []Tenant{{Name: "corp", ID: "tenant-1"}},
		Credentials:    []Credential{{Name: "ci", Type: CredentialTypeUser}},
		Contexts:       []Context{{Name: "dev", Tenant: "corp", Credential: "ci"}},
	}

	next := &Config{
		CurrentContext: "prod",
		Tenants: []Tenant{
			{Name: "corp", ID: "tenant-2"},
			{Name: "platform", ID: "tenant-3"},
		},
		Credentials: []Credential{
			{Name: "ci", Type: CredentialTypeManagedIdentity},
			{Name: "ops", Type: CredentialTypeUser},
		},
		Contexts: []Context{
			{Name: "dev", Tenant: "corp", Credential: "ci", Subscription: "sub-a"},
			{Name: "prod", Tenant: "platform", Credential: "ops"},
		},
	}

	base.Merge(next)

	assert.Equal(t, "dev", base.CurrentContext)
	require.Len(t, base.Tenants, 2)
	require.Len(t, base.Credentials, 2)
	require.Len(t, base.Contexts, 2)

	tenant, ok := base.TenantByName("corp")
	require.True(t, ok)
	assert.Equal(t, "tenant-1", tenant.ID)

	contextValue, ok := base.ContextByName("dev")
	require.True(t, ok)
	assert.Empty(t, contextValue.Subscription)
}

func TestValidateContextReferences(t *testing.T) {
	cfg := &Config{
		Tenants:     []Tenant{{Name: "corp", ID: "tenant-1"}},
		Credentials: []Credential{{Name: "ci", Type: CredentialTypeUser}},
	}

	tests := []struct {
		name    string
		context Context
		wantErr string
	}{
		{
			name:    "valid",
			context: Context{Name: "dev", Tenant: "corp", Credential: "ci"},
		},
		{
			name:    "missing name",
			context: Context{Tenant: "corp", Credential: "ci"},
			wantErr: "context name is required",
		},
		{
			name:    "missing tenant",
			context: Context{Name: "dev", Credential: "ci"},
			wantErr: "context tenant is required",
		},
		{
			name:    "missing credential",
			context: Context{Name: "dev", Tenant: "corp"},
			wantErr: "context credential is required",
		},
		{
			name:    "unknown tenant",
			context: Context{Name: "dev", Tenant: "missing", Credential: "ci"},
			wantErr: "tenant \"missing\" does not exist",
		},
		{
			name:    "unknown credential",
			context: Context{Name: "dev", Tenant: "corp", Credential: "missing"},
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
