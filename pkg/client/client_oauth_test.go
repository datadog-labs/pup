// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

//go:build !js

package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/datadog-labs/pup/pkg/auth/dcr"
	"github.com/datadog-labs/pup/pkg/auth/storage"
	"github.com/datadog-labs/pup/pkg/auth/types"
	"github.com/datadog-labs/pup/pkg/config"
)

// setupTestStorage creates a FileStorage backed by a temp dir and seeds it with
// the provided tokens and client credentials. It also installs the test hooks
// on getStorageFunc so that NewWithOptions uses this storage instead of the
// real keychain/file singleton. The returned cleanup function restores the
// original hooks.
func setupTestStorage(t *testing.T, site string, tokens *types.TokenSet, creds *types.ClientCredentials) func() {
	t.Helper()

	mock := &mockStorage{
		tokens: make(map[string]*types.TokenSet),
		creds:  make(map[string]*types.ClientCredentials),
	}
	if tokens != nil {
		mock.tokens[site] = tokens
	}
	if creds != nil {
		mock.creds[site] = creds
	}

	origGetStorage := getStorageFunc
	origNewDCR := newDCRClientFunc

	getStorageFunc = func() (storage.Storage, error) { return mock, nil }

	return func() {
		getStorageFunc = origGetStorage
		newDCRClientFunc = origNewDCR
	}
}

// mockStorage implements storage.Storage in-memory for tests.
type mockStorage struct {
	tokens map[string]*types.TokenSet
	creds  map[string]*types.ClientCredentials
}

func (m *mockStorage) GetBackendType() storage.BackendType { return storage.BackendFile }
func (m *mockStorage) GetStorageLocation() string          { return "test" }
func (m *mockStorage) DeleteTokens(site string) error      { delete(m.tokens, site); return nil }
func (m *mockStorage) DeleteClientCredentials(site string) error {
	delete(m.creds, site)
	return nil
}

func (m *mockStorage) SaveTokens(site string, tokens *types.TokenSet) error {
	m.tokens[site] = tokens
	return nil
}

func (m *mockStorage) LoadTokens(site string) (*types.TokenSet, error) {
	return m.tokens[site], nil
}

func (m *mockStorage) SaveClientCredentials(site string, creds *types.ClientCredentials) error {
	m.creds[site] = creds
	return nil
}

func (m *mockStorage) LoadClientCredentials(site string) (*types.ClientCredentials, error) {
	return m.creds[site], nil
}

// mockDCRTransport rewrites requests to point at the test server.
type mockDCRTransport struct {
	server *httptest.Server
}

func (m *mockDCRTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	req.URL.Host = strings.TrimPrefix(m.server.URL, "http://")
	return http.DefaultTransport.RoundTrip(req)
}

func TestNewWithOptions_ValidToken_UsesOAuth(t *testing.T) {
	site := "datadoghq.com"
	tokens := &types.TokenSet{
		AccessToken:  "valid-access-token",
		RefreshToken: "valid-refresh-token",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		IssuedAt:     time.Now().Unix(), // just issued, not expired
	}

	cleanup := setupTestStorage(t, site, tokens, nil)
	defer cleanup()

	client, err := NewWithOptions(&config.Config{Site: site}, false)
	if err != nil {
		t.Fatalf("NewWithOptions() error = %v", err)
	}

	// Should use OAuth, not API keys
	if client.GetAuthType() != AuthTypeOAuth {
		t.Errorf("expected AuthTypeOAuth, got %v", client.GetAuthType())
	}

	token, ok := client.ctx.Value(datadog.ContextAccessToken).(string)
	if !ok || token != "valid-access-token" {
		t.Errorf("expected access token 'valid-access-token', got %q", token)
	}
}

func TestNewWithOptions_ExpiredToken_NoRefreshToken_FallsBackToAPIKeys(t *testing.T) {
	site := "datadoghq.com"
	tokens := &types.TokenSet{
		AccessToken:  "expired-access-token",
		RefreshToken: "", // no refresh token
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		IssuedAt:     time.Now().Add(-2 * time.Hour).Unix(), // expired
	}

	cleanup := setupTestStorage(t, site, tokens, nil)
	defer cleanup()

	client, err := NewWithOptions(&config.Config{
		Site:   site,
		APIKey: "test-api-key",
		AppKey: "test-app-key",
	}, false)
	if err != nil {
		t.Fatalf("NewWithOptions() error = %v", err)
	}

	if client.GetAuthType() != AuthTypeAPIKeys {
		t.Errorf("expected AuthTypeAPIKeys, got %v", client.GetAuthType())
	}
}

