package styles

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var _ list.ItemDelegate = (*AzureRenderer)(nil)

// AzureRenderer is a custom list.ItemDelegate with Azure-inspired styling.
type AzureRenderer struct {
	selectedStyle lipgloss.Style
	normalStyle   lipgloss.Style
	dimStyle      lipgloss.Style
}

func NewAzureDelegate() *AzureRenderer {
	return &AzureRenderer{
		selectedStyle: lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true),
		normalStyle:   lipgloss.NewStyle(),
		dimStyle:      lipgloss.NewStyle().Foreground(ColorDim),
	}
}

func (d *AzureRenderer) Height() int                         { return 2 }
func (d *AzureRenderer) Spacing() int                        { return 1 }
func (d *AzureRenderer) Update(tea.Msg, *list.Model) tea.Cmd { return nil }

//nolint:gocritic // Render signature is required by list.ItemDelegate interface.
func (d *AzureRenderer) Render(w io.Writer, m list.Model, index int, item list.Item) {
	type titled interface {
		Title() string
		Description() string
	}

	i, ok := item.(titled)
	if !ok {
		return
	}

	title := i.Title()
	desc := i.Description()

	if m.Index() == index {
		title = d.selectedStyle.Render("> " + strings.TrimSpace(stripAnsi(title)))
		desc = d.dimStyle.Render("  " + desc)
	} else {
		title = "  " + title
		desc = d.dimStyle.Render("  " + desc)
	}

	fmt.Fprintf(w, "%s\n%s", title, desc)
}

// stripAnsi removes ANSI escape sequences for re-styling selected items.
func stripAnsi(s string) string {
	var b strings.Builder
	inEsc := false
	for _, r := range s {
		if r == '\x1b' {
			inEsc = true
			continue
		}
		if inEsc {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEsc = false
			}
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
