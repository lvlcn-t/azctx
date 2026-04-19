package az

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	ini "gopkg.in/ini.v1"
)

const (
	azureConfigDirEnv                          = "AZURE_CONFIG_DIR"
	azureConfigDirectory                       = ".azure"
	azureConfigFile                            = "config"
	azureConfigDirectoryMode       fs.FileMode = 0o700
	azureConfigFileMode            fs.FileMode = 0o600
	azureConfigCoreSection                     = "core"
	azureConfigLoginExperienceKey              = "login_experience_v2"
	azureConfigLoginExperienceOff              = "off"
	azConfigRestoreWarningTemplate             = "warning: failed to restore Azure CLI config %q: %v\n"
)

// azConfigPath returns the Azure CLI config file path.
func azConfigPath() (string, error) {
	if configDir := strings.TrimSpace(os.Getenv(azureConfigDirEnv)); configDir != "" {
		return filepath.Join(configDir, azureConfigFile), nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}

	return filepath.Join(homeDir, azureConfigDirectory, azureConfigFile), nil
}

func withLoginExperienceOff(fn func() error) error {
	if fn == nil {
		return errors.New("login callback is required")
	}

	path, err := azConfigPath()
	if err != nil {
		return fmt.Errorf("resolve Azure CLI config path: %w", err)
	}

	configFile, err := loadAZConfig(path)
	if err != nil {
		return err
	}

	hadCoreSection := configFile.HasSection(azureConfigCoreSection)
	coreSection, err := ensureSection(configFile, azureConfigCoreSection)
	if err != nil {
		return fmt.Errorf("ensure section %q in Azure CLI config %q: %w", azureConfigCoreSection, path, err)
	}

	hadLoginExperienceKey := coreSection.HasKey(azureConfigLoginExperienceKey)
	originalValue := coreSection.Key(azureConfigLoginExperienceKey).String()
	coreSection.Key(azureConfigLoginExperienceKey).SetValue(azureConfigLoginExperienceOff)

	if err := writeAZConfig(path, configFile); err != nil {
		return fmt.Errorf("write Azure CLI config %q: %w", path, err)
	}

	defer func() {
		if restoreErr := restoreLoginExperience(configFile, path, hadCoreSection, hadLoginExperienceKey, originalValue); restoreErr != nil {
			fmt.Fprintf(os.Stderr, azConfigRestoreWarningTemplate, path, restoreErr)
		}
	}()

	return fn()
}

func loadAZConfig(path string) (*ini.File, error) {
	raw, err := afero.ReadFile(fsys, path)
	if errors.Is(err, fs.ErrNotExist) {
		return ini.Empty(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("load Azure CLI config %q: %w", path, err)
	}

	configFile, err := ini.Load(raw)
	if err != nil {
		return nil, fmt.Errorf("parse Azure CLI config %q: %w", path, err)
	}

	return configFile, nil
}

func ensureSection(configFile *ini.File, sectionName string) (*ini.Section, error) {
	section, err := configFile.GetSection(sectionName)
	if err == nil {
		return section, nil
	}

	section, err = configFile.NewSection(sectionName)
	if err != nil {
		return nil, err
	}

	return section, nil
}

func restoreLoginExperience(
	configFile *ini.File,
	path string,
	hadCoreSection bool,
	hadLoginExperienceKey bool,
	originalValue string,
) error {
	coreSection, err := configFile.GetSection(azureConfigCoreSection)
	if err != nil {
		return fmt.Errorf("get section %q: %w", azureConfigCoreSection, err)
	}

	if hadLoginExperienceKey {
		coreSection.Key(azureConfigLoginExperienceKey).SetValue(originalValue)
	} else {
		coreSection.DeleteKey(azureConfigLoginExperienceKey)
	}

	if !hadCoreSection && len(coreSection.Keys()) == 0 {
		configFile.DeleteSection(azureConfigCoreSection)
	}

	if err := writeAZConfig(path, configFile); err != nil {
		return fmt.Errorf("write Azure CLI config %q: %w", path, err)
	}

	return nil
}

func writeAZConfig(path string, configFile *ini.File) error {
	if err := fsys.MkdirAll(filepath.Dir(path), azureConfigDirectoryMode); err != nil {
		return fmt.Errorf("create Azure CLI config directory %q: %w", filepath.Dir(path), err)
	}

	var encoded bytes.Buffer
	if _, err := configFile.WriteTo(&encoded); err != nil {
		return fmt.Errorf("encode Azure CLI config: %w", err)
	}

	if err := afero.WriteFile(fsys, path, encoded.Bytes(), azureConfigFileMode); err != nil {
		return fmt.Errorf("write Azure CLI config file %q: %w", path, err)
	}

	return nil
}
