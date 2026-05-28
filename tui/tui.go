package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lvlcn-t/azctx/config"
)

// Mode determines the behavior of the TUI on context selection.
type Mode int

const (
	// ModeInteractive selects a context on Enter and quits (used by root and use commands).
	ModeInteractive Mode = iota
	// ModeBrowse opens the detail view on Enter without quitting (used by list command).
	ModeBrowse
)

// Run launches the TUI and returns the selected context name (empty if canceled or browse mode).
func Run(loader config.Loader, mode Mode) (string, error) {
	m := newModel(loader, mode)
	p := tea.NewProgram(m, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("run tui: %w", err)
	}

	final, ok := result.(model)
	if !ok {
		return "", nil
	}

	// If config loading failed, surface the error.
	if final.splash.result != nil && final.splash.result.err != nil {
		return "", final.splash.result.err
	}

	return final.choice, nil
}
