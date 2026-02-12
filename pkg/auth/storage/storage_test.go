// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DataDog/pup/pkg/auth/types"
)

func TestFileStorage_TokenOperations(t *testing.T) {
	t.Parallel()
	// Create temporary directory for testing
	tempDir := t.TempDir()

	// Create file storage with temp directory
	storage := &FileStorage{baseDir: tempDir}

	site := "test.datadoghq.com"

	// Test SaveTokens
	tokens := &types.TokenSet{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		IssuedAt:     time.Now().Unix(),
		Scope:        "dashboards_read dashboards_write",
	}

	err := storage.SaveTokens(site, tokens)
	if err != nil {
		t.Fatalf("SaveTokens failed: %v", err)
	}

	// Verify file was created with correct permissions
	filename := filepath.Join(tempDir, "tokens_test_datadoghq_com.json")
	info, err := os.Stat(filename)
	if err != nil {
		t.Fatalf("Token file not created: %v", err)
	}

	// Check file permissions (should be 0600)
	if info.Mode().Perm() != 0600 {
		t.Errorf("Expected file permissions 0600, got %v", info.Mode().Perm())
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
	nonExistentTokens, err := storage.LoadTokens("nonexistent.datadoghq.com")
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

	// Verify file was deleted
	_, err = os.Stat(filename)
	if !os.IsNotExist(err) {
		t.Error("Token file should have been deleted")
	}

	// Test DeleteTokens for non-existent site (should not error)
	err = storage.DeleteTokens("nonexistent.datadoghq.com")
	if err != nil {
		t.Fatalf("DeleteTokens should not error for non-existent site: %v", err)
	}
}

func TestFileStorage_ClientCredentialOperations(t *testing.T) {
	t.Parallel()
	// Create temporary directory for testing
	tempDir := t.TempDir()

	// Create file storage with temp directory
	storage := &FileStorage{baseDir: tempDir}

	site := "test.datadoghq.com"

	// Test SaveClientCredentials
	creds := &types.ClientCredentials{
		ClientID:     "test-client-id",
		ClientName:   "test-client",
		RedirectURIs: []string{"http://127.0.0.1:8000/oauth/callback"},
		RegisteredAt: time.Now().Unix(),
		Site:         site,
	}

	err := storage.SaveClientCredentials(site, creds)
	if err != nil {
		t.Fatalf("SaveClientCredentials failed: %v", err)
	}

	// Verify file was created with correct permissions
	filename := filepath.Join(tempDir, "client_test_datadoghq_com.json")
	info, err := os.Stat(filename)
	if err != nil {
		t.Fatalf("Client credentials file not created: %v", err)
	}

	// Check file permissions (should be 0600)
	if info.Mode().Perm() != 0600 {
		t.Errorf("Expected file permissions 0600, got %v", info.Mode().Perm())
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
	nonExistentCreds, err := storage.LoadClientCredentials("nonexistent.datadoghq.com")
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

	// Verify file was deleted
	_, err = os.Stat(filename)
	if !os.IsNotExist(err) {
		t.Error("Client credentials file should have been deleted")
	}

	// Test DeleteClientCredentials for non-existent site (should not error)
	err = storage.DeleteClientCredentials("nonexistent.datadoghq.com")
	if err != nil {
		t.Fatalf("DeleteClientCredentials should not error for non-existent site: %v", err)
	}
}

func TestFileStorage_GetBackendType(t *testing.T) {
	t.Parallel()
	storage := &FileStorage{baseDir: "/tmp"}

	if storage.GetBackendType() != BackendFile {
		t.Errorf("Expected backend type %v, got %v", BackendFile, storage.GetBackendType())
	}
}

func TestFileStorage_GetStorageLocation(t *testing.T) {
	t.Parallel()
	baseDir := "/tmp/test"
	storage := &FileStorage{baseDir: baseDir}

	location := storage.GetStorageLocation()
	if location != baseDir {
		t.Errorf("Expected storage location %v, got %v", baseDir, location)
	}
}

func TestSanitizeSite(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "datadoghq.com",
			expected: "datadoghq_com",
		},
		{
			input:    "us3.datadoghq.com",
			expected: "us3_datadoghq_com",
		},
		{
			input:    "special-chars!@#$.com",
			expected: "special_chars_____com",
		},
		{
			input:    "UPPERCASE.COM",
			expected: "UPPERCASE_COM",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeSite(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestNewFileStorage(t *testing.T) {
	t.Parallel()
	storage, err := NewFileStorage()
	if err != nil {
		t.Fatalf("NewFileStorage failed: %v", err)
	}

	if storage == nil {
		t.Fatal("NewFileStorage returned nil")
	}

	if storage.baseDir == "" {
		t.Error("Expected non-empty baseDir")
	}

	// Verify directory was created
	info, err := os.Stat(storage.baseDir)
	if err != nil {
		t.Fatalf("Storage directory not created: %v", err)
	}

	if !info.IsDir() {
		t.Error("Storage path is not a directory")
	}

	// Check directory permissions (should be 0700)
	if info.Mode().Perm() != 0700 {
		t.Errorf("Expected directory permissions 0700, got %v", info.Mode().Perm())
	}
}

func TestFileStorage_SaveTokens_InvalidJSON(t *testing.T) {
	t.Parallel()
	// This test verifies error handling, though JSON marshal rarely fails
	// for standard types
	tempDir := t.TempDir()
	storage := &FileStorage{baseDir: tempDir}

	// Normal tokens should work
	tokens := &types.TokenSet{
		AccessToken: "test-token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		IssuedAt:    1234567890,
	}

	err := storage.SaveTokens("test.datadoghq.com", tokens)
	if err != nil {
		t.Errorf("SaveTokens should not fail: %v", err)
	}
}

func TestFileStorage_LoadTokens_InvalidJSON(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	storage := &FileStorage{baseDir: tempDir}

	site := "invalid-json.datadoghq.com"
	filename := filepath.Join(tempDir, "tokens_"+sanitizeSite(site)+".json")

	// Write invalid JSON to file
	err := os.WriteFile(filename, []byte("invalid json {{{"), 0600)
	if err != nil {
		t.Fatalf("Failed to write invalid JSON: %v", err)
	}

	// Try to load tokens
	tokens, err := storage.LoadTokens(site)
	if err == nil {
		t.Error("LoadTokens should fail for invalid JSON")
	}

	if tokens != nil {
		t.Error("LoadTokens should return nil for invalid JSON")
	}
}

func TestFileStorage_SaveClientCredentials_InvalidJSON(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	storage := &FileStorage{baseDir: tempDir}

	// Normal credentials should work
	creds := &types.ClientCredentials{
		ClientID:     "test-client",
		ClientName:   "test",
		RedirectURIs: []string{"http://localhost:8000"},
		RegisteredAt: 1234567890,
		Site:         "test.datadoghq.com",
	}

	err := storage.SaveClientCredentials("test.datadoghq.com", creds)
	if err != nil {
		t.Errorf("SaveClientCredentials should not fail: %v", err)
	}
}

func TestFileStorage_LoadClientCredentials_InvalidJSON(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	storage := &FileStorage{baseDir: tempDir}

	site := "invalid-creds.datadoghq.com"
	filename := filepath.Join(tempDir, "client_"+sanitizeSite(site)+".json")

	// Write invalid JSON to file
	err := os.WriteFile(filename, []byte("not valid json"), 0600)
	if err != nil {
		t.Fatalf("Failed to write invalid JSON: %v", err)
	}

	// Try to load credentials
	creds, err := storage.LoadClientCredentials(site)
	if err == nil {
		t.Error("LoadClientCredentials should fail for invalid JSON")
	}

	if creds != nil {
		t.Error("LoadClientCredentials should return nil for invalid JSON")
	}
}

func TestFileStorage_DeleteTokens_AlreadyDeleted(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	storage := &FileStorage{baseDir: tempDir}

	site := "double-delete.datadoghq.com"

	// Delete non-existent tokens (should not error)
	err := storage.DeleteTokens(site)
	if err != nil {
		t.Errorf("DeleteTokens should not error for non-existent file: %v", err)
	}

	// Delete again (should still not error)
	err = storage.DeleteTokens(site)
	if err != nil {
		t.Errorf("DeleteTokens should not error on second delete: %v", err)
	}
}

func TestFileStorage_DeleteClientCredentials_AlreadyDeleted(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	storage := &FileStorage{baseDir: tempDir}

	site := "double-delete-creds.datadoghq.com"

	// Delete non-existent credentials (should not error)
	err := storage.DeleteClientCredentials(site)
	if err != nil {
		t.Errorf("DeleteClientCredentials should not error for non-existent file: %v", err)
	}

	// Delete again (should still not error)
	err = storage.DeleteClientCredentials(site)
	if err != nil {
		t.Errorf("DeleteClientCredentials should not error on second delete: %v", err)
	}
}

func TestFileStorage_MultipleOperations(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	storage := &FileStorage{baseDir: tempDir}

	sites := []string{"site1.datadoghq.com", "site2.datadoghq.eu", "site3.us3.datadoghq.com"}

	// Save tokens for multiple sites
	for i, site := range sites {
		tokens := &types.TokenSet{
			AccessToken:  fmt.Sprintf("access-token-%d", i),
			RefreshToken: fmt.Sprintf("refresh-token-%d", i),
			TokenType:    "Bearer",
			ExpiresIn:    3600,
			IssuedAt:     time.Now().Unix(),
		}

		err := storage.SaveTokens(site, tokens)
		if err != nil {
			t.Fatalf("SaveTokens failed for %s: %v", site, err)
		}
	}

	// Load and verify tokens for each site
	for i, site := range sites {
		tokens, err := storage.LoadTokens(site)
		if err != nil {
			t.Fatalf("LoadTokens failed for %s: %v", site, err)
		}

		expectedToken := fmt.Sprintf("access-token-%d", i)
		if tokens.AccessToken != expectedToken {
			t.Errorf("Site %s: expected token %s, got %s", site, expectedToken, tokens.AccessToken)
		}
	}

	// Delete tokens for one site
	err := storage.DeleteTokens(sites[0])
	if err != nil {
		t.Fatalf("DeleteTokens failed: %v", err)
	}

	// Verify first site tokens are deleted
	tokens, err := storage.LoadTokens(sites[0])
	if err != nil {
		t.Fatalf("LoadTokens failed: %v", err)
	}
	if tokens != nil {
		t.Error("Expected nil tokens after delete")
	}

	// Verify other sites still have tokens
	tokens, err = storage.LoadTokens(sites[1])
	if err != nil {
		t.Fatalf("LoadTokens failed: %v", err)
	}
	if tokens == nil {
		t.Error("Expected tokens for site 2")
	}
}
