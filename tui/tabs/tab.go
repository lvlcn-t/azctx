package tabs

import (
	"reflect"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lvlcn-t/azctx/tui/details"
	"github.com/lvlcn-t/azctx/tui/keys"
)

// labelName is the shared display label for an entry's name field.
const labelName = "Name"

// Tab represents a single tab in the UI, responsible for rendering its content and handling interactions.
type Tab interface {
	Resize(width, height int)
	Update(msg tea.Msg) (TabAction, tea.Cmd)
	View() string
	Filtering() bool
	// Reload rebuilds the tab's items from the current store, e.g. after a write.
	Reload()
}

type tabKeys struct {
	Next   key.Binding
	Prev   key.Binding
	Select key.Binding
	View   key.Binding
	Create key.Binding
	Edit   key.Binding
	Rename key.Binding
	Delete key.Binding
	Close  key.Binding
	Quit   key.Binding
}

func newTabKeys(sel, view, close key.Binding) tabKeys { //nolint:gocritic // shadow is okay here
	return tabKeys{
		Next:   keys.New(keys.L).WithHelp("next").WithAliases(keys.Tab, keys.ArrowRight).Bind(),
		Prev:   keys.New(keys.H).WithHelp("prev").WithAliases(keys.ShiftTab, keys.ArrowLeft).Bind(),
		Select: sel,
		View:   view,
		Close:  close,
		Quit:   keys.New(keys.Quit).WithHelp("quit").WithAliases(keys.CtrlC).Bind(),
	}
}

func (k tabKeys) Help() func() []key.Binding { //nolint:gocritic // this must be a value receiver to avoid overwriting the original bindings
	// Exclude built-in keys from the help menu to avoid duplicates
	k.Quit = key.Binding{}

	var kys []key.Binding
	v := reflect.ValueOf(k)
	for _, field := range v.Fields() {
		k := field.Interface().(key.Binding)
		if !reflect.DeepEqual(k, key.Binding{}) {
			kys = append(kys, k)
		}
	}

	return func() []key.Binding { return kys }
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
	// TabActionCreate opens a form to create a new entry.
	TabActionCreate
	// TabActionEdit opens a form pre-filled from Item to edit it.
	TabActionEdit
	// TabActionRename opens a form to rename Item.
	TabActionRename
	// TabActionDelete requests confirmation to delete Item.
	TabActionDelete
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

// Create requests opening a create form for the active tab.
func Create() TabAction {
	return TabAction{Kind: TabActionCreate}
}

// Edit requests opening an edit form pre-filled from item.
func Edit(item details.Item) TabAction {
	return TabAction{Kind: TabActionEdit, Item: item}
}

// Delete requests confirmation to delete item.
func Delete(item details.Item) TabAction {
	return TabAction{Kind: TabActionDelete, Item: item}
}
