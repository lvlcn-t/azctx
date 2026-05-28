package tui

import "github.com/charmbracelet/lipgloss"

// Azure-inspired adaptive colors (Light = light terminal, Dark = dark terminal).
const (
	hexAzureBlue = "#0078D4"
	hexLightBlue = "#50E6FF"
	hexDarkBlue  = "#003B73"
	hexGreen     = "#107C10"
	hexGreenLt   = "#54B948"
	hexGrayDk    = "#8A8886"
	hexGrayLt    = "#A19F9D"
)

var (
	colorPrimary = lipgloss.AdaptiveColor{Light: hexAzureBlue, Dark: hexLightBlue}
	colorAccent  = lipgloss.AdaptiveColor{Light: hexDarkBlue, Dark: hexAzureBlue}
	colorSuccess = lipgloss.AdaptiveColor{Light: hexGreen, Dark: hexGreenLt}
	colorDim     = lipgloss.AdaptiveColor{Light: hexGrayDk, Dark: hexGrayLt}
	colorBorder  = lipgloss.AdaptiveColor{Light: hexAzureBlue, Dark: hexLightBlue}
	colorTitle   = lipgloss.AdaptiveColor{Light: hexDarkBlue, Dark: hexLightBlue}
)

// Title and text styles.
var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(colorTitle)
	dimStyle   = lipgloss.NewStyle().Foreground(colorDim)
	helpStyle  = lipgloss.NewStyle().Foreground(colorDim).MarginTop(1)
)

// Item marker styles.
var (
	currentMarkerStyle = lipgloss.NewStyle().Foreground(colorSuccess)
	normalMarkerStyle  = lipgloss.NewStyle().Foreground(colorDim)
)

// Tab styles.
var (
	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			Border(activeTabBorder(), true).
			BorderForeground(colorBorder).
			Padding(0, 1)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(colorDim).
				Border(inactiveTabBorder(), true).
				BorderForeground(colorBorder).
				Padding(0, 1)
)

// Panel styles.
var (
	viewerStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(1, 2)

	splashBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 2).
			Align(lipgloss.Center)
)

// Splash text styles.
var (
	splashCloudStyle = lipgloss.NewStyle().Foreground(colorPrimary)
	splashBoltStyle  = lipgloss.NewStyle().Foreground(colorAccent)
	splashNameStyle  = lipgloss.NewStyle().Bold(true).Foreground(colorTitle)
	splashDimStyle   = lipgloss.NewStyle().Foreground(colorDim)
)

func activeTabBorder() lipgloss.Border {
	b := lipgloss.RoundedBorder()
	b.BottomLeft = "┘"
	b.Bottom = " "
	b.BottomRight = "└"
	return b
}

func inactiveTabBorder() lipgloss.Border {
	b := lipgloss.RoundedBorder()
	b.BottomLeft = "┴"
	b.Bottom = "─"
	b.BottomRight = "┴"
	return b
}
