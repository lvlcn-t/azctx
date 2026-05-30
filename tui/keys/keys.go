package keys

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Vim movement keys
var (
	H key.Binding = key.NewBinding(key.WithKeys("h"), key.WithHelp("h", "←"))
	J key.Binding = key.NewBinding(key.WithKeys("j"), key.WithHelp("j", "↓"))
	K key.Binding = key.NewBinding(key.WithKeys("k"), key.WithHelp("k", "↑"))
	L key.Binding = key.NewBinding(key.WithKeys("l"), key.WithHelp("l", "→"))
)

// Arrow keys
var (
	ArrowUp    key.Binding = key.NewBinding(key.WithKeys("up"), key.WithHelp("↑", "move up"))
	ArrowDown  key.Binding = key.NewBinding(key.WithKeys("down"), key.WithHelp("↓", "move down"))
	ArrowLeft  key.Binding = key.NewBinding(key.WithKeys("left"), key.WithHelp("←", "move left"))
	ArrowRight key.Binding = key.NewBinding(key.WithKeys("right"), key.WithHelp("→", "move right"))
)

// Tab keys
var (
	Tab      key.Binding = key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next"))
	ShiftTab key.Binding = key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "previous"))
)

// Action keys
var (
	Use      key.Binding = key.NewBinding(key.WithKeys("u"), key.WithHelp("u", "use"))
	View     key.Binding = key.NewBinding(key.WithKeys("v"), key.WithHelp("v", "view"))
	Describe key.Binding = key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "describe"))
)

// Control keys
var (
	Enter  key.Binding = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select"))
	Escape key.Binding = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "close"))
	Quit   key.Binding = key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit"))
	CtrlC  key.Binding = key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit"))
)

func Matches(msg tea.Msg, b ...key.Binding) bool {
	m, ok := msg.(fmt.Stringer)
	if !ok {
		return false
	}

	return key.Matches(m, b...)
}
