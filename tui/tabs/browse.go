package tabs

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lvlcn-t/azctx/tui/control"
	"github.com/lvlcn-t/azctx/tui/details"
)

var _ Tab = (*browseTab)(nil)

type browseTab struct {
	list list.Model
}

func newBrowseTab(l listBuilder) browseTab { //nolint:gocritic // irrelevant on startup
	return browseTab{
		list: l.
			ShowStatusBar(true).
			ShowHelp(true).
			EnableFiltering(true).
			WithShortHelp(control.BrowseTabHelp()).
			WithFullHelp(control.BrowseTabHelp()).
			Build(),
	}
}

func (t *browseTab) Update(msg control.Trigger) (TabAction, tea.Cmd) {
	if t.Filtering() {
		var cmd tea.Cmd
		t.list, cmd = t.list.Update(msg.Msg)
		return NoAction(), cmd
	}

	switch msg.Event {
	case control.EventSelect, control.EventView:
		if item, ok := t.list.SelectedItem().(details.Item); ok {
			return ShowDetails(item), nil
		}
		return NoAction(), nil

	case control.EventClose:
		// Catch close events to prevent the list from exiting when the user
		// spams the esc key while filtering.
		return NoAction(), nil
	}

	var cmd tea.Cmd
	t.list, cmd = t.list.Update(msg.Msg)
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
