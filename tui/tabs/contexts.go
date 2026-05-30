package tabs

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lvlcn-t/azctx/tui/keys"
	"github.com/lvlcn-t/azctx/tui/state"
)

var _ Tab = (*ContextsTab)(nil)

type ContextsTab struct {
	list  list.Model
	state *state.UI
	keys  tabKeys
}

func contextsTab(s *state.UI, l listBuilder) *ContextsTab { //nolint:gocritic // irrelevant on startup
	sel := keys.New(keys.Enter).WithHelp("select").WithAliases(keys.Use).Bind()
	view := keys.New(keys.View).WithHelp("view").WithAliases(keys.Describe).Bind()
	if s.Mode() == state.ModeBrowse {
		sel = keys.New(keys.Enter).WithHelp("view").WithAliases(keys.View, keys.Describe).Bind()
		view = key.Binding{}
	}

	tk := newTabKeys(sel, view, keys.New(keys.Escape).WithHelp("close").Bind())
	items := contextItems(s.Config())
	return &ContextsTab{
		list: l.WithItems(items...).
			ShowStatusBar(true).
			ShowHelp(true).
			EnableFiltering(true).
			WithShortHelp(tk.Help()).
			WithFullHelp(tk.Help()).
			Build(),
		state: s,
		keys:  tk,
	}
}

func (t *ContextsTab) Update(msg tea.Msg) (TabAction, tea.Cmd) {
	// bubbles/list needs to receive keys even after filtering has ended if a
	// filter value is still applied, otherwise esc cannot clear the active filter.
	if t.Filtering() || t.list.FilterValue() != "" {
		var cmd tea.Cmd
		t.list, cmd = t.list.Update(msg)
		return NoAction(), cmd
	}

	switch {
	case keys.Matches(msg, t.keys.Select):
		item, ok := t.list.SelectedItem().(*ContextItem)
		if !ok {
			return NoAction(), nil
		}
		return Select(item), nil

	case keys.Matches(msg, t.keys.View):
		item, ok := t.list.SelectedItem().(*ContextItem)
		if !ok {
			return NoAction(), nil
		}
		return ShowDetails(item), nil

	case keys.Matches(msg, t.keys.Close):
		// Catch close events to prevent the list from exiting when the user
		// spams the esc key while filtering.
		return NoAction(), nil
	}

	var cmd tea.Cmd
	t.list, cmd = t.list.Update(msg)
	return NoAction(), cmd
}

func (t *ContextsTab) Filtering() bool {
	return t.list.FilterState() == list.Filtering
}

func (t *ContextsTab) View() string {
	return t.list.View()
}

func (t *ContextsTab) Resize(width, height int) {
	t.list.SetSize(width, height)
}
