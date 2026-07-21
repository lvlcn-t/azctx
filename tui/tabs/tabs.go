package tabs

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lvlcn-t/azctx/config"
	"github.com/lvlcn-t/azctx/contexts"
	"github.com/lvlcn-t/azctx/tui/details"
	"github.com/lvlcn-t/azctx/tui/form"
	"github.com/lvlcn-t/azctx/tui/keys"
	"github.com/lvlcn-t/azctx/tui/state"
	"github.com/lvlcn-t/azctx/tui/styles"
)

// Manager is the subset of contexts.Manager the tabs use to persist changes.
type Manager interface {
	CreateTenant(store *config.Store, name, id string) error
	UpdateTenant(store *config.Store, name, id string) error
	RenameTenant(store *config.Store, oldName, newName string) (contexts.RenameResult, error)
	DeleteTenant(store *config.Store, name string) (contexts.DeleteResult, error)

	CreateContext(store *config.Store, next config.Context) error
	UpdateContext(store *config.Store, next config.Context, subscriptionChanged bool) error
	RenameContext(store *config.Store, oldName, newName string) error
	DeleteContext(store *config.Store, name string) (contexts.DeleteResult, error)

	CreateCredential(store *config.Store, cred *config.Credential) error
	UpdateCredential(store *config.Store, cred *config.Credential) error
	RenameCredential(store *config.Store, oldName, newName string) (contexts.RenameResult, error)
	DeleteCredential(store *config.Store, name string) (contexts.DeleteResult, error)
}

// formIntent records what a submitted form should do.
type formIntent int

const (
	intentCreate formIntent = iota
	intentEdit
	intentRename
)

// Tabs is the main UI component that manages the different tabs and their content.
type Tabs struct {
	item    details.Item
	manager Manager
	pending details.Item
	state   *state.UI
	confirm confirm
	status  string
	keys    tabKeys
	details details.Viewer
	tabs    []Tab
	form    form.Model
	intent  formIntent
	active  int
}

// New creates a new Tabs component with the given state and manager.
func New(s *state.UI, manager Manager) *Tabs {
	t := &Tabs{
		state:   s,
		manager: manager,
		details: details.NewViewer(s),
		keys:    newTabKeys(key.Binding{}, key.Binding{}, key.Binding{}),
	}
	w, h := t.tabSize()
	l := newList(w, h)
	t.tabs = []Tab{
		contextsTab(s, l),
		tenantsTab(s, l),
		credentialsTab(s, l),
	}
	t.active = 0
	return t
}

// Resize resizes the active tab's content to fit the current terminal size.
func (t *Tabs) Resize() {
	w, h := t.tabSize()
	for _, tab := range t.tabs {
		tab.Resize(w, h)
	}
	// TODO: also resize details view if it's open?
}

// Tab layout content sizing constants.
const (
	verticalPadding   = 7  // title(1) + tab bar(3) + help(2) + spacing(1)
	horizontalPadding = 4  // left/right margin
	minContentHeight  = 5  // minimum usable list height
	minContentWidth   = 20 // minimum usable list width
)

func (t *Tabs) tabSize() (width, height int) {
	width = max(t.state.Width()-horizontalPadding, minContentWidth)
	height = max(t.state.Height()-verticalPadding, minContentHeight)
	return width, height
}

func (t *Tabs) Init() tea.Cmd {
	return nil
}

func (t *Tabs) Update(msg tea.Msg) tea.Cmd {
	switch {
	case t.state.Is(state.DetailView):
		var cmd tea.Cmd
		t.details, cmd = t.details.Update(msg)
		return cmd

	case t.state.Is(state.FormView):
		return t.updateForm(msg)

	case t.state.Is(state.ConfirmView):
		return t.updateConfirm(msg)
	}

	// If the active list is filtering, do not treat keys as global shortcuts.
	if t.tabs[t.active].Filtering() {
		action, cmd := t.tabs[t.active].Update(msg)
		return tea.Batch(cmd, t.handleAction(action))
	}

	switch {
	case keys.Matches(msg, t.keys.Next):
		t.next()
		return nil

	case keys.Matches(msg, t.keys.Prev):
		t.prev()
		return nil

	case keys.Matches(msg, t.keys.Quit):
		return t.state.Quit()
	}

	action, cmd := t.tabs[t.active].Update(msg)
	return tea.Batch(cmd, t.handleAction(action))
}

