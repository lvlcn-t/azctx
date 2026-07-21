package tabs

import (
	"reflect"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lvlcn-t/azctx/config"
	"github.com/lvlcn-t/azctx/tui/keys"
	"github.com/lvlcn-t/azctx/tui/state"
)

// labelName is the shared display label for an entry's name field.
const labelName = "Name"

// Form field keys shared across the entry forms.
const (
	fieldName         = "name"
	fieldID           = "id"
	fieldTenant       = "tenant"
	fieldCredential   = "credential"
	fieldSubscription = "subscription"

	fieldType         = "type"
	fieldClientID     = "client-id"
	fieldClientSecret = "client-secret"
	fieldCertPath     = "client-certificate-path"
	fieldTokenSource  = "token-source"
	fieldTokenFile    = "federated-token-file"
	fieldIssuer       = "issuer"
	fieldOIDCClientID = "oidc-client-id"
	fieldRedirectURI  = "redirect-uri"
	fieldScopes       = "scopes"
)

// Tab is a filterable list of one entry kind with create/edit/rename/delete
// keybindings. Every tab is the same type; the entry values it lists carry all
// entity-specific behavior.
type Tab struct {
	list    list.Model
	state   *state.UI
	rebuild func(*config.Store) []list.Item
	title   string
	keys    tabKeys
}

// newTab builds a tab for the given entry kind. rebuild produces the rows and
// sel/view are the Enter and view keybindings (which differ for contexts).
func newTab(s *state.UI, title string, rebuild func(*config.Store) []list.Item, sel, view key.Binding, l listBuilder) *Tab { //nolint:gocritic // listBuilder is a value builder, copied intentionally
	tk := newTabKeys(sel, view, keys.New(keys.Escape).WithHelp("close").Bind())
	tk.Create = keys.New(keys.Create).WithHelp("new").Bind()
	tk.Edit = keys.New(keys.Edit).WithHelp("edit").Bind()
	tk.Rename = keys.New(keys.Rename).WithHelp("rename").Bind()
	tk.Delete = keys.New(keys.Delete).WithHelp("delete").Bind()

	return &Tab{
		list: l.WithItems(rebuild(s.Config())...).
			ShowStatusBar(true).
			ShowHelp(true).
			EnableFiltering(true).
			WithShortHelp(tk.Help()).
			WithFullHelp(tk.Help()).
			Build(),
		keys:    tk,
		state:   s,
		rebuild: rebuild,
		title:   title,
	}
}

// Update handles list navigation and emits a TabAction for CRUD keys.
func (t *Tab) Update(msg tea.Msg) (TabAction, tea.Cmd) {
	// bubbles/list needs to receive keys even after filtering has ended if a
	// filter value is still applied, otherwise esc cannot clear the active filter.
	if t.Filtering() || t.list.FilterValue() != "" {
		var cmd tea.Cmd
		t.list, cmd = t.list.Update(msg)
		return noAction(), cmd
	}

	switch {
	case keys.Matches(msg, t.keys.Select):
		return t.actionOn(selectAction), nil
	case keys.Matches(msg, t.keys.View):
		return t.actionOn(showDetails), nil
	case keys.Matches(msg, t.keys.Create):
		return TabAction{kind: actionCreate}, nil
	case keys.Matches(msg, t.keys.Edit):
		return t.actionOn(editAction), nil
	case keys.Matches(msg, t.keys.Rename):
		return t.actionOn(renameAction), nil
	case keys.Matches(msg, t.keys.Delete):
		return t.actionOn(deleteAction), nil
	case keys.Matches(msg, t.keys.Close):
		// Swallow esc so the list does not exit while the user clears a filter.
		return noAction(), nil
	}

	var cmd tea.Cmd
	t.list, cmd = t.list.Update(msg)
	return noAction(), cmd
}

// actionOn builds an action of the given kind carrying the selected entry, or a
// no-op when the selection is not an entry.
func (t *Tab) actionOn(kind actionKind) TabAction {
	item, ok := t.list.SelectedItem().(entry)
	if !ok {
		return noAction()
	}
	return TabAction{kind: kind, item: item}
}

// blank returns a zero-value entry of this tab's kind, for building create
// forms and persisting new entries.
func (t *Tab) blank() entry {
	for _, item := range t.list.Items() {
		if e, ok := item.(entry); ok {
			return e.blank()
		}
	}
	// The list may be empty; rebuild from a throwaway to obtain the kind.
	for _, item := range t.rebuild(t.state.Config()) {
		if e, ok := item.(entry); ok {
			return e.blank()
		}
	}
	return nil
}

func (t *Tab) Filtering() bool { return t.list.FilterState() == list.Filtering }
func (t *Tab) View() string    { return t.list.View() }

func (t *Tab) Resize(width, height int) { t.list.SetSize(width, height) }

// Reload rebuilds the list items from the current store.
func (t *Tab) Reload() { t.list.SetItems(t.rebuild(t.state.Config())) }

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

func newTabKeys(sel, view, close key.Binding) tabKeys { //nolint:gocritic // shadowing the builtin close is fine for a local field name
	return tabKeys{
		Next:   keys.New(keys.L).WithHelp("next").WithAliases(keys.Tab, keys.ArrowRight).Bind(),
		Prev:   keys.New(keys.H).WithHelp("prev").WithAliases(keys.ShiftTab, keys.ArrowLeft).Bind(),
		Select: sel,
		View:   view,
		Close:  close,
		Quit:   keys.New(keys.Quit).WithHelp("quit").WithAliases(keys.CtrlC).Bind(),
	}
}

func (k tabKeys) Help() func() []key.Binding { //nolint:gocritic // value receiver avoids mutating the caller's bindings
	// Exclude built-in keys from the help menu to avoid duplicates.
	k.Quit = key.Binding{}

	var bindings []key.Binding
	v := reflect.ValueOf(k)
	for _, field := range v.Fields() {
		b := field.Interface().(key.Binding)
		if !reflect.DeepEqual(b, key.Binding{}) {
			bindings = append(bindings, b)
		}
	}

	return func() []key.Binding { return bindings }
}

// actionKind enumerates what a tab wants Tabs to do with the selected entry.
type actionKind int

const (
	actionNone actionKind = iota
	showDetails
	selectAction
	actionCreate
	editAction
	renameAction
	deleteAction
)

// TabAction is a tab's request to Tabs, optionally carrying the selected entry.
type TabAction struct {
	item entry
	kind actionKind
}

func noAction() TabAction { return TabAction{kind: actionNone} }
