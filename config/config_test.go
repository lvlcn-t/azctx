package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLookupHelpers(t *testing.T) {
	const (
		tenant     = "corp"
		credential = "sp"
		context    = "dev"
	)

	cfg := newTestConfig(t).
		withTenant(tenant).
		withSPCredential(credential).
		withContext(context, tenant, credential).
		build()

	tn, ok := cfg.TenantByName(tenant)
	require.True(t, ok)
	assert.NotEmpty(t, tn.Details.ID)

	cred, ok := cfg.CredentialByName(credential)
	require.True(t, ok)
	assert.Equal(t, CredentialTypeServicePrincipal, cred.Details.Type)

	ctx, ok := cfg.ContextByName(context)
	require.True(t, ok)
	assert.Equal(t, tenant, ctx.Details.Tenant)

	_, ok = cfg.TenantByName("missing")
	assert.False(t, ok)
	_, ok = cfg.CredentialByName("missing")
	assert.False(t, ok)
	_, ok = cfg.ContextByName("missing")
	assert.False(t, ok)
}

func TestUpsertsAndDeletes(t *testing.T) {
	const (
		tenant     = "corp"
		credential = "sp"
		context    = "dev"
	)

	cfg := newTestConfig(t).build()

	t1 := newTenant(t, tenant)
	cfg.UpsertTenant(t1)
	t2 := newTenant(t, tenant)
	cfg.UpsertTenant(t2)

	cred1 := newCredential(t, credential, CredentialTypeUser)
	cfg.UpsertCredential(&cred1)
	cred2 := newCredential(t, credential, CredentialTypeManagedIdentity)
	cfg.UpsertCredential(&cred2)

	ctx1 := newContext(t, context, tenant, credential)
	cfg.UpsertContext(ctx1)
	ctx2 := newContext(t, context, tenant, credential)
	ctx2.Details.Subscription = "sub-1"
	cfg.UpsertContext(ctx2)

	require.Len(t, cfg.Tenants, 1)
	assert.Equal(t, t2.Details.ID, cfg.Tenants[0].Details.ID)
	require.Len(t, cfg.Credentials, 1)
	assert.Equal(t, CredentialTypeManagedIdentity, cfg.Credentials[0].Details.Type)
	require.Len(t, cfg.Contexts, 1)
	assert.Equal(t, "sub-1", cfg.Contexts[0].Details.Subscription)

	assert.True(t, cfg.DeleteContext(ctx1.Name))
	assert.False(t, cfg.DeleteContext(ctx2.Name))
}

func TestDeleteTenant(t *testing.T) {
	cfg := newTestConfig(t).
		withTenant("corp").
		build()

	assert.True(t, cfg.DeleteTenant("corp"))
	assert.Empty(t, cfg.Tenants)
	assert.False(t, cfg.DeleteTenant("corp"))
}

func TestDeleteCredential(t *testing.T) {
	cfg := newTestConfig(t).
		withUserCredential("me").
		build()

	assert.True(t, cfg.DeleteCredential("me"))
	assert.Empty(t, cfg.Credentials)
	assert.False(t, cfg.DeleteCredential("me"))
}

func TestRenameContext(t *testing.T) {
	cfg := newTestConfig(t).
		withContext("old", "corp", "ci").
		build()

	assert.True(t, cfg.RenameContext("old", "new"))
	assert.Equal(t, "new", cfg.Contexts[0].Name)
	assert.False(t, cfg.RenameContext("old", "other"))
}

func TestRenameTenant(t *testing.T) {
	cfg := newTestConfig(t).
		withTenant("corp").
		build()

	assert.True(t, cfg.RenameTenant("corp", "corporate"))
	assert.Equal(t, "corporate", cfg.Tenants[0].Name)
	assert.False(t, cfg.RenameTenant("corp", "other"))
}

func TestRenameCredential(t *testing.T) {
	cfg := newTestConfig(t).
		withUserCredential("ci").
		build()

	assert.True(t, cfg.RenameCredential("ci", "ci-sp"))
	assert.Equal(t, "ci-sp", cfg.Credentials[0].Name)
	assert.False(t, cfg.RenameCredential("ci", "other"))
}

