package contexts

import (
	"testing"

	"github.com/lvlcn-t/azctx/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_SetContext(t *testing.T) {
	tests := []struct {
		wantErr             error
		name                string
		wantTenant          string
		wantCredential      string
		wantSubscription    string
		next                config.Context
		subscriptionChanged bool
		wantErrOnly         bool
		wantExisted         bool
	}{
		{
			name: "create",
			next: config.Context{
				Name:    "stage",
				Details: config.ContextDetails{Tenant: "corp", Credential: "user", Subscription: "sub-stage"},
			},
			subscriptionChanged: true,
			wantExisted:         false,
			wantTenant:          "corp",
			wantCredential:      "user",
			wantSubscription:    "sub-stage",
		},
		{
			name: "update preserves unchanged subscription",
			next: config.Context{
				Name:    devContext,
				Details: config.ContextDetails{Tenant: "platform", Credential: "sp"},
			},
			subscriptionChanged: false,
			wantExisted:         true,
			wantTenant:          "platform",
			wantCredential:      "sp",
			wantSubscription:    "sub-dev",
		},
		{
			name: "update replaces subscription when changed",
			next: config.Context{
				Name:    devContext,
				Details: config.ContextDetails{Subscription: "sub-new"},
			},
			subscriptionChanged: true,
			wantExisted:         true,
			wantTenant:          "corp",
			wantCredential:      "user",
			wantSubscription:    "sub-new",
		},
		{
			name:    "empty name",
			next:    config.Context{},
			wantErr: ErrContextNameRequired,
		},
		{
			name: "missing tenant",
			next: config.Context{
				Name:    "bad",
				Details: config.ContextDetails{Tenant: "missing", Credential: "user"},
			},
			wantErrOnly: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeConfig(t, baseConfig())
			store := loadStore(t)

			existed, err := New().SetContext(store, tt.next, tt.subscriptionChanged)

			switch {
			case tt.wantErr != nil:
				require.ErrorIs(t, err, tt.wantErr)
				return
			case tt.wantErrOnly:
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantExisted, existed)

			got, found := readConfig(t, path).ContextByName(tt.next.Name)
			require.True(t, found)
			assert.Equal(t, tt.wantTenant, got.Details.Tenant)
			assert.Equal(t, tt.wantCredential, got.Details.Credential)
			assert.Equal(t, tt.wantSubscription, got.Details.Subscription)
		})
	}
}

func TestManager_SetTenant(t *testing.T) {
	tests := []struct {
		wantErr     error
		name        string
		tenant      string
		id          string
		wantExisted bool
	}{
		{name: "create", tenant: "extra", id: "tenant-9", wantExisted: false},
		{name: "update", tenant: "corp", id: "tenant-updated", wantExisted: true},
		{name: "empty name", tenant: "", id: "id", wantErr: ErrTenantNameRequired},
		{name: "empty id", tenant: "name", id: "", wantErr: ErrTenantIDRequired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeConfig(t, baseConfig())
			store := loadStore(t)

			existed, err := New().SetTenant(store, tt.tenant, tt.id)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantExisted, existed)

			got, found := readConfig(t, path).TenantByName(tt.tenant)
			require.True(t, found)
			assert.Equal(t, tt.id, got.Details.ID)
		})
	}
}

func TestManager_SetCredential(t *testing.T) {
	path := writeConfig(t, baseConfig())
	store := loadStore(t)

	cred := &config.Credential{
		Name: "ci-sp",
		Details: config.CredentialDetails{
			Type: config.CredentialTypeServicePrincipal,
			Azure: config.AzureCredential{
				ClientID:     "app-1",
				ClientSecret: "shhh",
			},
		},
	}

	existed, err := New().SetCredential(store, cred)
	require.NoError(t, err)
	assert.False(t, existed)

	got, found := readConfig(t, path).CredentialByName("ci-sp")
	require.True(t, found)
	assert.Equal(t, config.CredentialTypeServicePrincipal, got.Details.Type)
}

