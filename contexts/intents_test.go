package contexts

import (
	"testing"

	"github.com/lvlcn-t/azctx/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_CreateTenant(t *testing.T) {
	path := writeConfig(t, baseConfig())
	store := loadStore(t)

	require.NoError(t, New().CreateTenant(store, "extra", "tenant-9"))

	got, found := readConfig(t, path).TenantByName("extra")
	require.True(t, found)
	assert.Equal(t, "tenant-9", got.Details.ID)
}

func TestManager_CreateTenant_Errors(t *testing.T) {
	writeConfig(t, baseConfig())

	require.ErrorIs(t, New().CreateTenant(loadStore(t), "corp", "x"), ErrTenantExists)
	require.ErrorIs(t, New().CreateTenant(loadStore(t), "", "x"), ErrTenantNameRequired)
	require.ErrorIs(t, New().CreateTenant(loadStore(t), "new", ""), ErrTenantIDRequired)
}

func TestManager_UpdateTenant(t *testing.T) {
	path := writeConfig(t, baseConfig())
	store := loadStore(t)

	require.NoError(t, New().UpdateTenant(store, "corp", "tenant-updated"))

	got, _ := readConfig(t, path).TenantByName("corp")
	assert.Equal(t, "tenant-updated", got.Details.ID)
}

func TestManager_UpdateTenant_NotFound(t *testing.T) {
	writeConfig(t, baseConfig())

	require.ErrorIs(t, New().UpdateTenant(loadStore(t), "ghost", "x"), ErrTenantNotFound)
}

func TestManager_RenameTenant_CascadesReferences(t *testing.T) {
	path := writeConfig(t, baseConfig())
	store := loadStore(t)

	// dev references tenant corp; prod references platform.
	result, err := New().RenameTenant(store, "corp", "corporate")
	require.NoError(t, err)
	assert.Equal(t, []string{devContext}, result.UpdatedContexts)

	cfg := readConfig(t, path)
	_, found := cfg.TenantByName("corporate")
	assert.True(t, found)
	_, found = cfg.TenantByName("corp")
	assert.False(t, found)

	// The referencing context now points at the new name; the other is untouched.
	dev, _ := cfg.ContextByName(devContext)
	assert.Equal(t, "corporate", dev.Details.Tenant)
	prod, _ := cfg.ContextByName(prodContext)
	assert.Equal(t, "platform", prod.Details.Tenant)
}

func TestManager_RenameTenant_CrossFileCascade(t *testing.T) {
	// tenant lives in file 0, the referencing context in file 1.
	tenantsFile := &config.Config{
		Tenants: []config.Tenant{{Name: "corp", Details: config.TenantDetails{ID: "tenant-1"}}},
		Credentials: []config.Credential{
			{Name: "user", Details: config.CredentialDetails{Type: config.CredentialTypeUser}},
		},
	}
	contextsFile := &config.Config{
		Contexts: []config.Context{
			{Name: devContext, Details: config.ContextDetails{Tenant: "corp", Credential: "user"}},
		},
	}
	paths := writeConfigs(t, tenantsFile, contextsFile)
	store := loadStore(t)

	result, err := New().RenameTenant(store, "corp", "corporate")
	require.NoError(t, err)
	assert.Equal(t, []string{devContext}, result.UpdatedContexts)

	// The tenant renamed in its own file.
	_, found := readConfig(t, paths[0]).TenantByName("corporate")
	assert.True(t, found)

	// The context in the OTHER file was retargeted.
	dev, _ := readConfig(t, paths[1]).ContextByName(devContext)
	assert.Equal(t, "corporate", dev.Details.Tenant)
}

func TestManager_RenameTenant_Errors(t *testing.T) {
	writeConfig(t, baseConfig())

	_, err := New().RenameTenant(loadStore(t), "ghost", "x")
	require.ErrorIs(t, err, ErrTenantNotFound)

	_, err = New().RenameTenant(loadStore(t), "corp", "platform")
	require.ErrorIs(t, err, ErrTenantExists)
}

func TestManager_CreateCredential_Errors(t *testing.T) {
	writeConfig(t, baseConfig())

	existing := &config.Credential{Name: "user", Details: config.CredentialDetails{Type: config.CredentialTypeUser}}
	require.ErrorIs(t, New().CreateCredential(loadStore(t), existing), ErrCredentialExists)
}

func TestManager_UpdateCredential_NotFound(t *testing.T) {
	writeConfig(t, baseConfig())

	cred := &config.Credential{Name: "ghost", Details: config.CredentialDetails{Type: config.CredentialTypeUser}}
	require.ErrorIs(t, New().UpdateCredential(loadStore(t), cred), ErrCredentialNotFound)
}

func TestManager_RenameCredential_CascadesReferences(t *testing.T) {
	path := writeConfig(t, baseConfig())
	store := loadStore(t)

	// dev references credential user.
	result, err := New().RenameCredential(store, "user", "personal")
	require.NoError(t, err)
	assert.Equal(t, []string{devContext}, result.UpdatedContexts)

	cfg := readConfig(t, path)
	dev, _ := cfg.ContextByName(devContext)
	assert.Equal(t, "personal", dev.Details.Credential)
	_, found := cfg.CredentialByName("personal")
	assert.True(t, found)
}

func TestManager_CreateContext_Errors(t *testing.T) {
	writeConfig(t, baseConfig())

	dup := config.Context{Name: devContext, Details: config.ContextDetails{Tenant: "corp", Credential: "user"}}
	require.ErrorIs(t, New().CreateContext(loadStore(t), dup), ErrContextExists)
}

func TestManager_UpdateContext_NotFound(t *testing.T) {
	writeConfig(t, baseConfig())

	ctx := config.Context{Name: "ghost", Details: config.ContextDetails{Tenant: "corp", Credential: "user"}}
	require.ErrorIs(t, New().UpdateContext(loadStore(t), ctx, false), ErrContextNotFound)
}

func TestManager_DeleteTenant_ReportsOrphans(t *testing.T) {
	writeConfig(t, baseConfig())
	store := loadStore(t)

	// corp is referenced by dev.
	result, err := New().DeleteTenant(store, "corp")
	require.NoError(t, err)
	assert.Equal(t, []string{devContext}, result.OrphanedContexts)
}

func TestManager_DeleteCredential_ReportsOrphans(t *testing.T) {
	writeConfig(t, baseConfig())
	store := loadStore(t)

	// user is referenced by dev.
	result, err := New().DeleteCredential(store, "user")
	require.NoError(t, err)
	assert.Equal(t, []string{devContext}, result.OrphanedContexts)
}
