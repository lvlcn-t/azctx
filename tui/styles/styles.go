package styles

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/lipgloss"
)

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
	ColorPrimary = lipgloss.AdaptiveColor{Light: hexAzureBlue, Dark: hexLightBlue}
	ColorAccent  = lipgloss.AdaptiveColor{Light: hexDarkBlue, Dark: hexAzureBlue}
	ColorSuccess = lipgloss.AdaptiveColor{Light: hexGreen, Dark: hexGreenLt}
	ColorDim     = lipgloss.AdaptiveColor{Light: hexGrayDk, Dark: hexGrayLt}
	ColorBorder  = lipgloss.AdaptiveColor{Light: hexAzureBlue, Dark: hexLightBlue}
	ColorTitle   = lipgloss.AdaptiveColor{Light: hexDarkBlue, Dark: hexLightBlue}
)

// Title and text styles.
var (
	TitleStyle = lipgloss.NewStyle().Bold(true).Foreground(ColorTitle)
	DimStyle   = lipgloss.NewStyle().Foreground(ColorDim)
	HelpStyle  = lipgloss.NewStyle().Foreground(ColorDim).MarginTop(1)
)

func NewHelpStyles() help.Styles {
	base := help.New().Styles
	return help.Styles{
		Ellipsis: base.Ellipsis,

		ShortKey:       base.ShortKey.Foreground(ColorPrimary),
		ShortDesc:      base.ShortDesc.Foreground(ColorDim),
		ShortSeparator: base.ShortSeparator.Foreground(ColorDim),

		FullKey:       base.FullKey,
		FullDesc:      base.FullDesc,
		FullSeparator: base.FullSeparator,
	}
}

type ItemMarker = lipgloss.Style

// Item marker styles.
var (
	CurrentMarkerStyle ItemMarker = lipgloss.NewStyle().Foreground(ColorSuccess)
	NormalMarkerStyle  ItemMarker = lipgloss.NewStyle().Foreground(ColorDim)
)

// Tab styles.
var (
	ActiveTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Border(activeTabBorder(), true).
			BorderForeground(ColorBorder).
			Padding(0, 1)

	InactiveTabStyle = lipgloss.NewStyle().
				Foreground(ColorDim).
				Border(inactiveTabBorder(), true).
				BorderForeground(ColorBorder).
				Padding(0, 1)
)

// Panel styles.
var (
	ViewerStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(1, 2)

	SplashBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(0, 2).
			Align(lipgloss.Center)
)

// Splash text styles.
var (
	SplashCloudStyle = lipgloss.NewStyle().Foreground(ColorPrimary)
	SplashBoltStyle  = lipgloss.NewStyle().Foreground(ColorAccent)
	SplashNameStyle  = lipgloss.NewStyle().Bold(true).Foreground(ColorTitle)
	SplashDimStyle   = lipgloss.NewStyle().Foreground(ColorDim)
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
