package az

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ini "gopkg.in/ini.v1"
)

func TestAZConfigPath(t *testing.T) {
	t.Run("uses AZURE_CONFIG_DIR when set", func(t *testing.T) {
		customDir := filepath.Clean("/tmp/custom-azure")
		t.Setenv(azureConfigDirEnv, customDir)

		path, err := azConfigPath()
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(customDir, azureConfigFile), path)
	})

	t.Run("falls back to home directory", func(t *testing.T) {
		t.Setenv(azureConfigDirEnv, "")

		homeDir, err := os.UserHomeDir()
		require.NoError(t, err)

		path, err := azConfigPath()
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(homeDir, azureConfigDirectory, azureConfigFile), path)
	})
}

func TestWithLoginExperienceOff(t *testing.T) {
	t.Run("sets and removes key when initially absent", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		SetFS(t, fs)

		configDir := filepath.Clean("/tmp/azure")
		t.Setenv(azureConfigDirEnv, configDir)
		configPath := filepath.Join(configDir, azureConfigFile)

		callbackInvoked := false
		err := withLoginExperienceOff(func() error {
			callbackInvoked = true
			current := readINIFile(t, fs, configPath)
			coreSection, sectionErr := current.GetSection(azureConfigCoreSection)
			require.NoError(t, sectionErr)
			assert.Equal(
				t,
				azureConfigLoginExperienceOff,
				coreSection.Key(azureConfigLoginExperienceKey).String(),
			)

			return nil
		})
		require.NoError(t, err)
		assert.True(t, callbackInvoked)

		restored := readINIFile(t, fs, configPath)
		assert.False(t, restored.HasSection(azureConfigCoreSection))
	})

	t.Run("restores existing key value", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		SetFS(t, fs)

		configDir := filepath.Clean("/tmp/azure")
		t.Setenv(azureConfigDirEnv, configDir)
		configPath := filepath.Join(configDir, azureConfigFile)

		original := ini.Empty()
		coreSection, err := original.NewSection(azureConfigCoreSection)
		require.NoError(t, err)
		coreSection.Key(azureConfigLoginExperienceKey).SetValue("on")
		coreSection.Key("other_key").SetValue("keep")
		writeINIFile(t, fs, configPath, original)

		err = withLoginExperienceOff(func() error {
			current := readINIFile(t, fs, configPath)
			currentCore, sectionErr := current.GetSection(azureConfigCoreSection)
			require.NoError(t, sectionErr)
			assert.Equal(t, azureConfigLoginExperienceOff, currentCore.Key(azureConfigLoginExperienceKey).String())
			assert.Equal(t, "keep", currentCore.Key("other_key").String())

			return nil
		})
		require.NoError(t, err)

		restored := readINIFile(t, fs, configPath)
		restoredCore, sectionErr := restored.GetSection(azureConfigCoreSection)
		require.NoError(t, sectionErr)
		assert.Equal(t, "on", restoredCore.Key(azureConfigLoginExperienceKey).String())
		assert.Equal(t, "keep", restoredCore.Key("other_key").String())
	})

	t.Run("restores key when callback fails", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		SetFS(t, fs)

		configDir := filepath.Clean("/tmp/azure")
		t.Setenv(azureConfigDirEnv, configDir)
		configPath := filepath.Join(configDir, azureConfigFile)

		original := ini.Empty()
		coreSection, err := original.NewSection(azureConfigCoreSection)
		require.NoError(t, err)
		coreSection.Key(azureConfigLoginExperienceKey).SetValue("on")
		writeINIFile(t, fs, configPath, original)

		sentinelErr := errors.New("callback failed")
		err = withLoginExperienceOff(func() error {
			return sentinelErr
		})
		require.ErrorIs(t, err, sentinelErr)

		restored := readINIFile(t, fs, configPath)
		restoredCore, sectionErr := restored.GetSection(azureConfigCoreSection)
		require.NoError(t, sectionErr)
		assert.Equal(t, "on", restoredCore.Key(azureConfigLoginExperienceKey).String())
	})
}

func readINIFile(t *testing.T, fs afero.Fs, path string) *ini.File {
	t.Helper()

	raw, err := afero.ReadFile(fs, path)
	require.NoError(t, err)

	parsed, err := ini.Load(raw)
	require.NoError(t, err)

	return parsed
}

func writeINIFile(t *testing.T, fs afero.Fs, path string, file *ini.File) {
	t.Helper()

	require.NoError(t, fs.MkdirAll(filepath.Dir(path), 0o700))

	var encoded bytes.Buffer
	_, err := file.WriteTo(&encoded)
	require.NoError(t, err)

	require.NoError(t, afero.WriteFile(fs, path, encoded.Bytes(), 0o600))
}
