package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// itemDelegate is a custom list.ItemDelegate with Azure-inspired styling.
type itemDelegate struct {
	selectedStyle lipgloss.Style
	normalStyle   lipgloss.Style
	dimStyle      lipgloss.Style
}

func newItemDelegate() *itemDelegate {
	return &itemDelegate{
		selectedStyle: lipgloss.NewStyle().Foreground(colorPrimary).Bold(true),
		normalStyle:   lipgloss.NewStyle(),
		dimStyle:      lipgloss.NewStyle().Foreground(colorDim),
	}
}

func (d *itemDelegate) Height() int                         { return 2 }
func (d *itemDelegate) Spacing() int                        { return 1 }
func (d *itemDelegate) Update(tea.Msg, *list.Model) tea.Cmd { return nil }

//nolint:gocritic // Render signature is required by list.ItemDelegate interface.
func (d *itemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
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
