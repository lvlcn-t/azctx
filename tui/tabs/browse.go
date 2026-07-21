package tabs

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lvlcn-t/azctx/config"
	"github.com/lvlcn-t/azctx/tui/details"
	"github.com/lvlcn-t/azctx/tui/keys"
	"github.com/lvlcn-t/azctx/tui/state"
)

var _ Tab = (*browseTab)(nil)

type browseTab struct {
	list    list.Model
	state   *state.UI
	rebuild func(*config.Store) []list.Item
	keys    tabKeys
}

// newCRUDBrowseTab builds a browse tab that also binds create, edit, rename, and
// delete keys so the tab can emit CRUD actions.
func newCRUDBrowseTab(s *state.UI, rebuild func(*config.Store) []list.Item, l listBuilder) browseTab { //nolint:gocritic // irrelevant on startup
	tk := newTabKeys(
		keys.New(keys.Enter).WithHelp("view").WithAliases(keys.View, keys.Describe).Bind(),
		key.Binding{},
		keys.New(keys.Escape).WithHelp("close").Bind(),
	)
	tk.Create = keys.New(keys.Create).WithHelp("new").Bind()
	tk.Edit = keys.New(keys.Edit).WithHelp("edit").Bind()
	tk.Rename = keys.New(keys.Rename).WithHelp("rename").Bind()
	tk.Delete = keys.New(keys.Delete).WithHelp("delete").Bind()
	return buildBrowseTab(s, rebuild, l, tk)
}

func buildBrowseTab(s *state.UI, rebuild func(*config.Store) []list.Item, l listBuilder, tk tabKeys) browseTab { //nolint:gocritic // irrelevant on startup
	return browseTab{
		list: l.WithItems(rebuild(s.Config())...).
			ShowStatusBar(true).
			ShowHelp(true).
			EnableFiltering(true).
			WithShortHelp(tk.Help()).
			WithFullHelp(tk.Help()).
			Build(),
		keys:    tk,
		state:   s,
		rebuild: rebuild,
	}
}

func (t *browseTab) Update(msg tea.Msg) (TabAction, tea.Cmd) {
	// bubbles/list needs to receive keys even after filtering has ended if a
	// filter value is still applied, otherwise esc cannot clear the active filter.
	if t.Filtering() || t.list.FilterValue() != "" {
		var cmd tea.Cmd
		t.list, cmd = t.list.Update(msg)
		return NoAction(), cmd
	}

	switch {
	case keys.Matches(msg, t.keys.Select, t.keys.View):
		if item, ok := t.list.SelectedItem().(details.Item); ok {
			return ShowDetails(item), nil
		}
		return NoAction(), nil

	case keys.Matches(msg, t.keys.Create):
		return Create(), nil

	case keys.Matches(msg, t.keys.Edit):
		if item, ok := t.list.SelectedItem().(details.Item); ok {
			return Edit(item), nil
		}
		return NoAction(), nil

	case keys.Matches(msg, t.keys.Rename):
		if item, ok := t.list.SelectedItem().(details.Item); ok {
			return Rename(item), nil
		}
		return NoAction(), nil

	case keys.Matches(msg, t.keys.Delete):
		if item, ok := t.list.SelectedItem().(details.Item); ok {
			return Delete(item), nil
		}
		return NoAction(), nil

	case keys.Matches(msg, t.keys.Close):
		// Catch close events to prevent the list from exiting when the user
		// spams the esc key while filtering.
		return NoAction(), nil
	}

	var cmd tea.Cmd
	t.list, cmd = t.list.Update(msg)
	return NoAction(), cmd
}

func (t *browseTab) Filtering() bool {
	return t.list.FilterState() == list.Filtering
}

// Reload rebuilds the list items from the current store.
func (t *browseTab) Reload() {
	t.list.SetItems(t.rebuild(t.state.Config()))
}

func (t *browseTab) Resize(width, height int) {
	t.list.SetSize(width, height)
}

func (t *browseTab) View() string {
	return t.list.View()
}
