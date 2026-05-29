package state

import (
	"sync"

	"github.com/lvlcn-t/azctx/config"
)

type UI struct {
	mu sync.RWMutex

	context       string
	config        *config.Store
	mode          Mode
	state         State
	width, height int
	quitting      bool
}

func New(cfg *config.Store, mode Mode) *UI {
	return &UI{
		config: cfg,
		mode:   mode,
	}
}

func (u *UI) Width() int {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.width
}

func (u *UI) Height() int {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.height
}

func (u *UI) Resize(width, height int) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.width = width
	u.height = height
}

func (u *UI) Config() *config.Store {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.config
}

func (u *UI) Mode() Mode {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.mode
}

func (u *UI) Transition(s State) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.state = s
}

func (u *UI) Current() State {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.state
}

func (u *UI) QuitNow() {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.quitting = true
}

func (u *UI) Quitting() bool {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.quitting
}

func (u *UI) Context() string {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.context
}

func (u *UI) SelectContext(ctx string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.context = ctx
}
