package tui

import (
	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lvlcn-t/azctx/config"
)

type appState int

const (
	stateSplash appState = iota
	stateApp
)

const (
	tabContexts    = 0
	tabTenants     = 1
	tabCredentials = 2
	tabCount       = 3
)

// Layout constants for content sizing.
const (
	contentVerticalPadding   = 7  // title(1) + tab bar(3) + help(2) + spacing(1)
	contentHorizontalPadding = 4  // left/right margin
	minContentHeight         = 5  // minimum usable list height
	minContentWidth          = 20 // minimum usable list width
)

// model is the top-level TUI model managing splash, tabs, and viewer.
type model struct {
	state appState
	mode  Mode

	// Splash
	splash splashModel

	// App state (populated after config load)
	activeTab   int
	contexts    contextsTab
	tenants     browseTab
	credentials browseTab

	// Viewer overlay
	viewing bool
	viewer  viewerModel

	// Help bar
	helpModel help.Model
	keyMap    helpMap

	// Result
	choice string

	width, height int
	quitting      bool
}

func newModel(loader config.Loader, mode Mode) model {
	return model{
		state:  stateSplash,
		mode:   mode,
		splash: newSplash(loader),
	}
}

//nolint:gocritic // Init must be a value receiver to satisfy tea.Model.
func (m model) Init() tea.Cmd {
	return m.splash.Init()
}

//nolint:gocritic // Update must be a value receiver to satisfy tea.Model.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.splash.width = msg.Width
		m.splash.height = msg.Height
		if m.state == stateApp {
			m.resizeTabs()
		}
		return m, nil

	case configLoadedMsg:
		m.splash.configDone = true
		m.splash.result = &msg
		if m.splash.ready() {
			return m.transitionToApp()
		}
		return m, nil

	case splashDoneMsg:
		m.splash.timerDone = true
		if m.splash.ready() {
			return m.transitionToApp()
		}
		return m, nil
	}

	if m.state == stateSplash {
		return m, nil
	}

	// App state message routing.
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if m.viewing {
			return m.updateViewer(keyMsg)
		}
		return m.updateApp(keyMsg)
	}

	// Pass non-key messages to the active tab.
	cmd := m.activeTabUpdateMsg(msg)
	return m, cmd
}

//nolint:gocritic // transitionToApp must be a value receiver for bubbletea's Elm architecture.
func (m model) transitionToApp() (tea.Model, tea.Cmd) {
	if m.splash.result.err != nil {
		m.quitting = true
		return m, tea.Quit
	}

	store := &m.splash.result.store
	cfg := &m.splash.result.store.Config
	contentW, contentH := m.contentSize()

	m.contexts = newContextsTab(store, m.mode, contentW, contentH)
	m.tenants = newTenantsTab(cfg, contentW, contentH)
	m.credentials = newCredentialsTab(cfg, contentW, contentH)
	m.keyMap = newHelpMap(m.mode)
	m.helpModel = help.New()
	m.helpModel.ShortSeparator = " • "
	m.helpModel.Styles.ShortKey = lipgloss.NewStyle().Foreground(colorPrimary)
	m.helpModel.Styles.ShortDesc = lipgloss.NewStyle().Foreground(colorDim)
	m.helpModel.Styles.ShortSeparator = lipgloss.NewStyle().Foreground(colorDim)
	m.state = stateApp
	return m, nil
}

//nolint:gocritic // updateApp must be a value receiver for bubbletea's Elm architecture.
func (m model) updateApp(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch keyName(msg.String()) {
	case keyTab, keyRight, keyL:
		m.activeTab = (m.activeTab + 1) % tabCount
		return m, nil
	case keyShiftTab, keyLeft, keyH:
		m.activeTab = (m.activeTab - 1 + tabCount) % tabCount
		return m, nil
	case keyQuit, keyCtrlC:
		m.quitting = true
		return m, tea.Quit
	}

	switch m.activeTab {
	case tabContexts:
		selected, content, cmd := m.contexts.Update(msg)
		if selected != "" {
			m.choice = selected
			m.quitting = true
			return m, tea.Quit
		}
		if content != nil {
			m.viewing = true
			m.viewer = newViewer(content, m.width)
		}
		return m, cmd
	case tabTenants:
		content, cmd := m.tenants.Update(msg)
		if content != nil {
			m.viewing = true
			m.viewer = newViewer(content, m.width)
		}
		return m, cmd
	case tabCredentials:
		content, cmd := m.credentials.Update(msg)
		if content != nil {
			m.viewing = true
			m.viewer = newViewer(content, m.width)
		}
		return m, cmd
	}

	return m, nil
}

//nolint:gocritic // updateViewer must be a value receiver for bubbletea's Elm architecture.
func (m model) updateViewer(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch keyName(msg.String()) {
	case keyEsc, keyQuit:
		m.viewing = false
		return m, nil
	case keyCtrlC:
		m.quitting = true
		return m, tea.Quit
	}
	return m, nil
}

func (m *model) activeTabUpdateMsg(msg tea.Msg) tea.Cmd {
	switch m.activeTab {
	case tabContexts:
		return m.contexts.UpdateMsg(msg)
	case tabTenants:
		return m.tenants.UpdateMsg(msg)
	case tabCredentials:
		return m.credentials.UpdateMsg(msg)
	}
	return nil
}

func (m *model) resizeTabs() {
	w, h := m.contentSize()
	m.contexts.SetSize(w, h)
	m.tenants.SetSize(w, h)
	m.credentials.SetSize(w, h)
}

func (m *model) contentSize() (width, height int) {
	h := m.height - contentVerticalPadding
	if h < minContentHeight {
		h = minContentHeight
	}
	w := m.width - contentHorizontalPadding
	if w < minContentWidth {
		w = minContentWidth
	}
	return w, h
}

//nolint:gocritic // View must be a value receiver to satisfy tea.Model.
func (m model) View() string {
	if m.quitting {
		return ""
	}

	if m.state == stateSplash {
		return m.splash.View()
	}

	if m.viewing {
		return m.viewer.View()
	}

	// Title
	cloud := splashCloudStyle.Render("☁")
	bolt := splashBoltStyle.Render("⚡")
	name := splashNameStyle.Render("azctx")
	title := " " + cloud + " " + bolt + " " + name

	// Tabs
	tabs := renderTabs(m.activeTab, m.width)

	// Active tab content
	var content string
	switch m.activeTab {
	case tabContexts:
		content = m.contexts.View()
	case tabTenants:
		content = m.tenants.View()
	case tabCredentials:
		content = m.credentials.View()
	}

	// Help bar
	helpBar := " " + m.helpModel.View(m.keyMap)

	return lipgloss.JoinVertical(lipgloss.Left, title, tabs, content, helpBar)
}