// updateForm drives the create/edit form and applies its result on submit.
func (t *Tabs) updateForm(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case form.Submitted:
		return t.applyForm(msg.Values)
	case form.Canceled:
		t.state.Transition(state.Tabs)
		return nil
	}

	var cmd tea.Cmd
	t.form, cmd = t.form.Update(msg)
	return cmd
}

// updateConfirm drives the delete confirmation prompt.
func (t *Tabs) updateConfirm(msg tea.Msg) tea.Cmd {
	if answer, ok := msg.(confirmed); ok {
		t.state.Transition(state.Tabs)
		if answer.ok {
			return t.applyDelete()
		}
		return nil
	}

	var cmd tea.Cmd
	t.confirm, cmd = t.confirm.Update(msg)
	return cmd
}

func (t *Tabs) View() string {
	switch {
	case t.state.Is(state.DetailView):
		return t.details.View(t.item)
	case t.state.Is(state.FormView):
		return t.form.View()
	case t.state.Is(state.ConfirmView):
		return t.confirm.View()
	}

	// Title
	cloud := styles.SplashCloudStyle.Render("☁")
	bolt := styles.SplashBoltStyle.Render("⚡")
	name := styles.SplashNameStyle.Render("azctx")
	title := " " + cloud + " " + bolt + " " + name

	// Tabs
	tabs := t.renderTabs()
	content := t.tabs[t.active].View()

	view := lipgloss.JoinVertical(lipgloss.Left, title, tabs, content)
	if t.status != "" {
		view = lipgloss.JoinVertical(lipgloss.Left, view, styles.HelpStyle.Render(t.status))
	}
	return view
}

func (t *Tabs) handleAction(action TabAction) tea.Cmd {
	switch action.Kind {
	case TabActionNone:
		return nil

	case TabActionShowDetails:
		if action.Item == nil {
			return nil
		}

		t.item = action.Item
		t.state.Transition(state.DetailView)
		return nil

	case TabActionSelect:
		if action.Item == nil {
			return nil
		}

		if t.state.Mode() == state.ModeInteractive {
			return t.handleInteractiveSelect(action.Item)
		}

		t.item = action.Item
		t.state.Transition(state.DetailView)
		return nil

	case TabActionCreate:
		return t.openForm(intentCreate, nil)

	case TabActionEdit:
		if action.Item == nil {
			return nil
		}
		return t.openForm(intentEdit, action.Item)

	case TabActionRename:
		if action.Item == nil {
			return nil
		}
		return t.openForm(intentRename, action.Item)

	case TabActionDelete:
		if action.Item == nil {
			return nil
		}
		return t.openDelete(action.Item)

	default:
		return nil
	}
}

// openForm opens a create, edit, or rename form for the active tab and records
// the intent and target item to apply on submit (nil item for create).
func (t *Tabs) openForm(intent formIntent, item details.Item) tea.Cmd {
	f, ok := t.buildForm(intent, item)
	if !ok {
		return nil
	}

	t.form = f
	t.intent = intent
	t.pending = item
	t.status = ""
	t.state.Transition(state.FormView)
	return nil
}

// openDelete opens the confirmation prompt for deleting item.
func (t *Tabs) openDelete(item details.Item) tea.Cmd {
	label, ok := deletableLabel(item)
	if !ok {
		return nil
	}

	t.confirm = newConfirm("Delete " + label + "?")
	t.pending = item
	t.status = ""
	t.state.Transition(state.ConfirmView)
	return nil
}