func TestManager_SetCredential_EmptyName(t *testing.T) {
	writeConfig(t, baseConfig())
	store := loadStore(t)

	_, err := New().SetCredential(store, &config.Credential{})
	require.ErrorIs(t, err, ErrCredentialNameRequired)
}

func TestManager_SetCredential_Invalid(t *testing.T) {
	writeConfig(t, baseConfig())
	store := loadStore(t)

	// Service principal without client credentials fails validation.
	cred := &config.Credential{
		Name:    "broken",
		Details: config.CredentialDetails{Type: config.CredentialTypeServicePrincipal},
	}

	_, err := New().SetCredential(store, cred)
	require.Error(t, err)
}

func TestManager_RenameContext(t *testing.T) {
	path := writeConfig(t, baseConfig())
	store := loadStore(t)

	require.NoError(t, New().RenameContext(store, prodContext, "production"))

	cfg := readConfig(t, path)
	_, found := cfg.ContextByName("production")
	assert.True(t, found)
	_, found = cfg.ContextByName(prodContext)
	assert.False(t, found)
}

func TestManager_RenameContext_UpdatesCurrent(t *testing.T) {
	path := writeConfig(t, baseConfig())
	store := loadStore(t)

	require.NoError(t, New().RenameContext(store, devContext, "development"))
	assert.Equal(t, "development", readConfig(t, path).CurrentContext)
}

func TestManager_RenameContext_Errors(t *testing.T) {
	writeConfig(t, baseConfig())

	require.ErrorIs(t, New().RenameContext(loadStore(t), "ghost", "new"), ErrContextNotFound)
	require.ErrorIs(t, New().RenameContext(loadStore(t), devContext, prodContext), ErrContextExists)
}

func TestManager_DeleteContext(t *testing.T) {
	path := writeConfig(t, baseConfig())
	store := loadStore(t)

	result, err := New().DeleteContext(store, prodContext)
	require.NoError(t, err)
	assert.Equal(t, path, result.Path)
	assert.False(t, result.WasActive)

	_, found := readConfig(t, path).ContextByName(prodContext)
	assert.False(t, found)
}

func TestManager_DeleteContext_Active(t *testing.T) {
	writeConfig(t, baseConfig())
	store := loadStore(t)

	result, err := New().DeleteContext(store, devContext)
	require.NoError(t, err)
	assert.True(t, result.WasActive)
}

func TestManager_DeleteContext_NotFound(t *testing.T) {
	writeConfig(t, baseConfig())
	store := loadStore(t)

	_, err := New().DeleteContext(store, "ghost")
	require.ErrorIs(t, err, ErrContextNotFound)
}

func TestManager_DeleteTenant(t *testing.T) {
	path := writeConfig(t, baseConfig())
	store := loadStore(t)

	result, err := New().DeleteTenant(store, "platform")
	require.NoError(t, err)
	assert.Equal(t, path, result.Path)
	assert.False(t, result.WasActive)

	_, found := readConfig(t, path).TenantByName("platform")
	assert.False(t, found)
}

func TestManager_DeleteCredential(t *testing.T) {
	path := writeConfig(t, baseConfig())
	store := loadStore(t)

	result, err := New().DeleteCredential(store, "sp")
	require.NoError(t, err)
	assert.Equal(t, path, result.Path)
	assert.False(t, result.WasActive)

	_, found := readConfig(t, path).CredentialByName("sp")
	assert.False(t, found)
}

func TestManager_Delete_NotFound(t *testing.T) {
	tests := []struct {
		wantErr error
		delete  func(m *Manager, store *config.Store) error
		name    string
	}{
		{
			name:    "tenant",
			delete:  func(m *Manager, s *config.Store) error { _, err := m.DeleteTenant(s, "ghost"); return err },
			wantErr: ErrTenantNotFound,
		},
		{
			name:    "credential",
			delete:  func(m *Manager, s *config.Store) error { _, err := m.DeleteCredential(s, "ghost"); return err },
			wantErr: ErrCredentialNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writeConfig(t, baseConfig())
			require.ErrorIs(t, tt.delete(New(), loadStore(t)), tt.wantErr)
		})
	}
}
