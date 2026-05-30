package tabs

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lvlcn-t/azctx/tui/control"
	"github.com/lvlcn-t/azctx/tui/details"
	"github.com/lvlcn-t/azctx/tui/state"
	"github.com/lvlcn-t/azctx/tui/styles"
)

// Tabs is the main UI component that manages the different tabs and their content.
type Tabs struct {
	item    details.Item
	details details.Viewer
	state   *state.UI
	tabs    []Tab
	active  int
}

// New creates a new Tabs component with the given state.
func New(s *state.UI) *Tabs {
	t := &Tabs{
		state:   s,
		details: details.NewViewer(s),
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
	if t.state.Is(state.DetailView) {
		var cmd tea.Cmd
		t.details, cmd = t.details.Update(msg)
		return cmd
	}

	trigger := control.New(msg, t.state.Mode())
	// If the active list is filtering, do not treat keys as global shortcuts.
	if t.tabs[t.active].Filtering() {
		action, cmd := t.tabs[t.active].Update(trigger)
		return tea.Batch(cmd, t.handleAction(action))
	}

	switch trigger.Event {
	case control.EventNext:
		t.next()
		return nil

	case control.EventPrev:
		t.prev()
		return nil

	case control.EventQuit:
		return t.state.Quit()
	}

	action, cmd := t.tabs[t.active].Update(trigger)
	return tea.Batch(cmd, t.handleAction(action))
}

func (t *Tabs) View() string {
	if t.state.Is(state.DetailView) {
		return t.details.View(t.item)
	}

	// Title
	cloud := styles.SplashCloudStyle.Render("☁")
	bolt := styles.SplashBoltStyle.Render("⚡")
	name := styles.SplashNameStyle.Render("azctx")
	title := " " + cloud + " " + bolt + " " + name

	// Tabs
	tabs := t.renderTabs()
	content := t.tabs[t.active].View()

	// Help bar
	// helpBar := " " + m.helpModel.View(m.keyMap)
	helpBar := ""

	return lipgloss.JoinVertical(lipgloss.Left, title, tabs, content, helpBar)
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

	default:
		return nil
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
