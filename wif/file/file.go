package file

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/lvlcn-t/azctx/config"
	"github.com/spf13/afero"
)

// Provider reads a federated identity token from a file.
type Provider struct {
	path string
	fsys afero.Fs
}

// ErrMissingFilePath indicates that the file token source config is missing a required file path.
var ErrMissingFilePath = errors.New("file token source requires a path")

// NewProvider creates a file-based token provider from the given config.
func NewProvider(cfg *config.FileSource) (*Provider, error) {
	if cfg == nil || cfg.Path == "" {
		return nil, ErrMissingFilePath
	}
	return &Provider{path: cfg.Path, fsys: afero.NewOsFs()}, nil
}

// AcquireToken reads and returns the token from the configured file path.
func (p *Provider) AcquireToken(_ context.Context) (string, error) {
	data, err := afero.ReadFile(p.fsys, p.path)
	if err != nil {
		return "", fmt.Errorf("read federated token file %q: %w", p.path, err)
	}

	token := strings.TrimSpace(string(data))
	if token == "" {
		return "", fmt.Errorf("federated token file %q is empty", p.path)
	}

	return token, nil
}
