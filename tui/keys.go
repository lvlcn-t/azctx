package tui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
)

// keyName represents a keyboard input string used in bubbletea KeyMsg matching.
type keyName string

func (k keyName) String() string {
	return string(k)
}

// Binding creates a key.Binding from this key name with the given help text.
func (k keyName) Binding(text string) key.Binding {
	return key.NewBinding(key.WithKeys(k.String()), key.WithHelp(k.String(), text))
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

var _ help.KeyMap = (*helpMap)(nil)

// helpMap implements [help.KeyMap] for the main app view.
type helpMap struct {
	bindings []key.Binding
}

func (h helpMap) ShortHelp() []key.Binding  { return h.bindings }
func (h helpMap) FullHelp() [][]key.Binding { return [][]key.Binding{h.bindings} }

func newHelpMap(mode Mode) helpMap {
	enterDesc := "use"
	if mode == ModeBrowse {
		enterDesc = "view"
	}

	return helpMap{bindings: []key.Binding{
		keyEnter.Binding(enterDesc),
		keyView.Binding("detail"),
		keyTab.Binding("next tab"),
		key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
		keyQuit.Binding("quit"),
	}}
}
