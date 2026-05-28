package tui

import (
	"cmp"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// viewerModel displays read-only details of a context.
type viewerModel struct {
	item *contextItem
}

// newViewer creates a detail viewer from a contextItem.
func newViewer(item *contextItem) viewerModel {
	return viewerModel{item: item}
}

func (m viewerModel) Init() tea.Cmd {
	return nil
}

func (m viewerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch keyName(msg.String()) {
		case keyQuit, keyCtrlC, keyEsc:
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m viewerModel) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Context: "+m.item.name) + "\n\n")

	const unset = "<unset>"
	rows := []struct{ label, value string }{
		{"Tenant", m.item.tenant},
		{"Tenant ID", cmp.Or(m.item.tenantID, unset)},
		{"Credential", m.item.credential},
		{"Credential Type", cmp.Or(m.item.credType, unset)},
		{"Subscription", cmp.Or(m.item.subscription, unset)},
		{"Current", fmt.Sprintf("%t", m.item.current)},
	}

	for _, row := range rows {
		label := dimStyle.Render(row.label + ":")
		fmt.Fprintf(&b, "  %s %s\n", label, row.value)
	}

	b.WriteString("\n" + helpStyle.Render("esc/q: back"))
	return b.String()
}
