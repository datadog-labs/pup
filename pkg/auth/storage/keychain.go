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
	// KeychainTokenService is the service name for token keychain entries
	// Matches TypeScript PR #84: "datadog-cli"
	KeychainTokenService = "datadog-cli"

	// KeychainClientService is the service name for client credential keychain entries
	// Matches TypeScript PR #84: "datadog-cli-dcr"
	KeychainClientService = "datadog-cli-dcr"

	// TokenPrefix is the prefix for token keychain account names
	// Matches TypeScript PR #84: "oauth:"
	TokenPrefix = "oauth:"

	// ClientPrefix is the prefix for client credential keychain account names
	// Matches TypeScript PR #84: "client:"
	ClientPrefix = "client:"
)

// KeychainStorage stores OAuth tokens and credentials in the OS keychain
// Uses separate keyring services to match TypeScript PR #84
type KeychainStorage struct {
	tokenKeyring  keyring.Keyring // For tokens (service: "datadog-cli")
	clientKeyring keyring.Keyring // For client credentials (service: "datadog-cli-dcr")
}

// NewKeychainStorage creates a new keychain storage instance
// Opens two separate keychains to match TypeScript PR #84 architecture
func NewKeychainStorage() (*KeychainStorage, error) {
	// Open keyring for tokens
	tokenRing, err := keyring.Open(keyring.Config{
		ServiceName:              KeychainTokenService,
		AllowedBackends:          []keyring.BackendType{keyring.KeychainBackend, keyring.WinCredBackend, keyring.SecretServiceBackend},
		KeychainTrustApplication: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open token keychain: %w", err)
	}

	// Open keyring for client credentials
	clientRing, err := keyring.Open(keyring.Config{
		ServiceName:              KeychainClientService,
		AllowedBackends:          []keyring.BackendType{keyring.KeychainBackend, keyring.WinCredBackend, keyring.SecretServiceBackend},
		KeychainTrustApplication: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open client keychain: %w", err)
	}

	return &KeychainStorage{
		tokenKeyring:  tokenRing,
		clientKeyring: clientRing,
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

	if err := s.tokenKeyring.Set(item); err != nil {
		return fmt.Errorf("failed to save tokens to keychain: %w", err)
	}

	return nil
}

// LoadTokens loads OAuth tokens for a site
func (s *KeychainStorage) LoadTokens(site string) (*types.TokenSet, error) {
	key := TokenPrefix + site
	item, err := s.tokenKeyring.Get(key)
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
	if err := s.tokenKeyring.Remove(key); err != nil {
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

	if err := s.clientKeyring.Set(item); err != nil {
		return fmt.Errorf("failed to save credentials to keychain: %w", err)
	}

	return nil
}

// LoadClientCredentials loads OAuth client credentials for a site
func (s *KeychainStorage) LoadClientCredentials(site string) (*types.ClientCredentials, error) {
	key := ClientPrefix + site
	item, err := s.clientKeyring.Get(key)
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
	if err := s.clientKeyring.Remove(key); err != nil {
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
		ServiceName:              KeychainTokenService + "-test",
		AllowedBackends:          []keyring.BackendType{keyring.KeychainBackend, keyring.WinCredBackend, keyring.SecretServiceBackend},
		KeychainTrustApplication: true,
	})
	return err == nil
}
