package tabs

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lvlcn-t/azctx/tui/control"
	"github.com/lvlcn-t/azctx/tui/details"
)

var _ Tab = (*browse)(nil)

type browse struct {
	list list.Model
}

func newBrowse(l listBuilder) browse { //nolint:gocritic // irrelevant on startup
	return browse{
		list: l.Build(),
	}
}

func (t *browse) Update(msg control.Trigger) (TabAction, tea.Cmd) {
	if t.Filtering() {
		return t.passthrough(msg)
	}

	switch msg.Event {
	case control.EventSelect, control.EventView:
		if item, ok := t.list.SelectedItem().(details.Item); ok {
			return ShowDetails(item), nil
		}
		return NoAction(), nil
	}

	return t.passthrough(msg)
}

func (t *browse) Filtering() bool {
	return t.list.FilterState() == list.Filtering
}

func (t *browse) Resize(width, height int) {
	t.list.SetSize(width, height)
}

func (t *browse) View() string {
	return t.list.View()
}

func (t *browse) passthrough(msg control.Trigger) (TabAction, tea.Cmd) {
	var cmd tea.Cmd
	t.list, cmd = t.list.Update(msg.Msg)
	return NoAction(), cmd
}
