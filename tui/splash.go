package tui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lvlcn-t/azctx/config"
)

const splashMinDuration = 500 * time.Millisecond

// Splash box size bounds.
const (
	splashMinBoxWidth  = 28
	splashMaxBoxWidth  = 50
	splashMinBoxHeight = 6
	splashMaxBoxHeight = 16
	splashBoxPadH      = 4 // horizontal padding inside box (border + Padding(0,2))
	splashBoxPadV      = 2 // vertical padding inside box (border top + bottom)
)

// configLoadedMsg is sent when the config has been loaded.
type configLoadedMsg struct {
	store config.Store
	err   error
}

// splashDoneMsg is sent when the minimum splash duration has elapsed.
type splashDoneMsg struct{}

// splashModel displays a loading screen while config is loading.
type splashModel struct {
	loader     config.Loader
	width      int
	height     int
	configDone bool
	timerDone  bool
	result     *configLoadedMsg
}

func newSplash(loader config.Loader) splashModel {
	return splashModel{loader: loader}
}

func (m splashModel) Init() tea.Cmd {
	return tea.Batch(
		m.loadConfig(),
		m.minDelay(),
	)
}

func (m splashModel) loadConfig() tea.Cmd {
	loader := m.loader
	return func() tea.Msg {
		store, err := loader.Load()
		return configLoadedMsg{store: store, err: err}
	}
}

func (m splashModel) minDelay() tea.Cmd {
	return tea.Tick(splashMinDuration, func(time.Time) tea.Msg {
		return splashDoneMsg{}
	})
}

func (m splashModel) ready() bool {
	return m.configDone && m.timerDone
}

func (m splashModel) View() string {
	boxWidth := min(splashMaxBoxWidth, max(splashMinBoxWidth, m.width/2))
	boxHeight := min(splashMaxBoxHeight, max(splashMinBoxHeight, m.height/3))

	// Logo (centered)
	cloud := splashCloudStyle.Render("☁")
	bolt := splashBoltStyle.Render("⚡")
	name := splashNameStyle.Render("azctx")
	logo := cloud + " " + bolt + "  " + name

	// Description (centered, dim)
	desc := splashDimStyle.Render("Switch Azure tenants & subscriptions fast")

	// Centered middle content
	innerWidth := boxWidth - splashBoxPadH
	innerHeight := boxHeight - splashBoxPadV

	top := lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).
		Render(lipgloss.JoinVertical(lipgloss.Center, logo, "", desc))

	// Status indicator (bottom-right)
	statusRight := lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Right).
		Render(splashDimStyle.Render("Loading config…"))

	// Fill vertical space around the main content, keeping status at the bottom-right
	topLines := lipgloss.Height(top)
	statusLines := lipgloss.Height(statusRight)

	remaining := max(innerHeight-topLines-statusLines, 0)
	topGap := remaining / 2
	bottomGap := remaining - topGap

	inner := strings.Repeat("\n", topGap) +
		top +
		strings.Repeat("\n", bottomGap) +
		statusRight

	box := splashBoxStyle.Width(boxWidth).Render(inner)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}
