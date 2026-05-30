package tabs

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lvlcn-t/azctx/tui/control"
	"github.com/lvlcn-t/azctx/tui/state"
)

var _ Tab = (*ContextsTab)(nil)

type ContextsTab struct {
	list  list.Model
	state *state.UI
}

func contextsTab(s *state.UI, l listBuilder) *ContextsTab { //nolint:gocritic // irrelevant on startup
	items := contextItems(s.Config())
	return &ContextsTab{
		list: l.WithItems(items...).
			ShowStatusBar(true).
			ShowHelp(true).
			EnableFiltering(true).
			WithShortHelp(control.InteractiveTabHelp(s.Mode())).
			WithFullHelp(control.InteractiveTabHelp(s.Mode())).
			Build(),
		state: s,
	}
}

func (t *ContextsTab) Update(msg control.Trigger) (TabAction, tea.Cmd) {
	if t.Filtering() {
		var cmd tea.Cmd
		t.list, cmd = t.list.Update(msg.Msg)
		return NoAction(), cmd
	}

	switch msg.Event {
	case control.EventSelect:
		item, ok := t.list.SelectedItem().(*ContextItem)
		if !ok {
			return NoAction(), nil
		}
		return Select(item), nil

	case control.EventView:
		item, ok := t.list.SelectedItem().(*ContextItem)
		if !ok {
			return NoAction(), nil
		}
		return ShowDetails(item), nil
	}

	var cmd tea.Cmd
	t.list, cmd = t.list.Update(msg.Msg)
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
