package control

import "github.com/charmbracelet/bubbles/key"

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

func (k Key) Binding(helpText string) key.Binding {
	return key.NewBinding(key.WithKeys(k.String()), key.WithHelp(k.String(), helpText))
}
