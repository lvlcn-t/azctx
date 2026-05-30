package details

import (
	"cmp"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lvlcn-t/azctx/tui/control"
	"github.com/lvlcn-t/azctx/tui/state"
	"github.com/lvlcn-t/azctx/tui/styles"
)

type Viewer struct {
	state *state.UI
}

func NewViewer(s *state.UI) Viewer {
	return Viewer{
		state: s,
	}
}

func (d *Viewer) Update(msg tea.Msg) (Viewer, tea.Cmd) {
	trigger := control.New(msg, d.state.Mode(), func(k control.Key) (control.Trigger, bool) {
		if k == control.KeyQuit {
			return control.Trigger{Event: control.EventClose, Msg: msg}, true
		}
		return control.Trigger{}, false
	})

	switch trigger.Event {
	case control.EventClose:
		d.state.Transition(state.Tabs)
		return *d, nil
	case control.EventQuit:
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
