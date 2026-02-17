// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package storage

import (
	"os"
	"testing"
)

// NOTE: Tests in this file must NOT use t.Parallel() because they mutate
// shared global state: the singleton storage cache (ResetStorage/GetStorage)
// and the DD_TOKEN_STORAGE environment variable (os.Setenv/t.Setenv).
// Running them concurrently causes races where one test's env var change
// or cache reset is visible to another test, leading to flaky failures.

func TestGetStorage(t *testing.T) {
	tests := []struct {
		name              string
		envValue          string
		forceBackend      BackendType
		expectedBackend   BackendType
		shouldError       bool
		keychainAvailable bool
	}{
		{
			name:              "auto-detect keychain when available",
			envValue:          "",
			forceBackend:      "",
			expectedBackend:   BackendKeychain,
			shouldError:       false,
			keychainAvailable: true,
		},
		{
			name:              "auto-detect falls back to file when keychain unavailable",
			envValue:          "",
			forceBackend:      "",
			expectedBackend:   BackendFile,
			shouldError:       false,
			keychainAvailable: false,
		},
		{
			name:              "DD_TOKEN_STORAGE=file forces file backend",
			envValue:          "file",
			forceBackend:      "",
			expectedBackend:   BackendFile,
			shouldError:       false,
			keychainAvailable: true,
		},
		{
			name:              "DD_TOKEN_STORAGE=keychain forces keychain when available",
			envValue:          "keychain",
			forceBackend:      "",
			expectedBackend:   BackendKeychain,
			shouldError:       false,
			keychainAvailable: true,
		},
		// Note: This test is commented out because we can't mock keychain availability
		// {
		// 	name:              "DD_TOKEN_STORAGE=keychain errors when unavailable",
		// 	envValue:          "keychain",
		// 	forceBackend:      "",
		// 	expectedBackend:   "",
		// 	shouldError:       true,
		// 	keychainAvailable: false,
		// },
		{
			name:              "forceBackend=keychain succeeds when available",
			envValue:          "",
			forceBackend:      BackendKeychain,
			expectedBackend:   BackendKeychain,
			shouldError:       false,
			keychainAvailable: true,
		},
		{
			name:              "forceBackend=file always succeeds",
			envValue:          "",
			forceBackend:      BackendFile,
			expectedBackend:   BackendFile,
			shouldError:       false,
			keychainAvailable: true,
		},
		{
			name:              "invalid DD_TOKEN_STORAGE value errors",
			envValue:          "invalid",
			forceBackend:      "",
			expectedBackend:   "",
			shouldError:       true,
			keychainAvailable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset storage state
			ResetStorage()

			// Skip tests that require keychain when it's not available
			if tt.keychainAvailable && !IsKeychainAvailable() {
				t.Skip("Keychain not available in test environment")
			}

			// Set environment variable
			if tt.envValue != "" {
				t.Setenv(StorageEnvVar, tt.envValue)
			}

			// Mock keychain availability
			// Note: This is a limitation - we can't easily mock IsKeychainAvailable() in the current design
			// In real tests, keychain availability depends on the environment

			var opts *StorageOptions
			if tt.forceBackend != "" {
				opts = &StorageOptions{ForceBackend: tt.forceBackend}
			}

			storage, err := GetStorage(opts)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if storage == nil {
				t.Fatal("Storage is nil")
			}

			backend := storage.GetBackendType()

			// Skip backend type check for auto-detect tests since we can't mock keychain availability
			if tt.envValue == "" && tt.forceBackend == "" {
				// Just verify we got some valid backend
				if backend != BackendKeychain && backend != BackendFile {
					t.Errorf("Expected valid backend, got %v", backend)
				}
				return
			}

			if backend != tt.expectedBackend {
				t.Errorf("Expected backend %v, got %v", tt.expectedBackend, backend)
			}
		})
	}
}

func TestGetActiveBackend(t *testing.T) {
	// Reset storage state
	ResetStorage()

	// Before any storage is requested, should be empty
	if backend := GetActiveBackend(); backend != "" {
		t.Errorf("Expected empty backend before GetStorage called, got %v", backend)
	}

	// After requesting storage, should return the active backend
	storage, err := GetStorage(&StorageOptions{ForceBackend: BackendFile})
	if err != nil {
		t.Fatalf("Failed to get storage: %v", err)
	}

	if storage.GetBackendType() != BackendFile {
		t.Errorf("Expected file backend, got %v", storage.GetBackendType())
	}

	if backend := GetActiveBackend(); backend != BackendFile {
		t.Errorf("Expected active backend to be file, got %v", backend)
	}
}

