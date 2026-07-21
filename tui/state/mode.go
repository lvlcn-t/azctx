package state

// Mode determines the behavior of the TUI on context selection.
type Mode int

const (
	// ModeInteractive selects a context on Enter and quits (used by root and use commands).
	ModeInteractive = iota
	// ModeBrowse opens the detail view on Enter without quitting (used by list command).
	ModeBrowse
)

type State int

const (
	// Splash is the initial loading screen while config is being read.
	Splash State = iota
	// Tabs is the main view showing contexts, tenants, and credentials.
	Tabs
	// DetailView is the overlay view showing details of a selected item.
	DetailView
	// FormView is the overlay view showing a create/edit form.
	FormView
	// ConfirmView is the overlay view prompting for delete confirmation.
	ConfirmView
	// Quitting is the state when the TUI is in the process of exiting.
	Quitting
)
