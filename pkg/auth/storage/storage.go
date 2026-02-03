// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DataDog/pup/pkg/auth/types"
)

// BackendType represents the type of storage backend
type BackendType string

const (
	BackendKeychain BackendType = "keychain"
	BackendFile     BackendType = "file"
)

// Storage interface for token and credential storage
type Storage interface {
	// GetBackendType returns the type of storage backend
	GetBackendType() BackendType

	// GetStorageLocation returns a human-readable description of storage location
	GetStorageLocation() string

	// SaveTokens saves OAuth2 tokens
	SaveTokens(site string, tokens *types.TokenSet) error

	// LoadTokens loads OAuth2 tokens
	LoadTokens(site string) (*types.TokenSet, error)

	// DeleteTokens deletes OAuth2 tokens
	DeleteTokens(site string) error

	// SaveClientCredentials saves DCR client credentials
	SaveClientCredentials(site string, creds *types.ClientCredentials) error

	// LoadClientCredentials loads DCR client credentials
	LoadClientCredentials(site string) (*types.ClientCredentials, error)

	// DeleteClientCredentials deletes DCR client credentials
	DeleteClientCredentials(site string) error
}

// FileStorage implements Storage using encrypted files
type FileStorage struct {
	baseDir string
}

// NewFileStorage creates a new file-based storage
func NewFileStorage() (*FileStorage, error) {
	// Get user config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config directory: %w", err)
	}

	// Create pup config directory
	baseDir := filepath.Join(configDir, "pup")
	if err := os.MkdirAll(baseDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	return &FileStorage{
		baseDir: baseDir,
	}, nil
}

// GetBackendType returns the backend type
func (s *FileStorage) GetBackendType() BackendType {
	return BackendFile
}

// GetStorageLocation returns the storage directory path
func (s *FileStorage) GetStorageLocation() string {
	return s.baseDir
}

// SaveTokens saves OAuth2 tokens to file
func (s *FileStorage) SaveTokens(site string, tokens *types.TokenSet) error {
	filename := filepath.Join(s.baseDir, fmt.Sprintf("tokens_%s.json", sanitizeSite(site)))

	data, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tokens: %w", err)
	}

	// Write with restricted permissions (0600 = rw-------)
	if err := os.WriteFile(filename, data, 0600); err != nil {
		return fmt.Errorf("failed to write tokens: %w", err)
	}

	return nil
}

// LoadTokens loads OAuth2 tokens from file
func (s *FileStorage) LoadTokens(site string) (*types.TokenSet, error) {
	filename := filepath.Join(s.baseDir, fmt.Sprintf("tokens_%s.json", sanitizeSite(site)))

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No tokens found
		}
		return nil, fmt.Errorf("failed to read tokens: %w", err)
	}

	var tokens types.TokenSet
	if err := json.Unmarshal(data, &tokens); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tokens: %w", err)
	}

	return &tokens, nil
}

// DeleteTokens deletes OAuth2 tokens
func (s *FileStorage) DeleteTokens(site string) error {
	filename := filepath.Join(s.baseDir, fmt.Sprintf("tokens_%s.json", sanitizeSite(site)))
	if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete tokens: %w", err)
	}
	return nil
}

// SaveClientCredentials saves DCR client credentials
func (s *FileStorage) SaveClientCredentials(site string, creds *types.ClientCredentials) error {
	filename := filepath.Join(s.baseDir, fmt.Sprintf("client_%s.json", sanitizeSite(site)))

	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	// Write with restricted permissions
	if err := os.WriteFile(filename, data, 0600); err != nil {
		return fmt.Errorf("failed to write credentials: %w", err)
	}

	return nil
}

// LoadClientCredentials loads DCR client credentials
func (s *FileStorage) LoadClientCredentials(site string) (*types.ClientCredentials, error) {
	filename := filepath.Join(s.baseDir, fmt.Sprintf("client_%s.json", sanitizeSite(site)))

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No credentials found
		}
		return nil, fmt.Errorf("failed to read credentials: %w", err)
	}

	var creds types.ClientCredentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("failed to unmarshal credentials: %w", err)
	}

	return &creds, nil
}

// DeleteClientCredentials deletes DCR client credentials
func (s *FileStorage) DeleteClientCredentials(site string) error {
	filename := filepath.Join(s.baseDir, fmt.Sprintf("client_%s.json", sanitizeSite(site)))
	if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete credentials: %w", err)
	}
	return nil
}

// sanitizeSite removes special characters from site name for filename
func sanitizeSite(site string) string {
	// Replace dots and other special chars with underscores
	safe := ""
	for _, c := range site {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			safe += string(c)
		} else {
			safe += "_"
		}
	}
	return safe
}
