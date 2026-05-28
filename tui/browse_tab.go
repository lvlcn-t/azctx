package tui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// browseTab is a generic read-only tab with a filterable list.
// Used by tenants and credentials tabs which share identical behavior.
type browseTab struct {
	list list.Model
}

func newBrowseTab(items []list.Item, width, height int) browseTab {
	l := list.New(items, newItemDelegate(), width, height)
	l.Title = ""
	l.SetShowTitle(false)
	l.SetShowStatusBar(true)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	return browseTab{list: l}
}

func (t *browseTab) SetSize(w, h int) {
	t.list.SetWidth(w)
	t.list.SetHeight(h)
}

// Update handles input. Returns a non-nil viewerContent if detail should be shown.
func (t *browseTab) Update(msg tea.KeyMsg) (viewerContent, tea.Cmd) {
	if t.list.FilterState() == list.Filtering {
		var cmd tea.Cmd
		t.list, cmd = t.list.Update(msg)
		return nil, cmd
	}

	switch keyName(msg.String()) {
	case keyEnter, keyView:
		if item, ok := t.list.SelectedItem().(viewerContent); ok {
			return item, nil
		}
		return nil, nil
	}

	var cmd tea.Cmd
	t.list, cmd = t.list.Update(msg)
	return nil, cmd
}

// UpdateMsg passes a generic message to the list.
func (t *browseTab) UpdateMsg(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	t.list, cmd = t.list.Update(msg)
	return cmd
}

func (t *browseTab) View() string {
	return t.list.View()
}