// buildForm returns the form for the active tab and intent, pre-filled from item
// for edit and rename. The second return is false when the active tab does not
// support the requested form.
func (t *Tabs) buildForm(intent formIntent, item details.Item) (form.Model, bool) {
	switch t.tabs[t.active].(type) {
	case *TenantsTab:
		if intent == intentRename {
			return renameForm("tenant", item), true
		}
		return tenantForm(intent, item), true

	case *ContextsTab:
		if intent == intentRename {
			return renameForm("context", item), true
		}
		return contextForm(intent, t.state.Config(), item), true

	case *CredentialsTab:
		if intent == intentRename {
			return renameForm("credential", item), true
		}
		return credentialForm(intent, item), true

	default:
		return form.Model{}, false
	}
}

// applyForm persists the submitted form values according to the active tab and
// recorded intent.
func (t *Tabs) applyForm(values map[string]string) tea.Cmd {
	t.state.Transition(state.Tabs)

	switch t.tabs[t.active].(type) {
	case *TenantsTab:
		return t.applyTenantForm(values)
	case *ContextsTab:
		return t.applyContextForm(values)
	case *CredentialsTab:
		return t.applyCredentialForm(values)
	default:
		return nil
	}
}

// applyCredentialForm maps a submitted credential form to the matching intent
// method.
func (t *Tabs) applyCredentialForm(values map[string]string) tea.Cmd {
	store := t.state.Config()
	cred := credentialFromValues(values)

	switch t.intent {
	case intentCreate:
		err := t.manager.CreateCredential(store, cred)
		return t.finish(err, "created credential "+cred.Name)

	case intentEdit:
		err := t.manager.UpdateCredential(store, cred)
		return t.finish(err, "updated credential "+cred.Name)

	case intentRename:
		item, ok := t.pending.(*CredentialItem)
		if !ok {
			return nil
		}
		result, err := t.manager.RenameCredential(store, item.Name, values[fieldName])
		return t.finish(err, renameStatus("credential", item.Name, values[fieldName], result.UpdatedContexts))

	default:
		return nil
	}
}

// applyContextForm maps a submitted context form to the matching intent method.
func (t *Tabs) applyContextForm(values map[string]string) tea.Cmd {
	store := t.state.Config()
	next := config.Context{
		Name: values[fieldName],
		Details: config.ContextDetails{
			Tenant:       values[fieldTenant],
			Credential:   values[fieldCredential],
			Subscription: values[fieldSubscription],
		},
	}

	switch t.intent {
	case intentCreate:
		err := t.manager.CreateContext(store, next)
		return t.finish(err, "created context "+next.Name)

	case intentEdit:
		// The form always carries the subscription value, so treat it as changed.
		err := t.manager.UpdateContext(store, next, true)
		return t.finish(err, "updated context "+next.Name)

	case intentRename:
		item, ok := t.pending.(*ContextItem)
		if !ok {
			return nil
		}
		err := t.manager.RenameContext(store, item.Name, values[fieldName])
		return t.finish(err, "renamed context "+item.Name+" to "+values[fieldName])

	default:
		return nil
	}
}

// applyTenantForm maps a submitted tenant form to the matching intent method.
func (t *Tabs) applyTenantForm(values map[string]string) tea.Cmd {
	store := t.state.Config()

	switch t.intent {
	case intentCreate:
		err := t.manager.CreateTenant(store, values[fieldName], values[fieldID])
		return t.finish(err, "created tenant "+values[fieldName])

	case intentEdit:
		err := t.manager.UpdateTenant(store, values[fieldName], values[fieldID])
		return t.finish(err, "updated tenant "+values[fieldName])

	case intentRename:
		item, ok := t.pending.(*TenantItem)
		if !ok {
			return nil
		}
		result, err := t.manager.RenameTenant(store, item.Name, values[fieldName])
		return t.finish(err, renameStatus("tenant", item.Name, values[fieldName], result.UpdatedContexts))

	default:
		return nil
	}
}

// applyDelete performs the pending delete.
func (t *Tabs) applyDelete() tea.Cmd {
	switch item := t.pending.(type) {
	case *TenantItem:
		result, err := t.manager.DeleteTenant(t.state.Config(), item.Name)
		return t.finish(err, deleteStatus("tenant", item.Name, result.OrphanedContexts))

	case *ContextItem:
		result, err := t.manager.DeleteContext(t.state.Config(), item.Name)
		status := "deleted context " + item.Name
		if result.WasActive {
			status += " (warning: removed the active context; use a context to select a new one)"
		}
		return t.finish(err, status)

	case *CredentialItem:
		result, err := t.manager.DeleteCredential(t.state.Config(), item.Name)
		return t.finish(err, deleteStatus("credential", item.Name, result.OrphanedContexts))

	default:
		return nil
	}
}

