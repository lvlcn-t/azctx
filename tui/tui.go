package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lvlcn-t/azctx/config"
	"github.com/lvlcn-t/azctx/tui/state"
)

// Run launches the TUI and returns the selected context name (empty if canceled or browse mode).
func Run(store *config.Store, mode state.Mode) (string, error) {
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
