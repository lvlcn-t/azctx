package config

import "cmp"

// Store is the fully loaded and indexed azctx config state.
type Store struct {
	sources     sourceIndex
	fileConfigs map[string]Config
	Config      Config
	WritePath   string
	Paths       []string
}

// sourceIndex tracks the source file for each merged stanza.
type sourceIndex struct {
	Tenants        map[string]string
	Credentials    map[string]string
	Contexts       map[string]string
	CurrentContext string
}

// FileConfig returns the parsed config for one source path.
func (s *Store) FileConfig(path string) Config {
	if s == nil {
		return Config{
			APIVersion: APIVersion,
			Kind:       Kind,
		}
	}

	if cfg, exists := s.fileConfigs[path]; exists {
		cfg.APIVersion = cmp.Or(cfg.APIVersion, APIVersion)
		cfg.Kind = cmp.Or(cfg.Kind, Kind)
		return cfg
	}

	return Config{
		APIVersion: APIVersion,
		Kind:       Kind,
	}
}

// PathForContext returns the source path responsible for a context entry.
func (s *Store) PathForContext(name string) string {
	if s == nil {
		return ""
	}

	if sourcePath, exists := s.sources.Contexts[name]; exists {
		return sourcePath
	}

	return s.defaultWritePath()
}

// PathForTenant returns the source path responsible for a tenant entry.
func (s *Store) PathForTenant(name string) string {
	if s == nil {
		return ""
	}

	if sourcePath, exists := s.sources.Tenants[name]; exists {
		return sourcePath
	}

	return s.defaultWritePath()
}

// PathForCredential returns the source path responsible for a credential entry.
func (s *Store) PathForCredential(name string) string {
	if s == nil {
		return ""
	}

	if sourcePath, exists := s.sources.Credentials[name]; exists {
		return sourcePath
	}

	return s.defaultWritePath()
}

// PathForCurrentContext returns the source path for current-context.
func (s *Store) PathForCurrentContext() string {
	if s == nil {
		return ""
	}

	if s.sources.CurrentContext != "" {
		return s.sources.CurrentContext
	}

	return s.defaultWritePath()
}

// defaultWritePath resolves the fallback path for newly created stanzas.
func (s *Store) defaultWritePath() string {
	if s == nil {
		return ""
	}

	for _, path := range s.Paths {
		if _, exists := s.fileConfigs[path]; exists {
			return path
		}
	}

	if len(s.Paths) > 0 {
		return s.Paths[0]
	}

	return ""
}

// indexSources records where each stanza came from during merge.
func (s *Store) indexSources(path string, cfg *Config) {
	if s == nil || cfg == nil {
		return
	}

	if s.sources.CurrentContext == "" && cfg.CurrentContext != "" {
		s.sources.CurrentContext = path
	}

	for _, tenant := range cfg.Tenants {
		if _, exists := s.sources.Tenants[tenant.Name]; exists {
			continue
		}

		s.sources.Tenants[tenant.Name] = path
	}

	for _, credential := range cfg.Credentials {
		if _, exists := s.sources.Credentials[credential.Name]; exists {
			continue
		}

		s.sources.Credentials[credential.Name] = path
	}

	for _, context := range cfg.Contexts {
		if _, exists := s.sources.Contexts[context.Name]; exists {
			continue
		}

		s.sources.Contexts[context.Name] = path
	}
}
