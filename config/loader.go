package config

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	"go.yaml.in/yaml/v4"
)

const (
	ConfigEnvVar = "AZCTX"
	configDir    = "azctx"
	configFile   = "config.yaml"
)

// Loader loads azctx config files.
type Loader struct {
	fsys afero.Fs
	env  string
}

var errEmptyConfig = errors.New("config is empty")

func NewLoader() Loader {
	return Loader{
		fsys: afero.NewOsFs(),
		env:  os.Getenv(ConfigEnvVar),
	}
}

// Load loads and merges azctx config files from AZCTX or default location.
func (l *Loader) Load() (Store, error) {
	paths, err := l.resolvePaths()
	if err != nil {
		return Store{}, err
	}

	loaded := Store{
		Config:      Config{},
		Paths:       paths,
		fileConfigs: make(map[string]Config, len(paths)),
		sources: sourceIndex{
			Tenants:     make(map[string]string),
			Credentials: make(map[string]string),
			Contexts:    make(map[string]string),
		},
	}

	for _, path := range paths {
		cfg, err := l.readConfig(path)
		switch {
		case errors.Is(err, errEmptyConfig):
			continue
		case err != nil:
			return Store{}, err
		}

		loaded.fileConfigs[path] = cfg
		loaded.indexSources(path, &cfg)
		loaded.Config.Merge(&cfg)
	}

	loaded.WritePath = loaded.defaultWritePath()

	return loaded, nil
}

// Read reads a single azctx config file without merge behavior.
func (l *Loader) Read(path string) (Config, error) {
	cfg, err := l.readConfig(path)
	if err != nil && !errors.Is(err, errEmptyConfig) {
		return Config{}, err
	}

	return cfg, nil
}

func (l *Loader) DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("determine user config directory: %w", err)
	}

	return filepath.Join(dir, configDir, configFile), nil
}

func (l *Loader) readConfig(path string) (Config, error) {
	raw, err := afero.ReadFile(l.fsys, path)
	if errors.Is(err, fs.ErrNotExist) {
		return Config{}, nil
	}

	if err != nil {
		return Config{}, fmt.Errorf("read config %q: %w", path, err)
	}

	if len(bytes.TrimSpace(raw)) == 0 {
		return Config{}, errEmptyConfig
	}

	var cfg Config
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config %q: %w", path, err)
	}

	return cfg, nil
}

func (l *Loader) resolvePaths() ([]string, error) {
	if strings.TrimSpace(l.env) == "" {
		p, err := l.DefaultPath()
		if err != nil {
			return nil, err
		}

		return []string{p}, nil
	}

	parts := strings.Split(l.env, string(os.PathListSeparator))
	paths := make([]string, 0, len(parts))
	known := make(map[string]struct{}, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		expanded, err := expandPath(part)
		if err != nil {
			return nil, fmt.Errorf("invalid azctx path %q: %w", part, err)
		}

		if _, exists := known[expanded]; exists {
			continue
		}

		paths = append(paths, expanded)
		known[expanded] = struct{}{}
	}

	if len(paths) == 0 {
		p, err := l.DefaultPath()
		if err != nil {
			return nil, err
		}

		return []string{p}, nil
	}

	return paths, nil
}

// expandPath expands supported home-path shorthand in config paths.
func expandPath(path string) (string, error) {
	if path == "~" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}

		return homeDir, nil
	}

	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}

		return filepath.Join(homeDir, path[2:]), nil
	}

	if strings.HasPrefix(path, "~") {
		return "", errors.New("only current-user (~) expansion is supported")
	}

	return filepath.Clean(path), nil
}
