package config

import (
	"cmp"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/spf13/afero"
	"go.yaml.in/yaml/v4"
)

type Writer struct {
	fsys afero.Fs
}

func NewWriter() Writer {
	return Writer{
		fsys: afero.NewOsFs(),
	}
}

// Write encodes and writes config data to a file.
func (w *Writer) Write(path string, cfg *Config) error {
	if cfg == nil {
		cfg = &Config{}
	}

	cfg.APIVersion = cmp.Or(cfg.APIVersion, apiVersion)
	cfg.Kind = cmp.Or(cfg.Kind, kind)

	const dirMode fs.FileMode = 0o700
	parent := filepath.Dir(path)
	if err := w.fsys.MkdirAll(parent, dirMode); err != nil {
		return fmt.Errorf("create config directory %q: %w", parent, err)
	}

	b, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	const fileMode fs.FileMode = 0o600
	if err := afero.WriteFile(w.fsys, path, b, fileMode); err != nil {
		return fmt.Errorf("write config %q: %w", path, err)
	}

	return nil
}
