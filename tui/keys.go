package tui

import "github.com/charmbracelet/bubbles/key"

// keyName represents a keyboard input string used in bubbletea KeyMsg matching.
type keyName string

func (k keyName) String() string {
	return string(k)
}

func (k keyName) Binding(help string) key.Binding {
	return key.NewBinding(key.WithKeys(k.String()), key.WithHelp(k.String(), help))
}

const (
	keyEnter keyName = "enter"
	keyView  keyName = "v"

	keyEsc   keyName = "esc"
	keyQuit  keyName = "q"
	keyCtrlC keyName = "ctrl+c"
)

func additionalKeys(mode Mode) func() []key.Binding {
	const (
		helpUse  = "use"
		helpView = "view"
	)

	enter := helpUse
	if mode == ModeBrowse {
		enter = helpView
	}

	return func() []key.Binding {
		return []key.Binding{
			keyEnter.Binding(enter),
			keyView.Binding(helpView),
		}
	}
}
