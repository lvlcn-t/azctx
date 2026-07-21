package tabs

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lvlcn-t/azctx/tui/keys"
	"github.com/lvlcn-t/azctx/tui/styles"
)

// confirmed is emitted when the user answers a confirmation prompt.
type confirmed struct {
	ok bool
}

// confirm is a small yes/no prompt shown before a destructive action.
type confirm struct {
	prompt string
}

// newConfirm builds a confirmation prompt with the given message.
func newConfirm(prompt string) confirm {
	return confirm{prompt: prompt}
}

// Update answers the prompt: y confirms, n or esc cancels.
func (c confirm) Update(msg tea.Msg) (confirm, tea.Cmd) {
	if _, ok := msg.(tea.KeyMsg); !ok {
		return c, nil
	}

	switch {
	case keys.Matches(msg, keys.Confirm):
		return c, answer(true)
	case keys.Matches(msg, keys.Cancel, keys.Escape):
		return c, answer(false)
	}

	return c, nil
}

// View renders the prompt as a bordered panel.
func (c confirm) View() string {
	body := lipgloss.JoinVertical(
		lipgloss.Left,
		styles.TitleStyle.Render(c.prompt),
		"",
		styles.HelpStyle.Render("y: confirm · n/esc: cancel"),
	)
	return styles.ViewerStyle.Render(body)
}

func answer(ok bool) tea.Cmd {
	return func() tea.Msg { return confirmed{ok: ok} }
}
