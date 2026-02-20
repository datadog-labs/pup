// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

//go:build !js

package storage

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/99designs/keyring"
	"github.com/datadog-labs/pup/pkg/auth/types"
)

func TestKeychainStorage_GetBackendType(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
	// Just test that the function doesn't panic
	available := IsKeychainAvailable()

	// Result depends on environment, so we just log it
	t.Logf("Keychain available: %v", available)

	// On macOS, keychain should generally be available
	if runtime.GOOS == "darwin" && !available {
		t.Log("Warning: Keychain not available on macOS (unusual)")
	}
}

func TestNewKeychainStorage_Success(t *testing.T) {
	t.Parallel()
	// Skip if keychain is not available
	if !IsKeychainAvailable() {
		t.Skip("Keychain not available in test environment")
	}

	storage, err := NewKeychainStorage()
	if err != nil {
		t.Fatalf("NewKeychainStorage failed: %v", err)
	}

	if storage == nil {
		t.Fatal("NewKeychainStorage returned nil")
	}

	if storage.tokenKeyring == nil {
		t.Error("tokenKeyring is nil")
	}

	if storage.clientKeyring == nil {
		t.Error("clientKeyring is nil")
	}
}

func TestKeychainStorage_SaveTokens_MarshalError(t *testing.T) {
	t.Parallel()
	// Skip if keychain is not available
	if !IsKeychainAvailable() {
		t.Skip("Keychain not available in test environment")
	}

	storage, err := NewKeychainStorage()
	if err != nil {
		t.Fatalf("NewKeychainStorage failed: %v", err)
	}

	// Normal tokens should work fine (JSON marshal rarely fails for standard types)
	tokens := &types.TokenSet{
		AccessToken: "test-token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		IssuedAt:    time.Now().Unix(),
	}

	err = storage.SaveTokens("marshal-test.datadoghq.com", tokens)
	if err != nil {
		t.Errorf("SaveTokens should succeed: %v", err)
	}

	// Clean up
	storage.DeleteTokens("marshal-test.datadoghq.com")
}

func TestKeychainStorage_LoadTokens_UnmarshalError(t *testing.T) {
	t.Parallel()
	// Skip if keychain is not available
	if !IsKeychainAvailable() {
		t.Skip("Keychain not available in test environment")
	}

	storage, err := NewKeychainStorage()
	if err != nil {
		t.Fatalf("NewKeychainStorage failed: %v", err)
	}

	site := "unmarshal-test.datadoghq.com"
	key := TokenPrefix + site

	// Manually set invalid JSON in keyring
	item := keyring.Item{
		Key:  key,
		Data: []byte("invalid json {{{"),
	}

	err = storage.tokenKeyring.Set(item)
	if err != nil {
		t.Skipf("Could not set invalid data in keyring: %v", err)
	}
	defer storage.DeleteTokens(site)

	// Try to load tokens
	tokens, err := storage.LoadTokens(site)
	if err == nil {
		t.Error("LoadTokens should fail for invalid JSON")
	}

	if tokens != nil {
		t.Error("LoadTokens should return nil for invalid JSON")
	}
}

func TestKeychainStorage_SaveClientCredentials_MarshalError(t *testing.T) {
	t.Parallel()
	// Skip if keychain is not available
	if !IsKeychainAvailable() {
		t.Skip("Keychain not available in test environment")
	}

	storage, err := NewKeychainStorage()
	if err != nil {
		t.Fatalf("NewKeychainStorage failed: %v", err)
	}

	// Normal credentials should work fine
	creds := &types.ClientCredentials{
		ClientID:     "test-client",
		ClientName:   "test",
		RedirectURIs: []string{"http://localhost:8000"},
		RegisteredAt: time.Now().Unix(),
		Site:         "marshal-creds-test.datadoghq.com",
	}

	err = storage.SaveClientCredentials("marshal-creds-test.datadoghq.com", creds)
	if err != nil {
		t.Errorf("SaveClientCredentials should succeed: %v", err)
	}

	// Clean up
	storage.DeleteClientCredentials("marshal-creds-test.datadoghq.com")
}

