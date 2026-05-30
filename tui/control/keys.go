package control

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lvlcn-t/azctx/tui/state"
)

type Key string

// Vim movement keys
const (
	KeyH Key = "h"
	KeyJ Key = "j"
	KeyK Key = "k"
	KeyL Key = "l"
)

// Arrow keys
const (
	KeyUp    Key = "up"
	KeyDown  Key = "down"
	KeyLeft  Key = "left"
	KeyRight Key = "right"
)

// Tab keys
const (
	KeyTab      Key = "tab"
	KeyShiftTab Key = "shift+tab"
)

// Action keys
const (
	KeyUse      Key = "u"
	KeyView     Key = "v"
	KeyDescribe Key = "d"
)

// Control keys
const (
	KeyEnter  Key = "enter"
	KeyEscape Key = "esc"
	KeyQuit   Key = "q"
	KeyCtrlC  Key = "ctrl+c"
)

func (k Key) String() string {
	return string(k)
}

func (k Key) Matches(msg tea.Msg) bool {
	m, ok := msg.(fmt.Stringer)
	if !ok {
		return false
	}
	return key.Matches(m, k.Binding(""))
}

// Binding creates a [key.Binding] for the given Key, with optional aliases and help text.
func (k Key) Binding(helpText string, aliases ...Key) key.Binding {
	keys := []string{k.String()}
	for _, alias := range aliases {
		keys = append(keys, alias.String())
	}
	return key.NewBinding(key.WithKeys(keys...), key.WithHelp(k.String(), helpText))
}

// TabHelp returns the common key bindings for navigating between tabs.
func TabHelp() func() []key.Binding {
	return func() []key.Binding {
		return []key.Binding{
			KeyL.Binding("next", KeyTab, KeyRight),
			KeyH.Binding("prev", KeyShiftTab, KeyLeft),
		}
	}
}

// InteractiveTabHelp returns key bindings for interactive tabs, which include both navigation and selection/viewing actions.
func InteractiveTabHelp(m state.Mode) func() []key.Binding {
	return func() []key.Binding {
		if m == state.ModeBrowse {
			return BrowseTabHelp()()
		}

		return append(
			[]key.Binding{KeyEnter.Binding("select", KeyUse)},
			TabHelp()()...,
		)
	}
}

// BrowseTabHelp returns key bindings for browse mode tabs, which prioritize viewing.
func BrowseTabHelp() func() []key.Binding {
	return func() []key.Binding {
		return append(
			[]key.Binding{KeyEnter.Binding("view", KeyView, KeyDescribe)},
			TabHelp()()...,
		)
	}
}
