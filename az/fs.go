package az

import (
	"sync"
	"testing"

	"github.com/spf13/afero"
)

var (
	// fsys is the filesystem used for config operations.
	// Tests may override this with an in-memory filesystem
	// to avoid disk I/O.
	fsys afero.Fs = afero.NewOsFs()
	mu   sync.Mutex
)

// SetFS allows tests to override the filesystem used by az package operations.
func SetFS(t *testing.T, f afero.Fs) {
	t.Helper()

	mu.Lock()
	previous := fsys
	fsys = f
	mu.Unlock()

	t.Cleanup(func() {
		mu.Lock()
		fsys = previous
		mu.Unlock()
	})
}
