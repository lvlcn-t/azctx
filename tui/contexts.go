package tui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lvlcn-t/azctx/config"
)

// contextsTab manages the contexts list.
type contextsTab struct {
	list list.Model
	mode Mode
}

func newContextsTab(cfg *config.Config, mode Mode, width, height int) contextsTab {
	items := buildContextItems(cfg)
	l := list.New(items, newItemDelegate(), width, height)
	l.Title = ""
	l.SetShowTitle(false)
	l.SetShowStatusBar(true)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	return contextsTab{list: l, mode: mode}
}

func (t *contextsTab) SetSize(w, h int) {
	t.list.SetWidth(w)
	t.list.SetHeight(h)
}

// Update handles input for the contexts tab.
// Returns a non-empty selected name if a context was chosen (ModeInteractive + Enter).
// Returns a non-nil viewerContent if the detail viewer should be shown.
func (t *contextsTab) Update(msg tea.KeyMsg) (string, viewerContent, tea.Cmd) {
	if t.list.FilterState() == list.Filtering {
		var cmd tea.Cmd
		t.list, cmd = t.list.Update(msg)
		return "", nil, cmd
	}

	switch keyName(msg.String()) {
	case keyEnter:
		item, ok := t.list.SelectedItem().(*contextItem)
		if !ok {
			return "", nil, nil
		}
		if t.mode == ModeInteractive {
			return item.name, nil, nil
		}
		return "", item, nil
	case keyView:
		if item, ok := t.list.SelectedItem().(*contextItem); ok {
			return "", item, nil
		}
		return "", nil, nil
	}

	var cmd tea.Cmd
	t.list, cmd = t.list.Update(msg)
	return "", nil, cmd
}

// UpdateMsg passes a generic message (e.g. WindowSizeMsg) to the list.
func (t *contextsTab) UpdateMsg(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	t.list, cmd = t.list.Update(msg)
	return cmd
}

func (t *contextsTab) View() string {
	return t.list.View()
}