func TestIsUsingSecureStorage(t *testing.T) {
	tests := []struct {
		name     string
		backend  BackendType
		expected bool
	}{
		{
			name:     "keychain is secure",
			backend:  BackendKeychain,
			expected: true,
		},
		{
			name:     "file is not secure",
			backend:  BackendFile,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ResetStorage()

			// Force the backend type we want to test
			// Note: BackendKeychain may fail if keychain is not available
			opts := &StorageOptions{ForceBackend: tt.backend}
			_, err := GetStorage(opts)

			// Skip test if keychain is not available
			if err != nil && tt.backend == BackendKeychain {
				t.Skip("Keychain not available in test environment")
			}

			if err != nil {
				t.Fatalf("Failed to get storage: %v", err)
			}

			result := IsUsingSecureStorage()
			if result != tt.expected {
				t.Errorf("Expected IsUsingSecureStorage()=%v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGetStorageDescription(t *testing.T) {
	ResetStorage()

	// Force file backend
	opts := &StorageOptions{ForceBackend: BackendFile}
	_, err := GetStorage(opts)
	if err != nil {
		t.Fatalf("Failed to get storage: %v", err)
	}

	desc := GetStorageDescription()
	if desc == "" {
		t.Error("Expected non-empty storage description")
	}
	if desc == "unknown" {
		t.Error("Expected valid storage description, got 'unknown'")
	}

	// Description should not include "(secure)" for file backend
	// Note: This is a simple check - could be more sophisticated
	// if desc contains "(secure)" && GetActiveBackend() == BackendFile {
	// 	t.Error("File backend should not be marked as secure")
	// }
}

func TestResetStorage(t *testing.T) {
	// Get storage to initialize state
	_, err := GetStorage(&StorageOptions{ForceBackend: BackendFile})
	if err != nil {
		t.Fatalf("Failed to get storage: %v", err)
	}

	// Verify state is initialized
	if GetActiveBackend() == "" {
		t.Error("Expected active backend to be set")
	}

	// Reset storage
	ResetStorage()

	// Verify state is cleared
	if GetActiveBackend() != "" {
		t.Errorf("Expected empty active backend after reset, got %v", GetActiveBackend())
	}
}

func TestGetStorageDescription_Keychain(t *testing.T) {
	// Skip if keychain is not available
	if !IsKeychainAvailable() {
		t.Skip("Keychain not available in test environment")
	}

	ResetStorage()

	// Force keychain backend
	opts := &StorageOptions{ForceBackend: BackendKeychain}
	_, err := GetStorage(opts)
	if err != nil {
		t.Fatalf("Failed to get storage: %v", err)
	}

	desc := GetStorageDescription()
	if desc == "" {
		t.Error("Expected non-empty storage description")
	}

	// Keychain description should include "(secure)"
	if !contains(desc, "(secure)") {
		t.Errorf("Keychain description should contain '(secure)', got: %s", desc)
	}
}

func TestGetStorageDescription_File(t *testing.T) {
	ResetStorage()

	// Force file backend
	opts := &StorageOptions{ForceBackend: BackendFile}
	_, err := GetStorage(opts)
	if err != nil {
		t.Fatalf("Failed to get storage: %v", err)
	}

	desc := GetStorageDescription()
	if desc == "" {
		t.Error("Expected non-empty storage description")
	}

	// File description should NOT include "(secure)"
	if contains(desc, "(secure)") {
		t.Errorf("File description should not contain '(secure)', got: %s", desc)
	}
}

func TestDetectBackend_AutoFallback(t *testing.T) {
	// This tests the auto-detect fallback warning path
	ResetStorage()

	// Clear environment
	_ = os.Unsetenv(StorageEnvVar)

	// Get storage with auto-detect
	storage, err := GetStorage(nil)
	if err != nil {
		t.Fatalf("GetStorage failed: %v", err)
	}

	// Should get a valid backend (either keychain or file)
	backend := storage.GetBackendType()
	if backend != BackendKeychain && backend != BackendFile {
		t.Errorf("Expected valid backend, got %v", backend)
	}
}

func TestDetectBackend_InvalidEnvValue(t *testing.T) {
	ResetStorage()

	// Set invalid environment value
	t.Setenv(StorageEnvVar, "invalid-backend")

	_, err := GetStorage(nil)
	if err == nil {
		t.Fatal("Expected error for invalid DD_TOKEN_STORAGE value")
	}

	if !contains(err.Error(), "invalid DD_TOKEN_STORAGE value") {
		t.Errorf("Error message should mention invalid value, got: %v", err)
	}
}

func TestCreateStorage_UnknownBackend(t *testing.T) {
	ResetStorage()

	// Try to create storage with unknown backend
	_, err := createStorage("unknown-backend")
	if err == nil {
		t.Fatal("Expected error for unknown backend type")
	}

	if !contains(err.Error(), "unknown backend type") {
		t.Errorf("Error message should mention unknown backend, got: %v", err)
	}
}

func TestGetStorage_CachedInstance(t *testing.T) {
	ResetStorage()

	// Get storage first time
	storage1, err := GetStorage(&StorageOptions{ForceBackend: BackendFile})
	if err != nil {
		t.Fatalf("First GetStorage failed: %v", err)
	}

	// Get storage second time - should return cached instance
	storage2, err := GetStorage(nil)
	if err != nil {
		t.Fatalf("Second GetStorage failed: %v", err)
	}

	// Should be the same instance
	if storage1 != storage2 {
		t.Error("GetStorage should return cached instance")
	}
}

func TestGetStorage_ForceBackendOverridesCache(t *testing.T) {
	ResetStorage()

	// Get file storage first
	storage1, err := GetStorage(&StorageOptions{ForceBackend: BackendFile})
	if err != nil {
		t.Fatalf("First GetStorage failed: %v", err)
	}

	if storage1.GetBackendType() != BackendFile {
		t.Errorf("Expected file backend, got %v", storage1.GetBackendType())
	}

	// Force keychain backend (if available)
	if IsKeychainAvailable() {
		storage2, err := GetStorage(&StorageOptions{ForceBackend: BackendKeychain})
		if err != nil {
			t.Fatalf("Second GetStorage failed: %v", err)
		}

		// Should get keychain backend, not cached file backend
		if storage2.GetBackendType() != BackendKeychain {
			t.Errorf("Expected keychain backend, got %v", storage2.GetBackendType())
		}

		// Should be a different instance
		if storage1 == storage2 {
			t.Error("ForceBackend should override cached instance")
		}
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && hasSubstring(s, substr)))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
