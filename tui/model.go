package tui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lvlcn-t/azctx/config"
)

type view int

const (
	viewList view = iota
	viewDetail
)

// model is the unified TUI model for all invocation modes.
type model struct {
	cfg      *config.Config
	mode     Mode
	view     view
	list     list.Model
	viewer   viewerModel
	choice   string
	width    int
	height   int
	quitting bool
}

func newModel(cfg *config.Config, mode Mode) model {
	items := buildItems(cfg)
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "azctx"
	l.SetShowStatusBar(true)
	l.SetShowHelp(true)
	l.SetFilteringEnabled(true)
	l.AdditionalShortHelpKeys = additionalKeys(mode)
	l.AdditionalFullHelpKeys = l.AdditionalShortHelpKeys
	return model{
		cfg:  cfg,
		mode: mode,
		view: viewList,
		list: l,
	}
}

//nolint:gocritic // Init must be a value receiver to satisfy tea.Model.
func (m model) Init() tea.Cmd {
	return nil
}

//nolint:gocritic // Update must be a value receiver to satisfy tea.Model.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height)
		return m, nil
	case tea.KeyMsg:
		switch m.view {
		case viewList:
			return m.updateList(msg)
		case viewDetail:
			return m.updateDetail(msg)
		}
	}

	if m.view == viewList {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}
	return m, nil
}

//nolint:gocritic // updateList must be a value receiver for bubbletea's Elm architecture.
func (m model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.list.FilterState() == list.Filtering {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	switch keyName(msg.String()) {
	case keyEnter:
		item, ok := m.list.SelectedItem().(*contextItem)
		if !ok {
			return m, nil
		}
		if m.mode == ModeInteractive {
			m.choice = item.name
			m.quitting = true
			return m, tea.Quit
		}
		// ModeBrowse: open detail view
		m.viewer = newViewer(item)
		m.view = viewDetail
		return m, nil
	case keyView:
		if item, ok := m.list.SelectedItem().(*contextItem); ok {
			m.viewer = newViewer(item)
			m.view = viewDetail
		}
		return m, nil
	case keyQuit, keyCtrlC:
		m.quitting = true
		return m, tea.Quit
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

//nolint:gocritic // updateDetail must be a value receiver for bubbletea's Elm architecture.
func (m model) updateDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch keyName(msg.String()) {
	case keyEsc, keyQuit:
		m.view = viewList
		return m, nil
	case keyCtrlC:
		m.quitting = true
		return m, tea.Quit
	}
	return m, nil
}

//nolint:gocritic // View must be a value receiver to satisfy tea.Model.
func (m model) View() string {
	if m.quitting {
		return ""
	}

	switch m.view {
	case viewDetail:
		return m.viewer.View()
	default:
		return m.list.View()
	}
}
