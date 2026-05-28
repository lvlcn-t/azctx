package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const minViewerWidth = 40

// viewerModel displays read-only details of an item in a bordered panel.
type viewerModel struct {
	content viewerContent
	width   int
}

func newViewer(content viewerContent, width int) viewerModel {
	return viewerModel{content: content, width: width}
}

func (m viewerModel) View() string {
	if m.content == nil {
		return ""
	}

	var b strings.Builder
	b.WriteString("\n")

	rows := m.content.detailRows()
	// Find max label width for alignment.
	maxLabel := 0
	for _, row := range rows {
		if len(row.label) > maxLabel {
			maxLabel = len(row.label)
		}
	}

	for _, row := range rows {
		label := dimStyle.Render(fmt.Sprintf("%-*s", maxLabel+1, row.label+":"))
		value := row.value
		if value == "" {
			value = dimStyle.Render("<unset>")
		}
		fmt.Fprintf(&b, "  %s  %s\n", label, value)
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  esc/q: back"))

	innerWidth := m.width - 6 // account for border + padding
	if innerWidth < minViewerWidth {
		innerWidth = minViewerWidth
	}

	title := titleStyle.Render(m.content.detailTitle())
	panel := viewerStyle.Width(innerWidth).Render(b.String())

	return lipgloss.JoinVertical(lipgloss.Left, title, panel)
}
