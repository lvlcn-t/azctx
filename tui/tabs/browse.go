package tabs

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lvlcn-t/azctx/tui/details"
	"github.com/lvlcn-t/azctx/tui/keys"
)

var _ Tab = (*browseTab)(nil)

type browseTab struct {
	list list.Model
	keys tabKeys
}

func newBrowseTab(l listBuilder) browseTab { //nolint:gocritic // irrelevant on startup
	tk := newTabKeys(
		keys.New(keys.Enter).WithHelp("view").WithAliases(keys.View, keys.Describe).Bind(),
		key.Binding{},
		keys.New(keys.Escape).WithHelp("close").Bind(),
	)
	return browseTab{
		list: l.
			ShowStatusBar(true).
			ShowHelp(true).
			EnableFiltering(true).
			WithShortHelp(tk.Help()).
			WithFullHelp(tk.Help()).
			Build(),
		keys: tk,
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

func (t *browseTab) Resize(width, height int) {
	t.list.SetSize(width, height)
}

func (t *browseTab) View() string {
	return t.list.View()
}
