package tui

// keyName represents a keyboard input string used in bubbletea KeyMsg matching.
type keyName string

func (k keyName) String() string {
	return string(k)
}

const (
	keyEnter keyName = "enter"
	keyView  keyName = "v"

	keyEsc   keyName = "esc"
	keyQuit  keyName = "q"
	keyCtrlC keyName = "ctrl+c"

	keyTab      keyName = "tab"
	keyShiftTab keyName = "shift+tab"
	keyRight    keyName = "right"
	keyLeft     keyName = "left"
)
