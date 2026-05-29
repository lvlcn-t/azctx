package control

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lvlcn-t/azctx/tui/state"
)

// KeyOverride allows mapping specific keys to custom triggers
type KeyOverride func(Key) (Trigger, bool)

type Trigger struct {
	Msg   tea.Msg
	Event Event
}

// New creates a Trigger based on the given tea.Msg and current UI mode, applying any provided key overrides.
func New(msg tea.Msg, mode state.Mode, overrides ...KeyOverride) Trigger {
	m, ok := msg.(tea.KeyMsg)
	if !ok {
		return Trigger{Msg: msg}
	}

	k := Key(m.String())
	for _, o := range overrides {
		if o != nil {
			if trigger, ok := o(k); ok {
				return trigger
			}
		}
	}

	switch k {
	case KeyL, KeyTab, KeyRight:
		return Trigger{Event: EventNext, Msg: msg}
	case KeyH, KeyShiftTab, KeyLeft:
		return Trigger{Event: EventPrev, Msg: msg}
	case KeyEnter:
		if mode == state.ModeBrowse {
			return Trigger{Event: EventView, Msg: msg}
		}
		return Trigger{Event: EventSelect, Msg: msg}
	case KeyView, KeyDescribe:
		return Trigger{Event: EventView, Msg: msg}
	case KeyEscape:
		return Trigger{Event: EventClose, Msg: msg}
	case KeyQuit, KeyCtrlC:
		return Trigger{Event: EventQuit, Msg: msg}
	}
	return Trigger{Msg: msg}
}