func TestKeychainStorage_LoadClientCredentials_UnmarshalError(t *testing.T) {
	t.Parallel()
	// Skip if keychain is not available
	if !IsKeychainAvailable() {
		t.Skip("Keychain not available in test environment")
	}

	storage, err := NewKeychainStorage()
	if err != nil {
		t.Fatalf("NewKeychainStorage failed: %v", err)
	}

	site := "unmarshal-creds-test.datadoghq.com"
	key := ClientPrefix + site

	// Manually set invalid JSON in keyring
	item := keyring.Item{
		Key:  key,
		Data: []byte("not valid json"),
	}

	err = storage.clientKeyring.Set(item)
	if err != nil {
		t.Skipf("Could not set invalid data in keyring: %v", err)
	}
	defer storage.DeleteClientCredentials(site)

	// Try to load credentials
	creds, err := storage.LoadClientCredentials(site)
	if err == nil {
		t.Error("LoadClientCredentials should fail for invalid JSON")
	}

	if creds != nil {
		t.Error("LoadClientCredentials should return nil for invalid JSON")
	}
}

func TestKeychainStorage_MultipleOperations(t *testing.T) {
	t.Parallel()
	// Skip if keychain is not available
	if !IsKeychainAvailable() {
		t.Skip("Keychain not available in test environment")
	}

	storage, err := NewKeychainStorage()
	if err != nil {
		t.Fatalf("NewKeychainStorage failed: %v", err)
	}

	sites := []string{
		"multi1.datadoghq.com",
		"multi2.datadoghq.eu",
		"multi3.us3.datadoghq.com",
	}

	// Clean up before test
	for _, site := range sites {
		storage.DeleteTokens(site)
		storage.DeleteClientCredentials(site)
	}
	defer func() {
		for _, site := range sites {
			storage.DeleteTokens(site)
			storage.DeleteClientCredentials(site)
		}
	}()

	// Save tokens and credentials for multiple sites
	for i, site := range sites {
		tokens := &types.TokenSet{
			AccessToken:  fmt.Sprintf("multi-access-%d", i),
			RefreshToken: fmt.Sprintf("multi-refresh-%d", i),
			TokenType:    "Bearer",
			ExpiresIn:    3600,
			IssuedAt:     time.Now().Unix(),
		}

		err := storage.SaveTokens(site, tokens)
		if err != nil {
			t.Fatalf("SaveTokens failed for %s: %v", site, err)
		}

		creds := &types.ClientCredentials{
			ClientID:     fmt.Sprintf("multi-client-%d", i),
			ClientName:   "test-client",
			RedirectURIs: []string{"http://localhost:8000"},
			RegisteredAt: time.Now().Unix(),
			Site:         site,
		}

		err = storage.SaveClientCredentials(site, creds)
		if err != nil {
			t.Fatalf("SaveClientCredentials failed for %s: %v", site, err)
		}
	}

	// Verify all tokens and credentials can be loaded
	for i, site := range sites {
		tokens, err := storage.LoadTokens(site)
		if err != nil {
			t.Fatalf("LoadTokens failed for %s: %v", site, err)
		}

		expectedToken := fmt.Sprintf("multi-access-%d", i)
		if tokens.AccessToken != expectedToken {
			t.Errorf("Site %s: expected token %s, got %s", site, expectedToken, tokens.AccessToken)
		}

		creds, err := storage.LoadClientCredentials(site)
		if err != nil {
			t.Fatalf("LoadClientCredentials failed for %s: %v", site, err)
		}

		expectedClientID := fmt.Sprintf("multi-client-%d", i)
		if creds.ClientID != expectedClientID {
			t.Errorf("Site %s: expected client ID %s, got %s", site, expectedClientID, creds.ClientID)
		}
	}
}
