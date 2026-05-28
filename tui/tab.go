package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var tabNames = []string{"Contexts", "Tenants", "Credentials"}

// renderTabs renders the tab bar with the given active index.
func renderTabs(activeIdx, width int) string {
	var rendered []string

	for i, name := range tabNames {
		if i == activeIdx {
			rendered = append(rendered, activeTabStyle.Render(name))
		} else {
			rendered = append(rendered, inactiveTabStyle.Render(name))
		}
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, rendered...)

	// Fill remaining width with a bottom border line.
	gap := width - lipgloss.Width(row)
	if gap > 0 {
		fill := lipgloss.NewStyle().
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(colorBorder).
			Render(strings.Repeat(" ", gap))
		row = lipgloss.JoinHorizontal(lipgloss.Bottom, row, fill)
	}

	return row
}
