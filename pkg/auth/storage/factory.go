// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

//go:build !js

package storage

import (
	"fmt"
	"os"
	"sync"
)

var (
	// activeBackend caches the detected backend type
	activeBackend BackendType

	// cachedStorage stores the singleton storage instance
	cachedStorage Storage

	// fallbackWarningShown tracks if we've shown the fallback warning
	fallbackWarningShown bool

	// mu protects access to global storage state
	mu sync.Mutex
)

// GetStorage returns a storage instance, automatically detecting the best backend
//
// Selection priority:
//  1. DD_TOKEN_STORAGE=file → use file storage
//  2. DD_TOKEN_STORAGE=keychain → use keychain (fail if unavailable)
//  3. Auto-detect: try keychain, fall back to file with warning
//
// Example:
//
//	storage := GetStorage() // Auto-detect
//	storage := GetStorage(&StorageOptions{ForceBackend: BackendKeychain}) // Force keychain
func GetStorage(opts *StorageOptions) (Storage, error) {
	mu.Lock()
	defer mu.Unlock()

	// If forcing a specific backend, create and cache it
	if opts != nil && opts.ForceBackend != "" {
		storage, err := createStorage(opts.ForceBackend)
		if err != nil {
			return nil, err
		}
		activeBackend = opts.ForceBackend
		cachedStorage = storage
		return storage, nil
	}

	// Detect backend if not cached
	if cachedStorage == nil {
		backend, err := detectBackend()
		if err != nil {
			return nil, err
		}

		storage, err := createStorage(backend)
		if err != nil {
			return nil, err
		}

		activeBackend = backend
		cachedStorage = storage
	}

	return cachedStorage, nil
}

// detectBackend determines which storage backend to use
func detectBackend() (BackendType, error) {
	envSetting := os.Getenv(StorageEnvVar)

	// Explicit environment variable setting
	if envSetting != "" {
		switch envSetting {
		case "file":
			return BackendFile, nil

		case "keychain":
			// User explicitly requested keychain - verify it's available
			if !IsKeychainAvailable() {
				return "", fmt.Errorf(
					"DD_TOKEN_STORAGE=keychain specified but OS keychain is not available. " +
						"This may happen in headless environments, CI/CD, or Docker containers. " +
						"Remove DD_TOKEN_STORAGE or set it to 'file' to use file-based storage",
				)
			}
			return BackendKeychain, nil

		default:
			return "", fmt.Errorf("invalid DD_TOKEN_STORAGE value: %s (must be 'file' or 'keychain')", envSetting)
		}
	}

	// Auto-detect: try keychain first
	if IsKeychainAvailable() {
		return BackendKeychain, nil
	}

	// Fall back to file storage
	if !fallbackWarningShown {
		fmt.Fprintln(os.Stderr,
			"⚠️  Warning: OS keychain not available, falling back to file-based token storage.\n"+
				"   Tokens will be stored in ~/.config/pup/ with file permissions 0600.\n"+
				"   Set DD_TOKEN_STORAGE=file to suppress this warning.")
		fallbackWarningShown = true
	}

	return BackendFile, nil
}

// createStorage creates a storage instance of the specified type
func createStorage(backend BackendType) (Storage, error) {
	switch backend {
	case BackendKeychain:
		storage, err := NewKeychainStorage()
		if err != nil {
			return nil, fmt.Errorf("failed to create keychain storage: %w", err)
		}
		return storage, nil

	case BackendFile:
		storage, err := NewFileStorage()
		if err != nil {
			return nil, fmt.Errorf("failed to create file storage: %w", err)
		}
		return storage, nil

	default:
		return nil, fmt.Errorf("unknown backend type: %s", backend)
	}
}

// GetActiveBackend returns the currently active storage backend type
func GetActiveBackend() BackendType {
	mu.Lock()
	defer mu.Unlock()
	return activeBackend
}

// IsUsingSecureStorage returns true if keychain storage is active
func IsUsingSecureStorage() bool {
	mu.Lock()
	defer mu.Unlock()
	return activeBackend == BackendKeychain
}

// GetStorageDescription returns a human-readable description of current storage
func GetStorageDescription() string {
	storage, err := GetStorage(nil)
	if err != nil {
		return "unknown"
	}

	location := storage.GetStorageLocation()
	if storage.GetBackendType() == BackendKeychain {
		return fmt.Sprintf("%s (secure)", location)
	}
	return location
}

// ResetStorage clears the cached storage instance (for testing)
func ResetStorage() {
	mu.Lock()
	defer mu.Unlock()

	activeBackend = ""
	cachedStorage = nil
	fallbackWarningShown = false
}
