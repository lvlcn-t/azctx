package config

import (
	"fmt"
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v4"
)

const (
	configDirectoryMode os.FileMode = 0o700
	configFileMode      os.FileMode = 0o600
)

// Write encodes and writes config data to a file.
func Write(path string, cfg *Config) error {
	if cfg == nil {
		cfg = &Config{}
	}

	if err := ensureParentDirectory(path); err != nil {
		return err
	}

	encoded, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, encoded, configFileMode); err != nil {
		return fmt.Errorf("write config %q: %w", path, err)
	}

	return nil
}

// ensureParentDirectory creates the destination directory when needed.
func ensureParentDirectory(path string) error {
	parentDirectory := filepath.Dir(path)
	if err := os.MkdirAll(parentDirectory, configDirectoryMode); err != nil {
		return fmt.Errorf("create config directory %q: %w", parentDirectory, err)
	}

	return nil
}
