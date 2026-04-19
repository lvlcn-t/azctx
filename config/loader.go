package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v4"
)

const (
	// ConfigEnvVar is the environment variable used for config path composition.
	ConfigEnvVar = "AZCTX"
	configDir    = ".azctx"
	configFile   = "config.yaml"
)

// Loaded is the fully loaded and indexed azctx config state.
type Loaded struct {
	Config    Config
	Paths     []string
	WritePath string

	fileConfigs map[string]Config
	sources     sourceIndex
}

// sourceIndex tracks the source file for each merged stanza.
type sourceIndex struct {
	CurrentContext string
	Tenants        map[string]string
	Credentials    map[string]string
	Contexts       map[string]string
}

// Load loads and merges azctx config files from AZCTX or default location.
func Load() (Loaded, error) {
	paths, err := resolvePaths(os.Getenv(ConfigEnvVar))
	if err != nil {
		return Loaded{}, err
	}

	loaded := Loaded{
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
		cfg, exists, readErr := readConfig(path)
		if readErr != nil {
			return Loaded{}, readErr
		}

		if !exists {
			continue
		}

		loaded.fileConfigs[path] = cfg
		loaded.indexSources(path, &cfg)
		loaded.Config.Merge(&cfg)
	}

	loaded.WritePath = loaded.defaultWritePath()

	return loaded, nil
}

// Read reads a single azctx config file without merge behavior.
func Read(path string) (Config, error) {
	cfg, _, err := readConfig(path)
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// FileConfig returns the parsed config for one source path.
func (loaded *Loaded) FileConfig(path string) Config {
	if loaded == nil {
		return Config{}
	}

	if cfg, exists := loaded.fileConfigs[path]; exists {
		return cfg
	}

	return Config{}
}

// PathForContext returns the source path responsible for a context entry.
func (loaded *Loaded) PathForContext(name string) string {
	if loaded == nil {
		return ""
	}

	if sourcePath, exists := loaded.sources.Contexts[name]; exists {
		return sourcePath
	}

	return loaded.defaultWritePath()
}

// PathForTenant returns the source path responsible for a tenant entry.
func (loaded *Loaded) PathForTenant(name string) string {
	if loaded == nil {
		return ""
	}

	if sourcePath, exists := loaded.sources.Tenants[name]; exists {
		return sourcePath
	}

	return loaded.defaultWritePath()
}

// PathForCredential returns the source path responsible for a credential entry.
func (loaded *Loaded) PathForCredential(name string) string {
	if loaded == nil {
		return ""
	}

	if sourcePath, exists := loaded.sources.Credentials[name]; exists {
		return sourcePath
	}

	return loaded.defaultWritePath()
}

// PathForCurrentContext returns the source path for current-context.
func (loaded *Loaded) PathForCurrentContext() string {
	if loaded == nil {
		return ""
	}

	if loaded.sources.CurrentContext != "" {
		return loaded.sources.CurrentContext
	}

	return loaded.defaultWritePath()
}

// defaultWritePath resolves the fallback path for newly created stanzas.
func (loaded *Loaded) defaultWritePath() string {
	if loaded == nil {
		return ""
	}

	for _, path := range loaded.Paths {
		if _, exists := loaded.fileConfigs[path]; exists {
			return path
		}
	}

	if len(loaded.Paths) > 0 {
		return loaded.Paths[len(loaded.Paths)-1]
	}

	return ""
}

// indexSources records where each stanza came from during merge.
func (loaded *Loaded) indexSources(path string, cfg *Config) {
	if loaded == nil || cfg == nil {
		return
	}

	if loaded.sources.CurrentContext == "" && cfg.CurrentContext != "" {
		loaded.sources.CurrentContext = path
	}

	for _, tenant := range cfg.Tenants {
		if _, exists := loaded.sources.Tenants[tenant.Name]; exists {
			continue
		}

		loaded.sources.Tenants[tenant.Name] = path
	}

	for _, credential := range cfg.Credentials {
		if _, exists := loaded.sources.Credentials[credential.Name]; exists {
			continue
		}

		loaded.sources.Credentials[credential.Name] = path
	}

	for _, context := range cfg.Contexts {
		if _, exists := loaded.sources.Contexts[context.Name]; exists {
			continue
		}

		loaded.sources.Contexts[context.Name] = path
	}
}

// resolvePaths resolves AZCTX into an ordered, de-duplicated path list.
func resolvePaths(rawEnv string) ([]string, error) {
	if strings.TrimSpace(rawEnv) == "" {
		defaultPath, err := DefaultPath()
		if err != nil {
			return nil, err
		}

		return []string{defaultPath}, nil
	}

	rawParts := strings.Split(rawEnv, string(os.PathListSeparator))
	paths := make([]string, 0, len(rawParts))
	known := make(map[string]struct{}, len(rawParts))

	for _, rawPart := range rawParts {
		part := strings.TrimSpace(rawPart)
		if part == "" {
			continue
		}

		expandedPath, err := expandPath(part)
		if err != nil {
			return nil, fmt.Errorf("invalid azctx path %q: %w", part, err)
		}

		if _, exists := known[expandedPath]; exists {
			continue
		}

		paths = append(paths, expandedPath)
		known[expandedPath] = struct{}{}
	}

	if len(paths) == 0 {
		defaultPath, err := DefaultPath()
		if err != nil {
			return nil, err
		}

		return []string{defaultPath}, nil
	}

	return paths, nil
}

// DefaultPath returns the default azctx config file path.
func DefaultPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("determine user home directory: %w", err)
	}

	return filepath.Join(homeDir, configDir, configFile), nil
}

// readConfig reads and decodes one config file.
func readConfig(path string) (Config, bool, error) {
	//nolint:gosec // Path comes from AZCTX or default config location by design.
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Config{}, false, nil
	}

	if err != nil {
		return Config{}, false, fmt.Errorf("read config %q: %w", path, err)
	}

	if len(bytes.TrimSpace(raw)) == 0 {
		return Config{}, true, nil
	}

	var cfg Config
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return Config{}, false, fmt.Errorf("parse config %q: %w", path, err)
	}

	return cfg, true, nil
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
