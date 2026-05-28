package tui

import "github.com/charmbracelet/lipgloss"

var (
	// titleStyle is the style for section titles.
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))

	// dimStyle is for secondary information.
	dimStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	// helpStyle is for the help bar at the bottom.
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)
