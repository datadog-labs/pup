// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DataDog/pup/pkg/auth/types"
)

func TestFileStorage_TokenOperations(t *testing.T) {
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
	storage := &FileStorage{baseDir: "/tmp"}

	if storage.GetBackendType() != BackendFile {
		t.Errorf("Expected backend type %v, got %v", BackendFile, storage.GetBackendType())
	}
}

func TestFileStorage_GetStorageLocation(t *testing.T) {
	baseDir := "/tmp/test"
	storage := &FileStorage{baseDir: baseDir}

	location := storage.GetStorageLocation()
	if location != baseDir {
		t.Errorf("Expected storage location %v, got %v", baseDir, location)
	}
}

func TestSanitizeSite(t *testing.T) {
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
