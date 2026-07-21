package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lvlcn-t/azctx/config"
	"github.com/lvlcn-t/azctx/contexts"
	"github.com/lvlcn-t/azctx/tui/splash"
	"github.com/lvlcn-t/azctx/tui/state"
	"github.com/lvlcn-t/azctx/tui/tabs"
)

var _ tea.Model = (*App)(nil)

type App struct {
	// state holds the global state of the TUI, including configuration and UI dimensions.
	state *state.UI

	// tabs holds all the different tab models for the main app view.
	tabs *tabs.Tabs

	// splash is the initial view shown when the TUI starts, showing the
	// config loading status and welcome message.
	splash splash.Model
}

// NewApp builds the root model wired with a production contexts.Manager.
func NewApp(store *config.Store, mode state.Mode) *App {
	s := state.New(store, mode)
	return &App{
		state:  s,
		splash: splash.New(s),
		tabs:   tabs.New(s, contexts.New()),
	}
}

func (app *App) Init() tea.Cmd {
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
