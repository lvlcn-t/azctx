package tabs

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lvlcn-t/azctx/config"
	"github.com/lvlcn-t/azctx/tui/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fieldTab moves focus forward count times.
func fieldTab(tabs *Tabs, count int) {
	for range count {
		tabs.Update(tea.KeyMsg{Type: tea.KeyTab})
	}
}

func TestTabs_CreateCredential_ServicePrincipal(t *testing.T) {
	path := writeConfig(t, baseConfig())
	tabs := newTabsOn(t, credentialsTabIndex)

	tabs.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	require.True(t, tabs.state.Is(state.FormView))

	// name (0)
	typeRunes(tabs, "ci-sp")
	fieldTab(tabs, 1) // -> type
	typeRunes(tabs, "service-principal")
	fieldTab(tabs, 1) // -> client-id
	typeRunes(tabs, "app-1")
	fieldTab(tabs, 1) // -> client-secret
	typeRunes(tabs, "shhh")
	cmd := tabs.Update(tea.KeyMsg{Type: tea.KeyEnter})
	drain(tabs, cmd)

	require.True(t, tabs.state.Is(state.Tabs), "status: %s", tabs.status)
	got, found := readConfig(t, path).CredentialByName("ci-sp")
	require.True(t, found)
	assert.Equal(t, config.CredentialTypeServicePrincipal, got.Details.Type)
	assert.Equal(t, "app-1", got.Details.Azure.ClientID)
	assert.Equal(t, "shhh", got.Details.Azure.ClientSecret)
}

func TestTabs_CreateCredential_WorkloadIdentityOAuth2(t *testing.T) {
	path := writeConfig(t, baseConfig())
	tabs := newTabsOn(t, credentialsTabIndex)

	tabs.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})

	typeRunes(tabs, "wi")
	fieldTab(tabs, 1) // type
	typeRunes(tabs, "workload-identity")
	fieldTab(tabs, 1) // client-id
	typeRunes(tabs, "wi-client")
	fieldTab(tabs, 3) // -> token-source (skip secret, cert)
	typeRunes(tabs, "oauth2")
	fieldTab(tabs, 2) // -> issuer (skip token-file)
	typeRunes(tabs, "https://issuer.example.com")
	fieldTab(tabs, 1) // -> oidc-client-id
	typeRunes(tabs, "oidc-client")
	fieldTab(tabs, 2) // -> scopes (skip redirect-uri)
	typeRunes(tabs, "openid,profile")
	cmd := tabs.Update(tea.KeyMsg{Type: tea.KeyEnter})
	drain(tabs, cmd)

	require.True(t, tabs.state.Is(state.Tabs), "status: %s", tabs.status)
	got, found := readConfig(t, path).CredentialByName("wi")
	require.True(t, found)
	assert.Equal(t, config.CredentialTypeWorkloadIdentity, got.Details.Type)
	require.NotNil(t, got.Details.Token.OAuth2)
	assert.Equal(t, "https://issuer.example.com", got.Details.Token.OAuth2.Issuer)
	assert.Equal(t, []string{"openid", "profile"}, got.Details.Token.OAuth2.Scopes)
}

func TestTabs_CreateCredential_RejectsInvalidType(t *testing.T) {
	writeConfig(t, baseConfig())
	tabs := newTabsOn(t, credentialsTabIndex)

	tabs.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	typeRunes(tabs, "bad")
	fieldTab(tabs, 1)
	typeRunes(tabs, "bogus")
	cmd := tabs.Update(tea.KeyMsg{Type: tea.KeyEnter})
	drain(tabs, cmd)

	// Inline validation keeps the form open.
	require.True(t, tabs.state.Is(state.FormView))
	assert.Contains(t, tabs.form.View(), "unsupported credential type")
}

func TestTabs_EditCredential_UpdatesInPlace(t *testing.T) {
	path := writeConfig(t, baseConfig())
	tabs := newTabsOn(t, credentialsTabIndex)

	// 'e' on the selected (only) credential 'user'.
	tabs.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	require.True(t, tabs.state.Is(state.FormView))
	require.Equal(t, "user", tabs.form.Values()[fieldName])

	// user -> managed-identity with a client id. Name is locked; focus starts
	// on type.
	tabs.Update(tea.KeyMsg{Type: tea.KeyCtrlU})
	typeRunes(tabs, "managed-identity")
	fieldTab(tabs, 1) // -> client-id
	typeRunes(tabs, "mi-client")
	cmd := tabs.Update(tea.KeyMsg{Type: tea.KeyEnter})
	drain(tabs, cmd)

	cfg := readConfig(t, path)
	require.Len(t, cfg.Credentials, 1)
	assert.Equal(t, "user", cfg.Credentials[0].Name)
	assert.Equal(t, config.CredentialTypeManagedIdentity, cfg.Credentials[0].Details.Type)
	assert.Equal(t, "mi-client", cfg.Credentials[0].Details.Azure.ClientID)
}

func TestTabs_RenameCredential(t *testing.T) {
	path := writeConfig(t, baseConfig())
	tabs := newTabsOn(t, credentialsTabIndex)

	tabs.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	require.True(t, tabs.state.Is(state.FormView))

	typeRunes(tabs, "personal")
	cmd := tabs.Update(tea.KeyMsg{Type: tea.KeyEnter})
	drain(tabs, cmd)

	cfg := readConfig(t, path)
	require.Len(t, cfg.Credentials, 1)
	assert.Equal(t, "personal", cfg.Credentials[0].Name)
	// The context referencing 'user' cascaded to 'personal'.
	dev, _ := cfg.ContextByName("dev")
	assert.Equal(t, "personal", dev.Details.Credential)
}

func TestTabs_DeleteCredential(t *testing.T) {
	path := writeConfig(t, baseConfig())
	tabs := newTabsOn(t, credentialsTabIndex)

	tabs.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	require.True(t, tabs.state.Is(state.ConfirmView))

	cmd := tabs.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	drain(tabs, cmd)

	_, found := readConfig(t, path).CredentialByName("user")
	assert.False(t, found)
}
