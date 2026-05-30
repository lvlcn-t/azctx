package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lvlcn-t/azctx/config"
	"github.com/lvlcn-t/azctx/tui/splash"
	"github.com/lvlcn-t/azctx/tui/state"
	"github.com/lvlcn-t/azctx/tui/tabs"
)

var _ tea.Model = (*App)(nil)

type App struct {
	// state holds the global state of the TUI, including configuration and UI dimensions.
	state *state.UI

	// splash is the initial view shown when the TUI starts, showing the
	// config loading status and welcome message.
	splash splash.Model

	// tabs holds all the different tab models for the main app view.
	tabs *tabs.Tabs
}

func NewApp(store *config.Store, mode state.Mode) *App {
	s := state.New(store, mode)
	return &App{
		state:  s,
		splash: splash.New(s),
		tabs:   tabs.New(s),
	}
}

// Run launches the TUI and returns the selected context name (empty if canceled or browse mode).
func RunV2(store *config.Store, mode state.Mode) (string, error) {
	app := NewApp(store, mode)
	p := tea.NewProgram(app, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("run tui: %w", err)
	}

	final, ok := result.(*App)
	if !ok {
		return "", nil
	}

	// If config loading failed, surface the error.
	if final.splash.Err() != nil {
		return "", final.splash.Err()
	}

	return final.state.Context(), nil
}

func (app *App) Init() tea.Cmd {
	// TODO: add ctrl+c quit handler that works during splash screen as well.
	return app.splash.Init()
}

func (app *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		app.state.Resize(msg.Width, msg.Height)
		app.tabs.Resize()
		return app, nil
	case splash.Done:
		return app.enter(msg)
	}

	if app.state.Is(state.Splash) {
		var cmd tea.Cmd
		app.splash, cmd = app.splash.Update(msg)
		return app, cmd
	}

	return app, app.tabs.Update(msg)
}

func (app *App) View() string {
	if app.state.Is(state.Quitting) {
		return ""
	}

	if app.state.Is(state.Splash) {
		return app.splash.View()
	}

	return app.tabs.View()
}

func (app *App) enter(msg splash.Done) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		return app, app.state.Quit()
	}

	// TODO: create help keybinding map
	app.state.Transition(state.Tabs)
	return app, nil
}
