package config

import (
	"bytes"
	"cmp"
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

var (
	errEmptyConfig = errors.New("config is empty")
	errNotFound    = errors.New("config not found")
)

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
		Config:      Config{APIVersion: APIVersion, Kind: Kind},
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
		case errors.Is(err, errNotFound):
			continue
		case err != nil:
			return Store{}, err
		}

		loaded.fileConfigs[path] = cfg
		loaded.indexSources(path, &cfg)
		if err := loaded.Config.Merge(&cfg); err != nil {
			return Store{}, fmt.Errorf("merge config %q: %w", path, err)
		}
	}

	loaded.WritePath = loaded.defaultWritePath()

	return loaded, nil
}

// Read reads a single azctx config file without merge behavior.
func (l *Loader) Read(path string) (Config, error) {
	cfg, err := l.readConfig(path)
	switch {
	case errors.Is(err, errEmptyConfig) || errors.Is(err, errNotFound):
		return cfg, nil
	case err != nil:
		return Config{}, err
	}

	return cfg, nil
}

// DefaultPath returns the default config path for the current user.
func (l *Loader) DefaultPath() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, configDir, configFile), nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("determine home directory: %w", err)
	}

	return filepath.Join(homeDir, ".config", configDir, configFile), nil
}

func (l *Loader) readConfig(path string) (Config, error) {
	cfg := Config{APIVersion: APIVersion, Kind: Kind}
	raw, err := afero.ReadFile(l.fsys, path)
	if errors.Is(err, fs.ErrNotExist) {
		return cfg, errNotFound
	}

	if err != nil {
		return Config{}, fmt.Errorf("read config %q: %w", path, err)
	}

	if len(bytes.TrimSpace(raw)) == 0 {
		return cfg, errEmptyConfig
	}

	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config %q: %w", path, err)
	}

	cfg.APIVersion = cmp.Or(cfg.APIVersion, APIVersion)
	cfg.Kind = cmp.Or(cfg.Kind, Kind)

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
