package tabs

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lvlcn-t/azctx/tui/control"
	"github.com/lvlcn-t/azctx/tui/details"
)

// Tab represents a single tab in the UI, responsible for rendering its content and handling interactions.
type Tab interface {
	Resize(width, height int)
	Update(e control.Trigger) (TabAction, tea.Cmd)
	View() string
	Filtering() bool
}

type TabAction struct {
	Item details.Item
	Kind TabActionKind
}

type TabActionKind int

const (
	TabActionNone TabActionKind = iota
	TabActionShowDetails
	TabActionSelect
)

func NoAction() TabAction {
	return TabAction{Kind: TabActionNone}
}

func ShowDetails(item details.Item) TabAction {
	return TabAction{Kind: TabActionShowDetails, Item: item}
}

func Select(item details.Item) TabAction {
	return TabAction{Kind: TabActionSelect, Item: item}
}
