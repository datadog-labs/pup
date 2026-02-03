// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package storage

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/99designs/keyring"
	"github.com/DataDog/pup/pkg/auth/types"
)

const (
	// KeychainService is the service name for keychain entries
	KeychainService = "datadog-pup-cli"

	// TokenPrefix is the prefix for token keychain entries
	TokenPrefix = "oauth-tokens:"

	// ClientPrefix is the prefix for client credential keychain entries
	ClientPrefix = "oauth-client:"
)

// KeychainStorage stores OAuth tokens and credentials in the OS keychain
type KeychainStorage struct {
	keyring keyring.Keyring
}

// NewKeychainStorage creates a new keychain storage instance
func NewKeychainStorage() (*KeychainStorage, error) {
	ring, err := keyring.Open(keyring.Config{
		ServiceName:              KeychainService,
		AllowedBackends:          []keyring.BackendType{keyring.KeychainBackend, keyring.WinCredBackend, keyring.SecretServiceBackend},
		KeychainTrustApplication: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open keychain: %w", err)
	}

	return &KeychainStorage{
		keyring: ring,
	}, nil
}

// GetBackendType returns the backend type
func (s *KeychainStorage) GetBackendType() BackendType {
	return BackendKeychain
}

// GetStorageLocation returns a human-readable description
func (s *KeychainStorage) GetStorageLocation() string {
	switch runtime.GOOS {
	case "darwin":
		return "macOS Keychain"
	case "windows":
		return "Windows Credential Manager"
	case "linux":
		return "System Keychain (Secret Service)"
	default:
		return "System Keychain"
	}
}

// SaveTokens saves OAuth tokens for a site
func (s *KeychainStorage) SaveTokens(site string, tokens *types.TokenSet) error {
	data, err := json.Marshal(tokens)
	if err != nil {
		return fmt.Errorf("failed to marshal tokens: %w", err)
	}

	key := TokenPrefix + site
	item := keyring.Item{
		Key:  key,
		Data: data,
	}

	if err := s.keyring.Set(item); err != nil {
		return fmt.Errorf("failed to save tokens to keychain: %w", err)
	}

	return nil
}

// LoadTokens loads OAuth tokens for a site
func (s *KeychainStorage) LoadTokens(site string) (*types.TokenSet, error) {
	key := TokenPrefix + site
	item, err := s.keyring.Get(key)
	if err != nil {
		if err == keyring.ErrKeyNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to load tokens from keychain: %w", err)
	}

	var tokens types.TokenSet
	if err := json.Unmarshal(item.Data, &tokens); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tokens: %w", err)
	}

	return &tokens, nil
}

// DeleteTokens deletes OAuth tokens for a site
func (s *KeychainStorage) DeleteTokens(site string) error {
	key := TokenPrefix + site
	if err := s.keyring.Remove(key); err != nil {
		if err == keyring.ErrKeyNotFound {
			return nil
		}
		return fmt.Errorf("failed to delete tokens from keychain: %w", err)
	}
	return nil
}

// SaveClientCredentials saves OAuth client credentials for a site
func (s *KeychainStorage) SaveClientCredentials(site string, creds *types.ClientCredentials) error {
	data, err := json.Marshal(creds)
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	key := ClientPrefix + site
	item := keyring.Item{
		Key:  key,
		Data: data,
	}

	if err := s.keyring.Set(item); err != nil {
		return fmt.Errorf("failed to save credentials to keychain: %w", err)
	}

	return nil
}

// LoadClientCredentials loads OAuth client credentials for a site
func (s *KeychainStorage) LoadClientCredentials(site string) (*types.ClientCredentials, error) {
	key := ClientPrefix + site
	item, err := s.keyring.Get(key)
	if err != nil {
		if err == keyring.ErrKeyNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to load credentials from keychain: %w", err)
	}

	var creds types.ClientCredentials
	if err := json.Unmarshal(item.Data, &creds); err != nil {
		return nil, fmt.Errorf("failed to unmarshal credentials: %w", err)
	}

	return &creds, nil
}

// DeleteClientCredentials deletes OAuth client credentials for a site
func (s *KeychainStorage) DeleteClientCredentials(site string) error {
	key := ClientPrefix + site
	if err := s.keyring.Remove(key); err != nil {
		if err == keyring.ErrKeyNotFound {
			return nil
		}
		return fmt.Errorf("failed to delete credentials from keychain: %w", err)
	}
	return nil
}

// IsKeychainAvailable checks if keychain storage is available on this system
func IsKeychainAvailable() bool {
	_, err := keyring.Open(keyring.Config{
		ServiceName:              KeychainService + "-test",
		AllowedBackends:          []keyring.BackendType{keyring.KeychainBackend, keyring.WinCredBackend, keyring.SecretServiceBackend},
		KeychainTrustApplication: true,
	})
	return err == nil
}
