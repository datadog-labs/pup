// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// AliasConfig represents the alias configuration structure
type AliasConfig struct {
	Aliases map[string]string `yaml:"aliases"`
}

// ConfigPathFunc is a function variable that returns the config file path
// This can be overridden in tests
var ConfigPathFunc = getDefaultConfigPath

// GetConfigPath returns the path to the config file
func GetConfigPath() (string, error) {
	return ConfigPathFunc()
}

// getDefaultConfigPath is the default implementation
func getDefaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".config", "pup", "config.yml"), nil
}

// LoadAliases loads aliases from the config file
func LoadAliases() (map[string]string, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// If config file doesn't exist, return empty aliases
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return make(map[string]string), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config AliasConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if config.Aliases == nil {
		config.Aliases = make(map[string]string)
	}

	return config.Aliases, nil
}

// SaveAliases saves aliases to the config file
func SaveAliases(aliases map[string]string) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	config := AliasConfig{
		Aliases: aliases,
	}

	data, err := yaml.Marshal(&config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetAlias retrieves a specific alias by name
func GetAlias(name string) (string, error) {
	aliases, err := LoadAliases()
	if err != nil {
		return "", err
	}

	command, ok := aliases[name]
	if !ok {
		return "", fmt.Errorf("alias '%s' not found", name)
	}

	return command, nil
}

// SetAlias sets or updates an alias
func SetAlias(name, command string) error {
	aliases, err := LoadAliases()
	if err != nil {
		return err
	}

	aliases[name] = command
	return SaveAliases(aliases)
}

// DeleteAlias removes an alias
func DeleteAlias(name string) error {
	aliases, err := LoadAliases()
	if err != nil {
		return err
	}

	if _, ok := aliases[name]; !ok {
		return fmt.Errorf("alias '%s' not found", name)
	}

	delete(aliases, name)
	return SaveAliases(aliases)
}

// ImportAliases imports aliases from a YAML file
func ImportAliases(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read import file: %w", err)
	}

	var importConfig AliasConfig
	if err := yaml.Unmarshal(data, &importConfig); err != nil {
		return fmt.Errorf("failed to parse import file: %w", err)
	}

	// Load existing aliases
	aliases, err := LoadAliases()
	if err != nil {
		return err
	}

	// Merge imported aliases
	for name, command := range importConfig.Aliases {
		aliases[name] = command
	}

	return SaveAliases(aliases)
}
