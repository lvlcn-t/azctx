package tabs

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lvlcn-t/azctx/tui/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTabs_CreateContext(t *testing.T) {
	path := writeConfig(t, baseConfig())
	tabs := newTabsOn(t, contextsTabIndex)

	// 'n' opens the create form.
	tabs.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	require.True(t, tabs.state.Is(state.FormView))

	// name, tenant, credential, subscription.
	typeRunes(tabs, "prod")
	tabs.Update(tea.KeyMsg{Type: tea.KeyTab})
	typeRunes(tabs, "corp")
	tabs.Update(tea.KeyMsg{Type: tea.KeyTab})
	typeRunes(tabs, "user")
	tabs.Update(tea.KeyMsg{Type: tea.KeyTab})
	typeRunes(tabs, "sub-prod")
	cmd := tabs.Update(tea.KeyMsg{Type: tea.KeyEnter})
	drain(tabs, cmd)

	require.True(t, tabs.state.Is(state.Tabs))
	got, found := readConfig(t, path).ContextByName("prod")
	require.True(t, found)
	assert.Equal(t, "corp", got.Details.Tenant)
	assert.Equal(t, "user", got.Details.Credential)
	assert.Equal(t, "sub-prod", got.Details.Subscription)
}

func TestTabs_CreateContext_RejectsUnknownTenant(t *testing.T) {
	writeConfig(t, baseConfig())
	tabs := newTabsOn(t, contextsTabIndex)

	tabs.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	typeRunes(tabs, "prod")
	tabs.Update(tea.KeyMsg{Type: tea.KeyTab})
	typeRunes(tabs, "ghost") // not an existing tenant
	tabs.Update(tea.KeyMsg{Type: tea.KeyTab})
	typeRunes(tabs, "user")
	cmd := tabs.Update(tea.KeyMsg{Type: tea.KeyEnter})
	drain(tabs, cmd)

	// Validation blocks submission; the form stays open with an inline error.
	require.True(t, tabs.state.Is(state.FormView))
	assert.Contains(t, tabs.form.View(), "does not exist")
}

func TestTabs_EditContext_UpdatesInPlace(t *testing.T) {
	path := writeConfig(t, baseConfig())
	tabs := newTabsOn(t, contextsTabIndex)

	// 'e' opens the edit form for the selected (only) context 'dev'.
	tabs.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	require.True(t, tabs.state.Is(state.FormView))
	require.Equal(t, "dev", tabs.form.Values()["name"])

	// Name is locked; focus starts on tenant. Move to subscription and edit it.
	tabs.Update(tea.KeyMsg{Type: tea.KeyTab}) // tenant -> credential
	tabs.Update(tea.KeyMsg{Type: tea.KeyTab}) // credential -> subscription
	typeRunes(tabs, "sub-new")
	cmd := tabs.Update(tea.KeyMsg{Type: tea.KeyEnter})
	drain(tabs, cmd)

	cfg := readConfig(t, path)
	require.Len(t, cfg.Contexts, 1)
	assert.Equal(t, "dev", cfg.Contexts[0].Name)
	assert.Equal(t, "sub-new", cfg.Contexts[0].Details.Subscription)
}

func TestTabs_RenameContext(t *testing.T) {
	path := writeConfig(t, baseConfig())
	tabs := newTabsOn(t, contextsTabIndex)

	tabs.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	require.True(t, tabs.state.Is(state.FormView))

	typeRunes(tabs, "development")
	cmd := tabs.Update(tea.KeyMsg{Type: tea.KeyEnter})
	drain(tabs, cmd)

	cfg := readConfig(t, path)
	require.Len(t, cfg.Contexts, 1)
	assert.Equal(t, "development", cfg.Contexts[0].Name)
}

func TestTabs_DeleteContext(t *testing.T) {
	path := writeConfig(t, baseConfig())
	tabs := newTabsOn(t, contextsTabIndex)

	tabs.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	require.True(t, tabs.state.Is(state.ConfirmView))

	cmd := tabs.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	drain(tabs, cmd)

	_, found := readConfig(t, path).ContextByName("dev")
	assert.False(t, found)
}
