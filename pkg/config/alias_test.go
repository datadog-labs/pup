// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadAliases(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")

	// Mock GetConfigPath
	originalGetConfigPath := ConfigPathFunc
	ConfigPathFunc = func() (string, error) {
		return configPath, nil
	}
	defer func() { ConfigPathFunc = originalGetConfigPath }()

	t.Run("empty config file", func(t *testing.T) {
		aliases, err := LoadAliases()
		require.NoError(t, err)
		assert.Empty(t, aliases)
	})

	t.Run("config file with aliases", func(t *testing.T) {
		content := `aliases:
  test1: version
  test2: test
`
		require.NoError(t, os.WriteFile(configPath, []byte(content), 0600))

		aliases, err := LoadAliases()
		require.NoError(t, err)
		assert.Len(t, aliases, 2)
		assert.Equal(t, "version", aliases["test1"])
		assert.Equal(t, "test", aliases["test2"])
	})
}

func TestSaveAliases(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")

	// Mock GetConfigPath
	originalGetConfigPath := ConfigPathFunc
	ConfigPathFunc = func() (string, error) {
		return configPath, nil
	}
	defer func() { ConfigPathFunc = originalGetConfigPath }()

	aliases := map[string]string{
		"test1": "version",
		"test2": "test",
	}

	err := SaveAliases(aliases)
	require.NoError(t, err)

	// Verify file was created with correct permissions
	info, err := os.Stat(configPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())

	// Verify content
	loaded, err := LoadAliases()
	require.NoError(t, err)
	assert.Equal(t, aliases, loaded)
}

