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

// formIntent records what a submitted form should do.
type formIntent int

const (
	intentCreate formIntent = iota
	intentEdit
	intentRename
)

// Tabs is the main UI component: a tab strip over uniform CRUD list tabs, plus
// the detail, form, and confirm overlays.
type Tabs struct {
	item    details.Item
	manager *contexts.Manager
	pending entry
	state   *state.UI
	confirm confirm
	status  string
	lastErr error
	keys    tabKeys
	details details.Viewer
	tabs    []*Tab
	form    form.Model
	intent  formIntent
	active  int
}

// New creates a Tabs component with the given state and manager.
func New(s *state.UI, manager *contexts.Manager) *Tabs {
	t := &Tabs{
		state:   s,
		manager: manager,
		details: details.NewViewer(s),
		keys:    newTabKeys(key.Binding{}, key.Binding{}, key.Binding{}),
	}
	w, h := t.tabSize()
	l := newList(w, h)

	view := keys.New(keys.View).WithHelp("view").WithAliases(keys.Describe).Bind()
	browseSel := keys.New(keys.Enter).WithHelp("view").WithAliases(keys.View, keys.Describe).Bind()

	// Contexts additionally support Enter=select (activate) in interactive mode.
	ctxSel := keys.New(keys.Enter).WithHelp("select").WithAliases(keys.Use).Bind()
	ctxView := view
	if s.Mode() == state.ModeBrowse {
		ctxSel = browseSel
		ctxView = key.Binding{}
	}

	t.tabs = []*Tab{
		newTab(s, "Contexts", contextItems, ctxSel, ctxView, l),
		newTab(s, "Tenants", tenantItems, browseSel, key.Binding{}, l),
		newTab(s, "Credentials", credentialItems, browseSel, key.Binding{}, l),
	}
	t.active = 0
	return t
}

// Resize resizes every tab's content to fit the current terminal size.
func (t *Tabs) Resize() {
	w, h := t.tabSize()
	for _, tab := range t.tabs {
		tab.Resize(w, h)
	}
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

func (t *Tabs) Init() tea.Cmd { return nil }

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

	// While filtering, keys belong to the list, not the global shortcuts.
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

// updateForm drives the create/edit/rename form and applies its result.
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

	cloud := styles.SplashCloudStyle.Render("☁")
	bolt := styles.SplashBoltStyle.Render("⚡")
	name := styles.SplashNameStyle.Render("azctx")
	title := " " + cloud + " " + bolt + " " + name

	content := t.tabs[t.active].View()
	view := lipgloss.JoinVertical(lipgloss.Left, title, t.renderTabs(), content)
	if t.status != "" {
		view = lipgloss.JoinVertical(lipgloss.Left, view, styles.HelpStyle.Render(t.status))
	}
	return view
}

// handleAction reacts to a tab's request. It is entity-agnostic: every branch
// operates on the entry the action carries (or the active tab's blank entry for
// create), never on a concrete item type.
func (t *Tabs) handleAction(action TabAction) tea.Cmd {
	switch action.kind {
	case showDetails:
		return t.showDetail(action.item)

	case selectAction:
		if t.state.Mode() == state.ModeInteractive {
			if a, ok := action.item.(activatable); ok {
				return a.activate(t.state)
			}
			return nil
		}
		return t.showDetail(action.item)

	case actionCreate:
		return t.openForm(intentCreate, t.tabs[t.active].blank())
	case editAction:
		return t.openForm(intentEdit, action.item)
	case renameAction:
		return t.openRename(action.item)
	case deleteAction:
		return t.openDelete(action.item)

	default:
		return nil
	}
}

func (t *Tabs) showDetail(item details.Item) tea.Cmd {
	if item == nil {
		return nil
	}
	t.item = item
	t.state.Transition(state.DetailView)
	return nil
}

// openForm opens a create or edit form for the given entry.
func (t *Tabs) openForm(intent formIntent, e entry) tea.Cmd {
	if e == nil {
		return nil
	}
	t.form = e.form(intent, t.state.Config())
	t.intent = intent
	t.pending = e
	t.status = ""
	t.state.Transition(state.FormView)
	return nil
}

// openRename opens the single-field rename form for the given entry.
func (t *Tabs) openRename(e entry) tea.Cmd {
	if e == nil {
		return nil
	}
	t.form = renameForm(e)
	t.intent = intentRename
	t.pending = e
	t.status = ""
	t.state.Transition(state.FormView)
	return nil
}

// openDelete opens the confirmation prompt for deleting the given entry.
func (t *Tabs) openDelete(e entry) tea.Cmd {
	if e == nil {
		return nil
	}
	t.confirm = newConfirm("Delete " + e.name() + "?")
	t.pending = e
	t.status = ""
	t.state.Transition(state.ConfirmView)
	return nil
}

// applyForm persists the submitted form via the pending entry.
func (t *Tabs) applyForm(values map[string]string) tea.Cmd {
	t.state.Transition(state.Tabs)
	if t.pending == nil {
		return nil
	}

	status, err := t.pending.save(t.manager, t.state.Config(), submission{intent: t.intent, values: values})
	return t.finish(status, err)
}

// applyDelete performs the pending delete via the pending entry.
func (t *Tabs) applyDelete() tea.Cmd {
	if t.pending == nil {
		return nil
	}
	status, err := t.pending.remove(t.manager, t.state.Config())
	return t.finish(status, err)
}

// finish reloads the tabs after a successful write and records a status line.
func (t *Tabs) finish(status string, writeErr error) tea.Cmd {
	t.lastErr = writeErr
	if writeErr != nil {
		t.status = "error: " + writeErr.Error()
		return nil
	}

	if err := t.reload(); err != nil {
		t.status = "error: " + err.Error()
		return nil
	}

	t.status = status
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

// renameForm builds the single-field new-name form for an entry.
func renameForm(e entry) form.Model {
	return form.New("Rename "+e.name(), []form.Field{
		{Key: fieldName, Label: "New name", Placeholder: e.name(), Required: true},
	})
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

func (t *Tabs) next() { t.active = (t.active + 1) % len(t.tabs) }
func (t *Tabs) prev() { t.active = (t.active - 1 + len(t.tabs)) % len(t.tabs) }

// renderTabs renders the tab bar with the active tab highlighted.
func (t *Tabs) renderTabs() string {
	rendered := make([]string, 0, len(t.tabs))
	for i, tab := range t.tabs {
		if i == t.active {
			rendered = append(rendered, styles.ActiveTabStyle.Render(tab.title))
			continue
		}
		rendered = append(rendered, styles.InactiveTabStyle.Render(tab.title))
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