func TestRetargetTenant(t *testing.T) {
	cfg := newTestConfig(t).
		withTenant("corp").
		withUserCredential("ci").
		withContext("dev", "corp", "ci").
		withContext("prod", "corp", "ci").
		withContext("other", "platform", "ci").
		build()

	assert.Equal(t, 2, cfg.RetargetTenant("corp", "corporate"))
	assert.Equal(t, "corporate", cfg.Contexts[0].Details.Tenant)
	assert.Equal(t, "corporate", cfg.Contexts[1].Details.Tenant)
	assert.Equal(t, "platform", cfg.Contexts[2].Details.Tenant)
}

func TestRetargetCredential(t *testing.T) {
	cfg := newTestConfig(t).
		withUserCredential("ci").
		withContext("dev", "corp", "ci").
		withContext("prod", "corp", "other").
		build()

	assert.Equal(t, 1, cfg.RetargetCredential("ci", "ci-sp"))
	assert.Equal(t, "ci-sp", cfg.Contexts[0].Details.Credential)
	assert.Equal(t, "other", cfg.Contexts[1].Details.Credential)
}

func TestContextsReferencingTenant(t *testing.T) {
	cfg := newTestConfig(t).
		withContext("dev", "corp", "ci").
		withContext("prod", "corp", "ci").
		withContext("other", "platform", "ci").
		build()

	assert.Equal(t, []string{"dev", "prod"}, cfg.ContextsReferencingTenant("corp"))
	assert.Nil(t, cfg.ContextsReferencingTenant("missing"))
}

func TestContextsReferencingCredential(t *testing.T) {
	cfg := newTestConfig(t).
		withContext("dev", "corp", "ci").
		withContext("prod", "corp", "other").
		build()

	assert.Equal(t, []string{"dev"}, cfg.ContextsReferencingCredential("ci"))
	assert.Nil(t, cfg.ContextsReferencingCredential("missing"))
}

func TestMerge(t *testing.T) {
	base := newTestConfig(t).
		withTenant("corp").
		withUserCredential("ci").
		withContext("dev", "corp", "ci").
		withCurrentContext("dev").
		build()

	next := newTestConfig(t).
		withTenant("corp").
		withTenant("tenant-3").
		withMICredential("ci").
		withUserCredential("ops").
		withContext("dev", "corp", "ci", withSubscription("sub-a")).
		withContext("prod", "tenant-2", "ops").
		withCurrentContext("prod").
		build()

	err := base.Merge(&next)
	require.NoError(t, err)

	assert.Equal(t, "dev", base.CurrentContext)
	require.Len(t, base.Tenants, 2)
	require.Len(t, base.Credentials, 2)
	require.Len(t, base.Contexts, 2)

	tenant, ok := base.TenantByName("corp")
	require.True(t, ok)
	assert.Equal(t, base.Tenants[0].Details.ID, tenant.Details.ID)

	contextValue, ok := base.ContextByName("dev")
	require.True(t, ok)
	assert.Empty(t, contextValue.Details.Subscription)
}

func TestValidateContextReferences(t *testing.T) {
	cfg := newTestConfig(t).
		withTenant("corp").
		withUserCredential("ci").
		build()

	tests := []struct {
		name    string
		wantErr string
		context Context
	}{
		{
			name:    "valid",
			context: newContext(t, "dev", "corp", "ci"),
		},
		{
			name:    "missing name",
			context: newContext(t, "", "corp", "ci"),
			wantErr: "context name is required",
		},
		{
			name:    "missing tenant",
			context: newContext(t, "dev", "", "ci"),
			wantErr: "context tenant is required",
		},
		{
			name:    "missing credential",
			context: newContext(t, "dev", "corp", ""),
			wantErr: "context credential is required",
		},
		{
			name:    "unknown tenant",
			context: newContext(t, "dev", "missing", "ci"),
			wantErr: "tenant \"missing\" does not exist",
		},
		{
			name:    "unknown credential",
			context: newContext(t, "dev", "corp", "missing"),
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