func TestGetAlias(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")

	// Mock GetConfigPath
	originalGetConfigPath := ConfigPathFunc
	ConfigPathFunc = func() (string, error) {
		return configPath, nil
	}
	defer func() { ConfigPathFunc = originalGetConfigPath }()

	// Set up test alias
	require.NoError(t, SetAlias("test-alias", "version"))

	t.Run("existing alias", func(t *testing.T) {
		command, err := GetAlias("test-alias")
		require.NoError(t, err)
		assert.Equal(t, "version", command)
	})

	t.Run("non-existing alias", func(t *testing.T) {
		_, err := GetAlias("nonexistent")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestSetAlias(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")

	// Mock GetConfigPath
	originalGetConfigPath := ConfigPathFunc
	ConfigPathFunc = func() (string, error) {
		return configPath, nil
	}
	defer func() { ConfigPathFunc = originalGetConfigPath }()

	t.Run("set new alias", func(t *testing.T) {
		err := SetAlias("test-alias", "version")
		require.NoError(t, err)

		// Verify it was saved
		command, err := GetAlias("test-alias")
		require.NoError(t, err)
		assert.Equal(t, "version", command)
	})

	t.Run("update existing alias", func(t *testing.T) {
		err := SetAlias("test-alias", "test")
		require.NoError(t, err)

		// Verify it was updated
		command, err := GetAlias("test-alias")
		require.NoError(t, err)
		assert.Equal(t, "test", command)
	})
}

func TestDeleteAlias(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")

	// Mock GetConfigPath
	originalGetConfigPath := ConfigPathFunc
	ConfigPathFunc = func() (string, error) {
		return configPath, nil
	}
	defer func() { ConfigPathFunc = originalGetConfigPath }()

	// Set up test alias
	require.NoError(t, SetAlias("test-alias", "version"))

	t.Run("delete existing alias", func(t *testing.T) {
		err := DeleteAlias("test-alias")
		require.NoError(t, err)

		// Verify it was deleted
		_, err = GetAlias("test-alias")
		require.Error(t, err)
	})

	t.Run("delete non-existing alias", func(t *testing.T) {
		err := DeleteAlias("nonexistent")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestImportAliases(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")

	// Mock GetConfigPath
	originalGetConfigPath := ConfigPathFunc
	ConfigPathFunc = func() (string, error) {
		return configPath, nil
	}
	defer func() { ConfigPathFunc = originalGetConfigPath }()

	t.Run("import valid file", func(t *testing.T) {
		// Create import file
		importFile := filepath.Join(tmpDir, "import.yml")
		content := `aliases:
  imported1: version
  imported2: test
`
		require.NoError(t, os.WriteFile(importFile, []byte(content), 0600))

		// Import
		err := ImportAliases(importFile)
		require.NoError(t, err)

		// Verify aliases were imported
		cmd1, err := GetAlias("imported1")
		require.NoError(t, err)
		assert.Equal(t, "version", cmd1)

		cmd2, err := GetAlias("imported2")
		require.NoError(t, err)
		assert.Equal(t, "test", cmd2)
	})

	t.Run("import merges with existing", func(t *testing.T) {
		// Set existing alias
		require.NoError(t, SetAlias("existing", "test"))

		// Create import file
		importFile := filepath.Join(tmpDir, "import2.yml")
		content := `aliases:
  new-alias: version
`
		require.NoError(t, os.WriteFile(importFile, []byte(content), 0600))

		// Import
		err := ImportAliases(importFile)
		require.NoError(t, err)

		// Verify both exist
		cmd1, err := GetAlias("existing")
		require.NoError(t, err)
		assert.Equal(t, "test", cmd1)

		cmd2, err := GetAlias("new-alias")
		require.NoError(t, err)
		assert.Equal(t, "version", cmd2)
	})

	t.Run("import non-existing file", func(t *testing.T) {
		err := ImportAliases("/nonexistent/file.yml")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read import file")
	})

	t.Run("import invalid yaml", func(t *testing.T) {
		importFile := filepath.Join(tmpDir, "invalid.yml")
		content := `this is not valid yaml: [[[`
		require.NoError(t, os.WriteFile(importFile, []byte(content), 0600))

		err := ImportAliases(importFile)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse import file")
	})
}

func TestGetConfigPath(t *testing.T) {
	t.Run("default config path", func(t *testing.T) {
		// Reset to default
		ConfigPathFunc = getDefaultConfigPath
		defer func() { ConfigPathFunc = getDefaultConfigPath }()

		path, err := GetConfigPath()
		require.NoError(t, err)
		assert.Contains(t, path, ".config/pup/config.yml")
		assert.Contains(t, path, string(filepath.Separator))
	})
}

func TestLoadAliasesErrorPaths(t *testing.T) {
	t.Run("error getting config path", func(t *testing.T) {
		originalGetConfigPath := ConfigPathFunc
		ConfigPathFunc = func() (string, error) {
			return "", fmt.Errorf("mock error")
		}
		defer func() { ConfigPathFunc = originalGetConfigPath }()

		_, err := LoadAliases()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "mock error")
	})

	t.Run("error reading existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "unreadable.yml")

		// Create a directory with the config file name so it can't be read as a file
		require.NoError(t, os.Mkdir(configPath, 0755))

		originalGetConfigPath := ConfigPathFunc
		ConfigPathFunc = func() (string, error) {
			return configPath, nil
		}
		defer func() { ConfigPathFunc = originalGetConfigPath }()

		_, err := LoadAliases()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read config file")
	})
}

func TestSaveAliasesErrorPaths(t *testing.T) {
	t.Run("error getting config path", func(t *testing.T) {
		originalGetConfigPath := ConfigPathFunc
		ConfigPathFunc = func() (string, error) {
			return "", fmt.Errorf("mock error")
		}
		defer func() { ConfigPathFunc = originalGetConfigPath }()

		err := SaveAliases(map[string]string{"test": "value"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "mock error")
	})

	t.Run("error creating directory", func(t *testing.T) {
		// Create a file where the directory should be
		tmpDir := t.TempDir()
		blockingFile := filepath.Join(tmpDir, "blocking")
		require.NoError(t, os.WriteFile(blockingFile, []byte("test"), 0644))

		configPath := filepath.Join(blockingFile, "config.yml")

		originalGetConfigPath := ConfigPathFunc
		ConfigPathFunc = func() (string, error) {
			return configPath, nil
		}
		defer func() { ConfigPathFunc = originalGetConfigPath }()

		err := SaveAliases(map[string]string{"test": "value"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create config directory")
	})

	t.Run("error writing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		// Create a directory where the file should be
		configPath := filepath.Join(tmpDir, "config.yml")
		require.NoError(t, os.Mkdir(configPath, 0755))

		originalGetConfigPath := ConfigPathFunc
		ConfigPathFunc = func() (string, error) {
			return configPath, nil
		}
		defer func() { ConfigPathFunc = originalGetConfigPath }()

		err := SaveAliases(map[string]string{"test": "value"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write config file")
	})
}

func TestGetAliasErrorPath(t *testing.T) {
	t.Run("error loading aliases", func(t *testing.T) {
		originalGetConfigPath := ConfigPathFunc
		ConfigPathFunc = func() (string, error) {
			return "", fmt.Errorf("mock error")
		}
		defer func() { ConfigPathFunc = originalGetConfigPath }()

		_, err := GetAlias("test")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "mock error")
	})
}

func TestSetAliasErrorPath(t *testing.T) {
	t.Run("error loading aliases", func(t *testing.T) {
		originalGetConfigPath := ConfigPathFunc
		ConfigPathFunc = func() (string, error) {
			return "", fmt.Errorf("mock error")
		}
		defer func() { ConfigPathFunc = originalGetConfigPath }()

		err := SetAlias("test", "value")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "mock error")
	})
}

func TestDeleteAliasErrorPath(t *testing.T) {
	t.Run("error loading aliases", func(t *testing.T) {
		originalGetConfigPath := ConfigPathFunc
		ConfigPathFunc = func() (string, error) {
			return "", fmt.Errorf("mock error")
		}
		defer func() { ConfigPathFunc = originalGetConfigPath }()

		err := DeleteAlias("test")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "mock error")
	})
}

func TestImportAliasesErrorPath(t *testing.T) {
	t.Run("error loading existing aliases", func(t *testing.T) {
		tmpDir := t.TempDir()
		importFile := filepath.Join(tmpDir, "import.yml")
		content := `aliases:
  test: value
`
		require.NoError(t, os.WriteFile(importFile, []byte(content), 0600))

		originalGetConfigPath := ConfigPathFunc
		ConfigPathFunc = func() (string, error) {
			return "", fmt.Errorf("mock error")
		}
		defer func() { ConfigPathFunc = originalGetConfigPath }()

		err := ImportAliases(importFile)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "mock error")
	})
}
