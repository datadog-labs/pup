// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package storage

import (
	"runtime"
	"testing"
	"time"

	"github.com/DataDog/pup/pkg/auth/types"
)

func TestKeychainStorage_GetBackendType(t *testing.T) {
	// Skip if keychain is not available
	if !IsKeychainAvailable() {
		t.Skip("Keychain not available in test environment")
	}

	storage, err := NewKeychainStorage()
	if err != nil {
		t.Fatalf("NewKeychainStorage failed: %v", err)
	}

	if storage.GetBackendType() != BackendKeychain {
		t.Errorf("Expected backend type %v, got %v", BackendKeychain, storage.GetBackendType())
	}
}

func TestKeychainStorage_GetStorageLocation(t *testing.T) {
	// Skip if keychain is not available
	if !IsKeychainAvailable() {
		t.Skip("Keychain not available in test environment")
	}

	storage, err := NewKeychainStorage()
	if err != nil {
		t.Fatalf("NewKeychainStorage failed: %v", err)
	}

	location := storage.GetStorageLocation()
	if location == "" {
		t.Error("Expected non-empty storage location")
	}

	// Verify location makes sense for the current OS
	switch runtime.GOOS {
	case "darwin":
		if location != "macOS Keychain" {
			t.Errorf("Expected 'macOS Keychain' on darwin, got %v", location)
		}
	case "windows":
		if location != "Windows Credential Manager" {
			t.Errorf("Expected 'Windows Credential Manager' on windows, got %v", location)
		}
	case "linux":
		if location != "System Keychain (Secret Service)" {
			t.Errorf("Expected 'System Keychain (Secret Service)' on linux, got %v", location)
		}
	}
}

func TestKeychainStorage_TokenOperations(t *testing.T) {
	// Skip if keychain is not available
	if !IsKeychainAvailable() {
		t.Skip("Keychain not available in test environment")
	}

	storage, err := NewKeychainStorage()
	if err != nil {
		t.Fatalf("NewKeychainStorage failed: %v", err)
	}

	site := "test-keychain.datadoghq.com"

	// Clean up any existing test data
	_ = storage.DeleteTokens(site)
	defer storage.DeleteTokens(site)

	// Test SaveTokens
	tokens := &types.TokenSet{
		AccessToken:  "test-keychain-access-token",
		RefreshToken: "test-keychain-refresh-token",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		IssuedAt:     time.Now().Unix(),
		Scope:        "dashboards_read dashboards_write",
	}

	err = storage.SaveTokens(site, tokens)
	if err != nil {
		t.Fatalf("SaveTokens failed: %v", err)
	}

	// Test LoadTokens
	loadedTokens, err := storage.LoadTokens(site)
	if err != nil {
		t.Fatalf("LoadTokens failed: %v", err)
	}

	if loadedTokens == nil {
		t.Fatal("LoadTokens returned nil")
	}

	if loadedTokens.AccessToken != tokens.AccessToken {
		t.Errorf("Expected access token %v, got %v", tokens.AccessToken, loadedTokens.AccessToken)
	}

	if loadedTokens.RefreshToken != tokens.RefreshToken {
		t.Errorf("Expected refresh token %v, got %v", tokens.RefreshToken, loadedTokens.RefreshToken)
	}

	// Test LoadTokens for non-existent site
	nonExistentTokens, err := storage.LoadTokens("nonexistent-keychain.datadoghq.com")
	if err != nil {
		t.Fatalf("LoadTokens should not error for non-existent site: %v", err)
	}

	if nonExistentTokens != nil {
		t.Error("Expected nil for non-existent site")
	}

	// Test DeleteTokens
	err = storage.DeleteTokens(site)
	if err != nil {
		t.Fatalf("DeleteTokens failed: %v", err)
	}

	// Verify tokens were deleted
	deletedTokens, err := storage.LoadTokens(site)
	if err != nil {
		t.Fatalf("LoadTokens failed after delete: %v", err)
	}

	if deletedTokens != nil {
		t.Error("Expected nil tokens after delete")
	}

	// Test DeleteTokens for non-existent site (should not error)
	err = storage.DeleteTokens("nonexistent-keychain.datadoghq.com")
	if err != nil {
		t.Fatalf("DeleteTokens should not error for non-existent site: %v", err)
	}
}

func TestKeychainStorage_ClientCredentialOperations(t *testing.T) {
	// Skip if keychain is not available
	if !IsKeychainAvailable() {
		t.Skip("Keychain not available in test environment")
	}

	storage, err := NewKeychainStorage()
	if err != nil {
		t.Fatalf("NewKeychainStorage failed: %v", err)
	}

	site := "test-client-keychain.datadoghq.com"

	// Clean up any existing test data
	_ = storage.DeleteClientCredentials(site)
	defer storage.DeleteClientCredentials(site)

	// Test SaveClientCredentials
	creds := &types.ClientCredentials{
		ClientID:     "test-keychain-client-id",
		ClientName:   "test-client",
		RedirectURIs: []string{"http://127.0.0.1:8000/oauth/callback"},
		RegisteredAt: time.Now().Unix(),
		Site:         site,
	}

	err = storage.SaveClientCredentials(site, creds)
	if err != nil {
		t.Fatalf("SaveClientCredentials failed: %v", err)
	}

	// Test LoadClientCredentials
	loadedCreds, err := storage.LoadClientCredentials(site)
	if err != nil {
		t.Fatalf("LoadClientCredentials failed: %v", err)
	}

	if loadedCreds == nil {
		t.Fatal("LoadClientCredentials returned nil")
	}

	if loadedCreds.ClientID != creds.ClientID {
		t.Errorf("Expected client ID %v, got %v", creds.ClientID, loadedCreds.ClientID)
	}

	if loadedCreds.ClientName != creds.ClientName {
		t.Errorf("Expected client name %v, got %v", creds.ClientName, loadedCreds.ClientName)
	}

	// Test LoadClientCredentials for non-existent site
	nonExistentCreds, err := storage.LoadClientCredentials("nonexistent-client-keychain.datadoghq.com")
	if err != nil {
		t.Fatalf("LoadClientCredentials should not error for non-existent site: %v", err)
	}

	if nonExistentCreds != nil {
		t.Error("Expected nil for non-existent site")
	}

	// Test DeleteClientCredentials
	err = storage.DeleteClientCredentials(site)
	if err != nil {
		t.Fatalf("DeleteClientCredentials failed: %v", err)
	}

	// Verify credentials were deleted
	deletedCreds, err := storage.LoadClientCredentials(site)
	if err != nil {
		t.Fatalf("LoadClientCredentials failed after delete: %v", err)
	}

	if deletedCreds != nil {
		t.Error("Expected nil credentials after delete")
	}

	// Test DeleteClientCredentials for non-existent site (should not error)
	err = storage.DeleteClientCredentials("nonexistent-client-keychain.datadoghq.com")
	if err != nil {
		t.Fatalf("DeleteClientCredentials should not error for non-existent site: %v", err)
	}
}

func TestIsKeychainAvailable(t *testing.T) {
	// Just test that the function doesn't panic
	available := IsKeychainAvailable()

	// Result depends on environment, so we just log it
	t.Logf("Keychain available: %v", available)

	// On macOS, keychain should generally be available
	if runtime.GOOS == "darwin" && !available {
		t.Log("Warning: Keychain not available on macOS (unusual)")
	}
}
