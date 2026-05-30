package splash

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lvlcn-t/azctx/tui/keys"
	"github.com/lvlcn-t/azctx/tui/state"
	"github.com/lvlcn-t/azctx/tui/styles"
)

const minDuration = 500 * time.Millisecond

type Model struct {
	state  *state.UI
	result *Done
}

func New(s *state.UI) Model {
	return Model{state: s}
}

type Done struct {
	Err          error
	LoadedConfig bool
}

// Init starts the splash lifecycle and emits Done once the config load and minimum display time complete.
func (s *Model) Init() tea.Cmd {
	// TODO: batch a lazy load on s.state.Config with a minimum delay
	// to simulate loading time and show the splash screen.
	// On slow disks, the config load may be slow enough that the splash screen is visible without an artificial delay,
	// but on fast disks, it may load too quickly to see the splash screen at all.
	// By adding a minimum delay, we can ensure that the splash screen is visible for a short time even on fast disks.
	return tea.Batch(
		// ...,
		tea.Tick(minDuration, func(time.Time) tea.Msg {
			s.result = &Done{LoadedConfig: true}
			return Done{LoadedConfig: true}
		}),
	)
}

const (
	// minBoxW and maxBoxW define the minimum and maximum width of the splash box.
	minBoxW = 28
	maxBoxW = 50

	// minBoxH and maxBoxH define the minimum and maximum height of the splash box.
	minBoxH = 6
	maxBoxH = 16

	// boxPadHz is the horizontal padding inside box
	// (border + Padding(0,2))
	boxPadHz = 4
	// boxPadVt is the vertical padding inside box
	// (border top + bottom)
	boxPadVt = 2
)

// View renders the splash screen, which consists of a centered box with the app logo,
// a description, and a loading status at the bottom-right.
func (m *Model) View() string {
	boxWidth := min(maxBoxW, max(minBoxW, m.state.Width()/2))
	boxHeight := min(maxBoxH, max(minBoxH, m.state.Height()/3))

	// Logo (centered)
	cloud := styles.SplashCloudStyle.Render("☁")
	bolt := styles.SplashBoltStyle.Render("⚡")
	name := styles.SplashNameStyle.Render("azctx")
	logo := cloud + " " + bolt + "  " + name

	// Description (centered, dim)
	desc := styles.SplashDimStyle.Render("Switch Azure tenants & subscriptions fast")

	// Centered middle content
	innerWidth := boxWidth - boxPadHz
	innerHeight := boxHeight - boxPadVt

	top := lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).
		Render(lipgloss.JoinVertical(lipgloss.Center, logo, "", desc))

	// Status indicator (bottom-right)
	statusRight := lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Right).
		Render(styles.SplashDimStyle.Render("Loading config…"))

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

	box := styles.SplashBoxStyle.Width(boxWidth).Render(inner)

	return lipgloss.Place(m.state.Width(), m.state.Height(), lipgloss.Center, lipgloss.Center, box)
}

// Update updates any internal state of the splash screen.
func (s *Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if keys.Matches(msg, keys.CtrlC) {
		return *s, s.state.Quit()
	}
	return *s, nil
}

func (s *Model) Err() error {
	if s.result == nil {
		// This only happens if the user aborts the splash screen (e.g. ctrl+c)
		// before the minimum display time has elapsed.
		// In that case, we can just return nil since it's not an error.
		return nil
	}
	return s.result.Err
}
