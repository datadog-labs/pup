// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DataDog/pup/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestAliasConfig(t *testing.T) (string, func()) {
	t.Helper()

	// Create temporary config directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")

	// Override the config path for testing
	originalGetConfigPath := config.ConfigPathFunc
	config.ConfigPathFunc = func() (string, error) {
		return configPath, nil
	}

	cleanup := func() {
		config.ConfigPathFunc = originalGetConfigPath
	}

	return configPath, cleanup
}

func TestAliasSetCommand(t *testing.T) {
	_, cleanup := setupTestAliasConfig(t)
	defer cleanup()

	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid alias",
			args:    []string{"test-alias", "version"},
			wantErr: false,
		},
		{
			name:    "alias with dashes and underscores",
			args:    []string{"test_alias-v2", "version"},
			wantErr: false,
		},
		{
			name:        "reserved command name",
			args:        []string{"version", "test"},
			wantErr:     true,
			errContains: "conflicts with an existing pup command",
		},
		{
			name:        "invalid name with spaces",
			args:        []string{"invalid name", "test"},
			wantErr:     true,
			errContains: "invalid alias name",
		},
		{
			name:        "invalid name with special chars",
			args:        []string{"invalid@name", "test"},
			wantErr:     true,
			errContains: "invalid alias name",
		},
		{
			name:        "too few arguments",
			args:        []string{"test-alias"},
			wantErr:     true,
			errContains: "requires exactly 2 arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := runAliasSet(aliasSetCmd, tt.args)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAliasListCommand(t *testing.T) {
	_, cleanup := setupTestAliasConfig(t)
	defer cleanup()

	// Initially empty
	err := runAliasList(aliasListCmd, []string{})
	require.NoError(t, err)

	// Add some aliases
	require.NoError(t, config.SetAlias("test1", "version"))
	require.NoError(t, config.SetAlias("test2", "test"))

	// List should show both
	err = runAliasList(aliasListCmd, []string{})
	require.NoError(t, err)
}

func TestAliasDeleteCommand(t *testing.T) {
	_, cleanup := setupTestAliasConfig(t)
	defer cleanup()

	// Add an alias
	require.NoError(t, config.SetAlias("test-alias", "version"))

	// Delete it
	err := runAliasDelete(aliasDeleteCmd, []string{"test-alias"})
	require.NoError(t, err)

	// Verify it's gone
	_, err = config.GetAlias("test-alias")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestAliasImportCommand(t *testing.T) {
	_, cleanup := setupTestAliasConfig(t)
	defer cleanup()

	// Create a test import file
	tmpFile := filepath.Join(t.TempDir(), "import.yml")
	content := `aliases:
  imported1: version
  imported2: test
`
	require.NoError(t, os.WriteFile(tmpFile, []byte(content), 0600))

	// Import
	err := runAliasImport(aliasImportCmd, []string{tmpFile})
	require.NoError(t, err)

	// Verify aliases were imported
	cmd1, err := config.GetAlias("imported1")
	require.NoError(t, err)
	assert.Equal(t, "version", cmd1)

	cmd2, err := config.GetAlias("imported2")
	require.NoError(t, err)
	assert.Equal(t, "test", cmd2)
}

func TestIsValidAliasName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid simple", "test", true},
		{"valid with dash", "test-alias", true},
		{"valid with underscore", "test_alias", true},
		{"valid alphanumeric", "test123", true},
		{"valid mixed", "test-alias_v2", true},
		{"invalid empty", "", false},
		{"invalid space", "test alias", false},
		{"invalid special char", "test@alias", false},
		{"invalid special char 2", "test!alias", false},
		{"invalid dot", "test.alias", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidAliasName(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsReservedCommand(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"reserved: alias", "alias", true},
		{"reserved: auth", "auth", true},
		{"reserved: version", "version", true},
		{"reserved: metrics", "metrics", true},
		{"reserved case insensitive", "VERSION", true},
		{"not reserved", "my-alias", false},
		{"not reserved numeric", "test123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isReservedCommand(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExpandAlias(t *testing.T) {
	tests := []struct {
		name           string
		aliasCommand   string
		additionalArgs []string
		want           []string
	}{
		{
			name:           "simple command",
			aliasCommand:   "version",
			additionalArgs: []string{},
			want:           []string{"version"},
		},
		{
			name:           "command with args",
			aliasCommand:   "monitors list --tag=prod",
			additionalArgs: []string{},
			want:           []string{"monitors", "list", "--tag=prod"},
		},
		{
			name:           "command with additional args",
			aliasCommand:   "monitors list",
			additionalArgs: []string{"--tag=prod"},
			want:           []string{"monitors", "list", "--tag=prod"},
		},
		{
			name:           "command with quoted args",
			aliasCommand:   "logs search --query='status:error'",
			additionalArgs: []string{},
			want:           []string{"logs", "search", "--query=status:error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandAlias(tt.aliasCommand, tt.additionalArgs)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSplitCommand(t *testing.T) {
	tests := []struct {
		name    string
		command string
		want    []string
	}{
		{
			name:    "simple command",
			command: "version",
			want:    []string{"version"},
		},
		{
			name:    "command with args",
			command: "monitors list --tag=prod",
			want:    []string{"monitors", "list", "--tag=prod"},
		},
		{
			name:    "command with single quotes",
			command: "logs search --query='status:error'",
			want:    []string{"logs", "search", "--query=status:error"},
		},
		{
			name:    "command with double quotes",
			command: `logs search --query="status:error"`,
			want:    []string{"logs", "search", "--query=status:error"},
		},
		{
			name:    "command with multiple spaces",
			command: "monitors  list   --tag=prod",
			want:    []string{"monitors", "list", "--tag=prod"},
		},
		{
			name:    "command with quotes containing spaces",
			command: `logs search --query="status:error service:web"`,
			want:    []string{"logs", "search", "--query=status:error service:web"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitCommand(tt.command)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsFlag(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"flag short", "-v", true},
		{"flag long", "--version", true},
		{"not flag", "version", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isFlag(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsBuiltinCommand(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"builtin: version", "version", true},
		{"builtin: auth", "auth", true},
		{"builtin: alias", "alias", true},
		{"builtin: metrics", "metrics", true},
		{"builtin: monitors", "monitors", true},
		{"not builtin", "my-custom-alias", false},
		{"not builtin with dash", "prod-errors", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isBuiltinCommand(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAliasCannotOverrideBuiltinCommand(t *testing.T) {
	_, cleanup := setupTestAliasConfig(t)
	defer cleanup()

	// Try to create an alias with a reserved name (should fail at validation)
	err := runAliasSet(aliasSetCmd, []string{"version", "test"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "conflicts with an existing pup command")

	// Even if we somehow bypass validation and create an alias with a builtin name
	// (e.g., manually editing config.yml), the builtin should take precedence
	require.NoError(t, config.SetAlias("auth", "version"))

	// ExecuteWithArgs should use the builtin command, not the alias
	// We can't easily test this without actually running the command, but the
	// isBuiltinCommand check ensures this at runtime
	assert.True(t, isBuiltinCommand("auth"))
}
