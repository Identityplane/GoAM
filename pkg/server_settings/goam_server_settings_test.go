package server_settings

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestServerSettings_InitWithViper(t *testing.T) {

	config_file := "test_settings.yaml"
	working_dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	working_dirconfig_file_full_path := filepath.Join(working_dir, config_file)

	t.Run("Read settings from file", func(t *testing.T) {

		// Arrange
		viper.SetConfigFile(working_dirconfig_file_full_path)

		// Act
		settings, err := InitWithViper()
		assert.NoError(t, err)

		// Assert
		assert.Equal(t, "test-from-file", settings.Banner)

	})

	t.Run("Read settings from env variables takes precedence over config file", func(t *testing.T) {

		// Arrange
		viper.SetConfigFile(working_dirconfig_file_full_path)
		os.Setenv("GOAM_BANNER", "test-from-env")
		defer os.Unsetenv("GOAM_BANNER")

		// Act
		settings, err := InitWithViper()
		assert.NoError(t, err)

		// Assert
		assert.Equal(t, "test-from-env", settings.Banner)
	})

	t.Run("Default settings if no config file or env is present", func(t *testing.T) {

		// Act
		settings, err := InitWithViper()
		assert.NoError(t, err)

		// Assert
		assert.Equal(t, "GoAM", settings.Banner)
	})

	t.Run("Env if no config file is present", func(t *testing.T) {

		// Arrange
		os.Setenv("GOAM_BANNER", "test-from-env-2")
		defer os.Unsetenv("GOAM_BANNER")

		// Act
		settings, err := InitWithViper()
		assert.NoError(t, err)

		// Assert
		assert.Equal(t, "test-from-env-2", settings.Banner)
	})

	t.Run("Read settings from mapstructure", func(t *testing.T) {

		// Arrange
		viper.SetConfigFile(working_dirconfig_file_full_path)

		// Arrange
		settings, err := InitWithViper()
		assert.NoError(t, err)

		// Assert
		assert.Equal(t, "test-from-file", settings.ExtensionSettings["test_setting"])
	})
}