// finish reloads the tabs after a write and records a status message.
func (t *Tabs) finish(writeErr error, okStatus string) tea.Cmd {
	if writeErr != nil {
		t.status = "error: " + writeErr.Error()
		return nil
	}

	if err := t.reload(); err != nil {
		t.status = "error: " + err.Error()
		return nil
	}

	t.status = okStatus
	return nil
}

// reload re-reads the config from disk and rebuilds every tab's items.
func (t *Tabs) reload() error {
	loader := config.NewLoader()
	store, err := loader.Load()
	if err != nil {
		return err
	}

	t.state.SetConfig(&store)
	for _, tab := range t.tabs {
		tab.Reload()
	}
	return nil
}

// deletableLabel returns a human label for a deletable item.
func deletableLabel(item details.Item) (string, bool) {
	switch it := item.(type) {
	case *TenantItem:
		return "tenant " + it.Name, true
	case *ContextItem:
		return "context " + it.Name, true
	case *CredentialItem:
		return "credential " + it.Name, true
	default:
		return "", false
	}
}

// deleteStatus builds the status line for a delete, warning about orphans.
func deleteStatus(kind, name string, orphans []string) string {
	msg := "deleted " + kind + " " + name
	if len(orphans) > 0 {
		msg += " (warning: orphaned contexts: " + strings.Join(orphans, ", ") + ")"
	}
	return msg
}

// renameStatus builds the status line for a rename, noting cascaded contexts.
func renameStatus(kind, oldName, newName string, updated []string) string {
	msg := "renamed " + kind + " " + oldName + " to " + newName
	if len(updated) > 0 {
		msg += " (updated contexts: " + strings.Join(updated, ", ") + ")"
	}
	return msg
}

// renameForm builds a single-field form asking for the entry's new name.
func renameForm(kind string, item details.Item) form.Model {
	current := entryName(item)
	return form.New("Rename "+kind+" "+current, []form.Field{
		{Key: fieldName, Label: "New name", Placeholder: current, Required: true},
	})
}

// entryName returns the config name of a list item, without any display marker.
func entryName(item details.Item) string {
	switch it := item.(type) {
	case *TenantItem:
		return it.Name
	case *ContextItem:
		return it.Name
	case *CredentialItem:
		return it.Name
	default:
		return ""
	}
}

func (t *Tabs) handleInteractiveSelect(item details.Item) tea.Cmd {
	ctx, ok := item.(*ContextItem)
	if !ok {
		// In interactive mode, only context selection is meaningful right now.
		// Other tabs can either no-op or fall back to details, depending on taste.
		return nil
	}

	t.state.SelectContext(ctx.Name)
	return t.state.Quit()
}

func (t *Tabs) next() {
	t.active = (t.active + 1) % len(t.tabs)
}

func (t *Tabs) prev() {
	t.active = (t.active - 1 + len(t.tabs)) % len(t.tabs)
}

var tabLabels = []string{
	"Contexts",
	"Tenants",
	"Credentials",
}

// renderTabs renders the tab bar with the given active index.
func (t *Tabs) renderTabs() string {
	var rendered []string
	for i, label := range tabLabels {
		if i == t.active {
			rendered = append(rendered, styles.ActiveTabStyle.Render(label))
			continue
		}

		rendered = append(rendered, styles.InactiveTabStyle.Render(label))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, rendered...)

	// Fill remaining width with a bottom border line.
	gap := t.state.Width() - lipgloss.Width(row)
	if gap > 0 {
		fill := lipgloss.NewStyle().
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(styles.ColorBorder).
			Render(strings.Repeat(" ", gap))
		row = lipgloss.JoinHorizontal(lipgloss.Bottom, row, fill)
	}

	return row
}