func TestNewWithOptions_ExpiredToken_NoClientCreds_FallsBackToAPIKeys(t *testing.T) {
	site := "datadoghq.com"
	tokens := &types.TokenSet{
		AccessToken:  "expired-access-token",
		RefreshToken: "valid-refresh-token",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		IssuedAt:     time.Now().Add(-2 * time.Hour).Unix(), // expired
	}
	// No client credentials stored

	cleanup := setupTestStorage(t, site, tokens, nil)
	defer cleanup()

	client, err := NewWithOptions(&config.Config{
		Site:   site,
		APIKey: "test-api-key",
		AppKey: "test-app-key",
	}, false)
	if err != nil {
		t.Fatalf("NewWithOptions() error = %v", err)
	}

	if client.GetAuthType() != AuthTypeAPIKeys {
		t.Errorf("expected AuthTypeAPIKeys, got %v", client.GetAuthType())
	}
}

func TestNewWithOptions_ExpiredToken_RefreshFails_FallsBackToAPIKeys(t *testing.T) {
	site := "datadoghq.com"
	tokens := &types.TokenSet{
		AccessToken:  "expired-access-token",
		RefreshToken: "bad-refresh-token",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		IssuedAt:     time.Now().Add(-2 * time.Hour).Unix(),
	}
	creds := &types.ClientCredentials{
		ClientID:     "test-client-id",
		ClientName:   "test-client",
		RegisteredAt: time.Now().Unix(),
		Site:         site,
	}

	cleanup := setupTestStorage(t, site, tokens, creds)
	defer cleanup()

	// Mock DCR server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error":             "invalid_grant",
			"error_description": "Refresh token has expired",
		})
	}))
	defer server.Close()

	newDCRClientFunc = func(site string) *dcr.Client {
		return dcr.NewClientWithHTTPClient(site, &http.Client{
			Transport: &mockDCRTransport{server: server},
		})
	}

	client, err := NewWithOptions(&config.Config{
		Site:   site,
		APIKey: "test-api-key",
		AppKey: "test-app-key",
	}, false)
	if err != nil {
		t.Fatalf("NewWithOptions() error = %v", err)
	}

	if client.GetAuthType() != AuthTypeAPIKeys {
		t.Errorf("expected AuthTypeAPIKeys after failed refresh, got %v", client.GetAuthType())
	}
}

func TestNewWithOptions_ExpiredToken_RefreshSucceeds_UsesNewToken(t *testing.T) {
	site := "datadoghq.com"
	tokens := &types.TokenSet{
		AccessToken:  "expired-access-token",
		RefreshToken: "valid-refresh-token",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		IssuedAt:     time.Now().Add(-2 * time.Hour).Unix(),
	}
	creds := &types.ClientCredentials{
		ClientID:     "test-client-id",
		ClientName:   "test-client",
		RegisteredAt: time.Now().Unix(),
		Site:         site,
	}

	cleanup := setupTestStorage(t, site, tokens, creds)
	defer cleanup()

	// Mock DCR server that returns fresh tokens
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("failed to parse form: %v", err)
		}

		// Verify the refresh request
		if r.Form.Get("grant_type") != "refresh_token" {
			t.Errorf("grant_type = %s, want refresh_token", r.Form.Get("grant_type"))
		}
		if r.Form.Get("client_id") != "test-client-id" {
			t.Errorf("client_id = %s, want test-client-id", r.Form.Get("client_id"))
		}
		if r.Form.Get("refresh_token") != "valid-refresh-token" {
			t.Errorf("refresh_token = %s, want valid-refresh-token", r.Form.Get("refresh_token"))
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token":  "fresh-access-token",
			"refresh_token": "fresh-refresh-token",
			"token_type":    "Bearer",
			"expires_in":    3600,
		})
	}))
	defer server.Close()

	newDCRClientFunc = func(site string) *dcr.Client {
		return dcr.NewClientWithHTTPClient(site, &http.Client{
			Transport: &mockDCRTransport{server: server},
		})
	}

	client, err := NewWithOptions(&config.Config{Site: site}, false)
	if err != nil {
		t.Fatalf("NewWithOptions() error = %v", err)
	}

	// Should use OAuth with the NEW token
	if client.GetAuthType() != AuthTypeOAuth {
		t.Fatalf("expected AuthTypeOAuth after refresh, got %v", client.GetAuthType())
	}

	token, ok := client.ctx.Value(datadog.ContextAccessToken).(string)
	if !ok || token != "fresh-access-token" {
		t.Errorf("expected 'fresh-access-token', got %q", token)
	}
}

