package tabs

import (
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lvlcn-t/azctx/config"
	"github.com/lvlcn-t/azctx/contexts"
	"github.com/lvlcn-t/azctx/tui/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// tenantsTabIndex is the position of the tenants tab in New's slice.
const tenantsTabIndex = 1

func writeConfig(t *testing.T, cfg *config.Config) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "azctx.yaml")
	writer := config.NewWriter()
	require.NoError(t, writer.Write(path, cfg))
	t.Setenv(config.ConfigEnvVar, path)

	return path
}

func readConfig(t *testing.T, path string) *config.Config {
	t.Helper()

	loader := config.NewLoader()
	cfg, err := loader.Read(path)
	require.NoError(t, err)

	return &cfg
}

func baseConfig() *config.Config {
	return &config.Config{
		Tenants: []config.Tenant{
			{Name: "corp", Details: config.TenantDetails{ID: "tenant-1"}},
		},
		Credentials: []config.Credential{
			{Name: "user", Details: config.CredentialDetails{Type: config.CredentialTypeUser}},
		},
		Contexts: []config.Context{
			{Name: "dev", Details: config.ContextDetails{Tenant: "corp", Credential: "user"}},
		},
	}
}

// newTenantTabs builds a Tabs on the tenants tab, backed by a real manager and
// the config at the AZCTX path. It sizes the tabs so the list is usable.
func newTenantTabs(t *testing.T) *Tabs {
	t.Helper()

	loader := config.NewLoader()
	store, err := loader.Load()
	require.NoError(t, err)

	s := state.New(&store, state.ModeBrowse)
	s.Resize(80, 24)
	s.Transition(state.Tabs)

	tabs := New(s, contexts.New())
	tabs.Resize()
	tabs.active = tenantsTabIndex

	return tabs
}

func typeRunes(tabs *Tabs, s string) {
	for _, r := range s {
		tabs.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
}

func TestTabs_CreateTenant(t *testing.T) {
	path := writeConfig(t, baseConfig())
	tabs := newTenantTabs(t)

	// 'n' opens the create form.
	tabs.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	require.True(t, tabs.state.Is(state.FormView))

	// Fill name, tab to id, fill id, submit.
	typeRunes(tabs, "dev")
	tabs.Update(tea.KeyMsg{Type: tea.KeyTab})
	typeRunes(tabs, "tenant-9")
	cmd := tabs.Update(tea.KeyMsg{Type: tea.KeyEnter})
	drain(tabs, cmd)

	require.True(t, tabs.state.Is(state.Tabs))
	got, found := readConfig(t, path).TenantByName("dev")
	require.True(t, found)
	assert.Equal(t, "tenant-9", got.Details.ID)
}

func TestTabs_EditTenant_UpdatesInPlace(t *testing.T) {
	path := writeConfig(t, baseConfig())
	tabs := newTenantTabs(t)

	// 'e' opens the edit form pre-filled from the selected (only) tenant.
	tabs.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	require.True(t, tabs.state.Is(state.FormView))
	require.Equal(t, "corp", tabs.form.Values()["name"])

	// The name field is locked; typing edits the id field, not the name.
	tabs.Update(tea.KeyMsg{Type: tea.KeyCtrlU})
	typeRunes(tabs, "tenant-updated")
	require.Equal(t, "corp", tabs.form.Values()["name"], "name must stay locked during edit")
	cmd := tabs.Update(tea.KeyMsg{Type: tea.KeyEnter})
	drain(tabs, cmd)

	cfg := readConfig(t, path)
	// Updated in place: exactly one tenant, with the new id, same name.
	require.Len(t, cfg.Tenants, 1)
	assert.Equal(t, "corp", cfg.Tenants[0].Name)
	assert.Equal(t, "tenant-updated", cfg.Tenants[0].Details.ID)
}

func TestTabs_RenameTenant_CascadesAndNoDuplicate(t *testing.T) {
	path := writeConfig(t, baseConfig())
	tabs := newTenantTabs(t)

	// 'r' opens the rename form.
	tabs.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	require.True(t, tabs.state.Is(state.FormView))

	typeRunes(tabs, "corporate")
	cmd := tabs.Update(tea.KeyMsg{Type: tea.KeyEnter})
	drain(tabs, cmd)

	cfg := readConfig(t, path)
	// Renamed, not duplicated: exactly one tenant under the new name.
	require.Len(t, cfg.Tenants, 1)
	assert.Equal(t, "corporate", cfg.Tenants[0].Name)

	// The referencing context was cascaded to the new name.
	dev, found := cfg.ContextByName("dev")
	require.True(t, found)
	assert.Equal(t, "corporate", dev.Details.Tenant)
}

func TestTabs_CreateTenant_RejectsDuplicate(t *testing.T) {
	writeConfig(t, baseConfig())
	tabs := newTenantTabs(t)

	tabs.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	typeRunes(tabs, "corp") // already exists
	tabs.Update(tea.KeyMsg{Type: tea.KeyTab})
	typeRunes(tabs, "tenant-2")
	cmd := tabs.Update(tea.KeyMsg{Type: tea.KeyEnter})
	drain(tabs, cmd)

	// The write is rejected and surfaced in the status line.
	assert.Contains(t, tabs.status, "already exists")
}

func TestTabs_DeleteTenant(t *testing.T) {
	tests := []struct {
		name      string
		answer    string
		wantFound bool
	}{
		{name: "confirmed", answer: "y", wantFound: false},
		{name: "canceled", answer: "n", wantFound: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeConfig(t, baseConfig())
			tabs := newTenantTabs(t)

			tabs.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
			require.True(t, tabs.state.Is(state.ConfirmView))

			cmd := tabs.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.answer)})
			drain(tabs, cmd)
			require.True(t, tabs.state.Is(state.Tabs))

			_, found := readConfig(t, path).TenantByName("corp")
			assert.Equal(t, tt.wantFound, found)
		})
	}
}

// drain runs the command returned by a submit/confirm and feeds the resulting
// message back into the model, mirroring the Bubble Tea runtime.
func drain(tabs *Tabs, cmd tea.Cmd) {
	if cmd == nil {
		return
	}
	if msg := cmd(); msg != nil {
		tabs.Update(msg)
	}
}

// compile-time guard: the real manager satisfies the tabs.Manager interface.
var _ Manager = (*contexts.Manager)(nil)
