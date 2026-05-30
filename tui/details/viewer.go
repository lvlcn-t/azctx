package details

import (
	"cmp"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lvlcn-t/azctx/tui/keys"
	"github.com/lvlcn-t/azctx/tui/state"
	"github.com/lvlcn-t/azctx/tui/styles"
)

type Viewer struct {
	state *state.UI
	keys  viewerKeys
}

type viewerKeys struct {
	Close key.Binding
	Quit  key.Binding
}

func NewViewer(s *state.UI) Viewer {
	return Viewer{
		state: s,
		keys: viewerKeys{
			Close: keys.New(keys.Escape).WithHelp("close").WithAliases(keys.Quit).Bind(),
			Quit:  keys.New(keys.CtrlC).WithHelp("quit").Bind(),
		},
	}
}

func (d *Viewer) Update(msg tea.Msg) (Viewer, tea.Cmd) {
	switch {
	case keys.Matches(msg, d.keys.Close):
		d.state.Transition(state.Tabs)
		return *d, nil
	case keys.Matches(msg, d.keys.Quit):
		return *d, d.state.Quit()
	}
	return *d, nil
}

func (d *Viewer) View(item Item) string {
	if item == nil {
		return ""
	}

	var b strings.Builder
	b.WriteString("\n")

	rows := item.Details().Rows
	// Find max label width for alignment.
	maxLabel := 0
	for _, row := range rows {
		if len(row.Label) > maxLabel {
			maxLabel = len(row.Label)
		}
	}

	for _, row := range rows {
		label := styles.DimStyle.Render(fmt.Sprintf("%-*s", maxLabel+1, row.Label+":"))
		value := cmp.Or(row.Value, styles.DimStyle.Render("<unset>"))
		fmt.Fprintf(&b, "  %s  %s\n", label, value)
	}

	b.WriteString("\n")
	b.WriteString(styles.HelpStyle.Render("  esc/q: back"))

	const (
		// minWidth defines the minimum width for the detail viewer panel to ensure proper formatting.
		minWidth = 40
		// padding defines the total horizontal space taken by the panel's padding and borders.
		padding = 6
	)
	innerWidth := max(d.state.Width()-padding, minWidth)

	title := styles.TitleStyle.Render(item.Details().Title)
	panel := styles.ViewerStyle.Width(innerWidth).Render(b.String())
	return lipgloss.JoinVertical(lipgloss.Left, title, panel)
}