func TestNewWithOptions_ExpiredToken_RefreshSucceeds_PersistsNewToken(t *testing.T) {
	site := "datadoghq.com"
	tokens := &types.TokenSet{
		AccessToken:  "expired-access-token",
		RefreshToken: "valid-refresh-token",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		IssuedAt:     time.Now().Add(-2 * time.Hour).Unix(),
	}
	creds := &types.ClientCredentials{
		ClientID:     "test-client-id",
		ClientName:   "test-client",
		RegisteredAt: time.Now().Unix(),
		Site:         site,
	}

	cleanup := setupTestStorage(t, site, tokens, creds)
	defer cleanup()

	// Capture the mock storage to inspect saved tokens later
	mock := &mockStorage{
		tokens: map[string]*types.TokenSet{site: tokens},
		creds:  map[string]*types.ClientCredentials{site: creds},
	}
	getStorageFunc = func() (storage.Storage, error) { return mock, nil }

	// Mock DCR server returning fresh tokens
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token":  "persisted-access-token",
			"refresh_token": "persisted-refresh-token",
			"token_type":    "Bearer",
			"expires_in":    3600,
		})
	}))
	defer server.Close()

	newDCRClientFunc = func(site string) *dcr.Client {
		return dcr.NewClientWithHTTPClient(site, &http.Client{
			Transport: &mockDCRTransport{server: server},
		})
	}

	_, err := NewWithOptions(&config.Config{Site: site}, false)
	if err != nil {
		t.Fatalf("NewWithOptions() error = %v", err)
	}

	// Verify the refreshed token was persisted to storage
	saved := mock.tokens[site]
	if saved == nil {
		t.Fatal("expected tokens to be saved to storage")
	}
	if saved.AccessToken != "persisted-access-token" {
		t.Errorf("saved AccessToken = %q, want 'persisted-access-token'", saved.AccessToken)
	}
	if saved.RefreshToken != "persisted-refresh-token" {
		t.Errorf("saved RefreshToken = %q, want 'persisted-refresh-token'", saved.RefreshToken)
	}
}

func TestNewWithOptions_ExpiredToken_NoAPIKeys_ReturnsError(t *testing.T) {
	site := "datadoghq.com"
	tokens := &types.TokenSet{
		AccessToken:  "expired-access-token",
		RefreshToken: "", // no refresh token
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		IssuedAt:     time.Now().Add(-2 * time.Hour).Unix(),
	}

	cleanup := setupTestStorage(t, site, tokens, nil)
	defer cleanup()

	// No API keys either — should get an auth error
	_, err := NewWithOptions(&config.Config{Site: site}, false)
	if err == nil {
		t.Fatal("expected error when token expired and no API keys")
	}
	if !strings.Contains(err.Error(), "authentication required") {
		t.Errorf("expected 'authentication required' error, got: %v", err)
	}
}

func TestNewWithOptions_ForceAPIKeys_SkipsOAuth(t *testing.T) {
	site := "datadoghq.com"
	tokens := &types.TokenSet{
		AccessToken:  "valid-access-token",
		RefreshToken: "valid-refresh-token",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		IssuedAt:     time.Now().Unix(), // not expired
	}

	cleanup := setupTestStorage(t, site, tokens, nil)
	defer cleanup()

	client, err := NewWithOptions(&config.Config{
		Site:   site,
		APIKey: "test-api-key",
		AppKey: "test-app-key",
	}, true) // forceAPIKeys = true
	if err != nil {
		t.Fatalf("NewWithOptions() error = %v", err)
	}

	// Should use API keys even though valid OAuth token exists
	if client.GetAuthType() != AuthTypeAPIKeys {
		t.Errorf("expected AuthTypeAPIKeys with forceAPIKeys=true, got %v", client.GetAuthType())
	}
}
