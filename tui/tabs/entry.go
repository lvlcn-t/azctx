package tabs

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lvlcn-t/azctx/config"
	"github.com/lvlcn-t/azctx/contexts"
	"github.com/lvlcn-t/azctx/tui/details"
	"github.com/lvlcn-t/azctx/tui/form"
	"github.com/lvlcn-t/azctx/tui/state"
)

// entry is a list row that maps to a config stanza and knows how to render and
// persist itself. The three item types (tenant, context, credential) implement
// it, so it is deliberately a multi-method interface shared by several
// implementations rather than a one- or two-method one.
type entry interface {
	list.Item    // Title, Description, FilterValue
	details.Item // Details

	// name returns the entry's config name, without any display marker.
	name() string
	// blank returns a fresh, empty entry of the same kind, used to build the
	// create form and to persist a newly created entry.
	blank() entry
	// form builds the create or edit form. On edit the name is locked.
	form(intent formIntent, store *config.Store) form.Model
	// save persists a create, edit, or rename described by sub, returning a
	// human-readable status line for display.
	save(m *contexts.Manager, store *config.Store, sub submission) (string, error)
	// remove deletes the entry, returning a human-readable status line.
	remove(m *contexts.Manager, store *config.Store) (string, error)
}

// activatable is the optional "use" capability. Only a context implements it:
// selecting it in interactive mode activates the context and quits.
type activatable interface {
	activate(s *state.UI) tea.Cmd
}

// submission is the payload of a form submit: the intent and the field values.
type submission struct {
	values map[string]string
	intent formIntent
}
