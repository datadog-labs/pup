// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package dcr

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DataDog/pup/pkg/auth/types"
)

// mockTransport implements http.RoundTripper to redirect requests to test server
type mockTransport struct {
	server *httptest.Server
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Rewrite the URL to point to the test server
	// Original: https://api.datadoghq.com/path
	// Rewritten: http://testserver/path
	req.URL.Scheme = "http"
	req.URL.Host = strings.TrimPrefix(m.server.URL, "http://")
	return http.DefaultTransport.RoundTrip(req)
}

func TestNewClient(t *testing.T) {
	client := NewClient("datadoghq.com")
	if client == nil {
		t.Fatal("NewClient() returned nil")
	}

	if client.site != "datadoghq.com" {
		t.Errorf("site = %s, want datadoghq.com", client.site)
	}

	if client.httpClient == nil {
		t.Error("httpClient is nil")
	}
}

func TestGetRedirectURIs(t *testing.T) {
	uris := GetRedirectURIs()

	expectedURIs := []string{
		"http://127.0.0.1:8000/oauth/callback",
		"http://127.0.0.1:8080/oauth/callback",
		"http://127.0.0.1:8888/oauth/callback",
		"http://127.0.0.1:9000/oauth/callback",
	}

	if len(uris) != len(expectedURIs) {
		t.Fatalf("GetRedirectURIs() returned %d URIs, want %d", len(uris), len(expectedURIs))
	}

	for i, expected := range expectedURIs {
		if uris[i] != expected {
			t.Errorf("URI[%d] = %s, want %s", i, uris[i], expected)
		}
	}
}

func TestClient_Register_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		if r.Method != "POST" {
			t.Errorf("Method = %s, want POST", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/api/v2/oauth2/register") {
			t.Errorf("Path = %s, want /api/v2/oauth2/register", r.URL.Path)
		}

		// Verify headers
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %s, want application/json", r.Header.Get("Content-Type"))
		}

		// Parse request body
		var req RegistrationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		// Verify request fields
		if req.ClientName != DCRClientName {
			t.Errorf("ClientName = %s, want %s", req.ClientName, DCRClientName)
		}
		if len(req.RedirectURIs) != 4 {
			t.Errorf("RedirectURIs count = %d, want 4", len(req.RedirectURIs))
		}
		if len(req.GrantTypes) != 2 {
			t.Errorf("GrantTypes count = %d, want 2", len(req.GrantTypes))
		}

		// Send successful response
		resp := RegistrationResponse{
			ClientID:                "test-client-id",
			ClientName:              DCRClientName,
			RedirectURIs:            GetRedirectURIs(),
			TokenEndpointAuthMethod: "none",
			GrantTypes:              []string{"authorization_code", "refresh_token"},
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client with custom HTTP client that redirects to test server
	httpClient := &http.Client{
		Transport: &mockTransport{server: server},
	}
	client := NewClientWithHTTPClient("datadoghq.com", httpClient)

	// Test registration
	creds, err := client.Register("http://127.0.0.1:8000/oauth/callback", []string{"test_scope"})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if creds.ClientID != "test-client-id" {
		t.Errorf("ClientID = %s, want test-client-id", creds.ClientID)
	}
	if creds.ClientName != DCRClientName {
		t.Errorf("ClientName = %s, want %s", creds.ClientName, DCRClientName)
	}
	if len(creds.RedirectURIs) != 4 {
		t.Errorf("RedirectURIs count = %d, want 4", len(creds.RedirectURIs))
	}
	if creds.Site == "" {
		t.Error("Site is empty")
	}
	if creds.RegisteredAt == 0 {
		t.Error("RegisteredAt is 0")
	}
}

func TestClient_Register_HTTPError(t *testing.T) {
	// Create mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid request"))
	}))
	defer server.Close()

	httpClient := &http.Client{
		Transport: &mockTransport{server: server},
	}
	client := NewClientWithHTTPClient("datadoghq.com", httpClient)

	_, err := client.Register("http://127.0.0.1:8000/oauth/callback", []string{"test_scope"})
	if err == nil {
		t.Error("Register() expected error but got none")
	}
	if !strings.Contains(err.Error(), "DCR failed") {
		t.Errorf("Error = %v, want DCR failed error", err)
	}
}

func TestClient_Register_OAuthError(t *testing.T) {
	// Create mock server that returns OAuth error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		oauthErr := types.OAuthError{
			Error:            "invalid_client",
			ErrorDescription: "Client registration failed",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(oauthErr)
	}))
	defer server.Close()

	httpClient := &http.Client{
		Transport: &mockTransport{server: server},
	}
	client := NewClientWithHTTPClient("datadoghq.com", httpClient)

	_, err := client.Register("http://127.0.0.1:8000/oauth/callback", []string{"test_scope"})
	if err == nil {
		t.Error("Register() expected error but got none")
	}
	if !strings.Contains(err.Error(), "Client registration failed") {
		t.Errorf("Error = %v, want OAuth error with description", err)
	}
}

func TestClient_ExchangeCode_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		if r.Method != "POST" {
			t.Errorf("Method = %s, want POST", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/oauth2/v1/token") {
			t.Errorf("Path = %s, want /oauth2/v1/token", r.URL.Path)
		}

		// Verify Content-Type is form-encoded
		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("Content-Type = %s, want application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
		}

		// Parse form data
		if err := r.ParseForm(); err != nil {
			t.Errorf("Failed to parse form: %v", err)
		}

		// Verify form fields
		if r.Form.Get("grant_type") != "authorization_code" {
			t.Errorf("grant_type = %s, want authorization_code", r.Form.Get("grant_type"))
		}
		if r.Form.Get("client_id") != "test-client-id" {
			t.Errorf("client_id = %s, want test-client-id", r.Form.Get("client_id"))
		}
		if r.Form.Get("code") != "test-code" {
			t.Errorf("code = %s, want test-code", r.Form.Get("code"))
		}
		if r.Form.Get("code_verifier") != "test-verifier" {
			t.Errorf("code_verifier = %s, want test-verifier", r.Form.Get("code_verifier"))
		}
		// Verify no client_secret (public client)
		if r.Form.Get("client_secret") != "" {
			t.Error("client_secret should not be sent for public clients")
		}

		// Send successful response
		resp := TokenResponse{
			AccessToken:  "test-access-token",
			TokenType:    "Bearer",
			ExpiresIn:    3600,
			RefreshToken: "test-refresh-token",
			Scope:        "test_scope",
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	httpClient := &http.Client{
		Transport: &mockTransport{server: server},
	}
	client := NewClientWithHTTPClient("datadoghq.com", httpClient)

	creds := &types.ClientCredentials{
		ClientID: "test-client-id",
		Site:     "datadoghq.com",
	}

	tokens, err := client.ExchangeCode("test-code", "http://127.0.0.1:8000/oauth/callback", "test-verifier", creds)
	if err != nil {
		t.Fatalf("ExchangeCode() error = %v", err)
	}

	if tokens.AccessToken != "test-access-token" {
		t.Errorf("AccessToken = %s, want test-access-token", tokens.AccessToken)
	}
	if tokens.TokenType != "Bearer" {
		t.Errorf("TokenType = %s, want Bearer", tokens.TokenType)
	}
	if tokens.ExpiresIn != 3600 {
		t.Errorf("ExpiresIn = %d, want 3600", tokens.ExpiresIn)
	}
	if tokens.RefreshToken != "test-refresh-token" {
		t.Errorf("RefreshToken = %s, want test-refresh-token", tokens.RefreshToken)
	}
	if tokens.IssuedAt == 0 {
		t.Error("IssuedAt is 0")
	}
}

func TestClient_ExchangeCode_Error(t *testing.T) {
	// Create mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		oauthErr := types.OAuthError{
			Error:            "invalid_grant",
			ErrorDescription: "Authorization code has expired",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(oauthErr)
	}))
	defer server.Close()

	httpClient := &http.Client{
		Transport: &mockTransport{server: server},
	}
	client := NewClientWithHTTPClient("datadoghq.com", httpClient)

	creds := &types.ClientCredentials{
		ClientID: "test-client-id",
		Site:     "datadoghq.com",
	}

	_, err := client.ExchangeCode("invalid-code", "http://127.0.0.1:8000/oauth/callback", "test-verifier", creds)
	if err == nil {
		t.Error("ExchangeCode() expected error but got none")
	}
	if !strings.Contains(err.Error(), "Authorization code has expired") {
		t.Errorf("Error = %v, want OAuth error with description", err)
	}
}

func TestClient_RefreshToken_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != "POST" {
			t.Errorf("Method = %s, want POST", r.Method)
		}

		// Verify Content-Type is form-encoded
		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("Content-Type = %s, want application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
		}

		// Parse form data
		if err := r.ParseForm(); err != nil {
			t.Errorf("Failed to parse form: %v", err)
		}

		// Verify form fields
		if r.Form.Get("grant_type") != "refresh_token" {
			t.Errorf("grant_type = %s, want refresh_token", r.Form.Get("grant_type"))
		}
		if r.Form.Get("client_id") != "test-client-id" {
			t.Errorf("client_id = %s, want test-client-id", r.Form.Get("client_id"))
		}
		if r.Form.Get("refresh_token") != "test-refresh-token" {
			t.Errorf("refresh_token = %s, want test-refresh-token", r.Form.Get("refresh_token"))
		}
		// Verify no client_secret (public client)
		if r.Form.Get("client_secret") != "" {
			t.Error("client_secret should not be sent for public clients")
		}

		// Send successful response
		resp := TokenResponse{
			AccessToken:  "new-access-token",
			TokenType:    "Bearer",
			ExpiresIn:    3600,
			RefreshToken: "new-refresh-token",
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	httpClient := &http.Client{
		Transport: &mockTransport{server: server},
	}
	client := NewClientWithHTTPClient("datadoghq.com", httpClient)

	creds := &types.ClientCredentials{
		ClientID: "test-client-id",
		Site:     "datadoghq.com",
	}

	tokens, err := client.RefreshToken("test-refresh-token", creds)
	if err != nil {
		t.Fatalf("RefreshToken() error = %v", err)
	}

	if tokens.AccessToken != "new-access-token" {
		t.Errorf("AccessToken = %s, want new-access-token", tokens.AccessToken)
	}
	if tokens.RefreshToken != "new-refresh-token" {
		t.Errorf("RefreshToken = %s, want new-refresh-token", tokens.RefreshToken)
	}
}

func TestClient_RefreshToken_Error(t *testing.T) {
	// Create mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		oauthErr := types.OAuthError{
			Error:            "invalid_grant",
			ErrorDescription: "Refresh token has expired",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(oauthErr)
	}))
	defer server.Close()

	httpClient := &http.Client{
		Transport: &mockTransport{server: server},
	}
	client := NewClientWithHTTPClient("datadoghq.com", httpClient)

	creds := &types.ClientCredentials{
		ClientID: "test-client-id",
		Site:     "datadoghq.com",
	}

	_, err := client.RefreshToken("invalid-refresh-token", creds)
	if err == nil {
		t.Error("RefreshToken() expected error but got none")
	}
	if !strings.Contains(err.Error(), "Refresh token has expired") {
		t.Errorf("Error = %v, want OAuth error with description", err)
	}
}

func TestClient_RequestTokens_NetworkError(t *testing.T) {
	// Create client with invalid URL to simulate network error
	client := NewClient("invalid.example.com")

	creds := &types.ClientCredentials{
		ClientID: "test-client-id",
		Site:     "invalid.example.com",
	}

	_, err := client.ExchangeCode("test-code", "http://127.0.0.1:8000/oauth/callback", "test-verifier", creds)
	if err == nil {
		t.Error("ExchangeCode() expected network error but got none")
	}
}

func TestClient_RequestTokens_InvalidJSON(t *testing.T) {
	// Create mock server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	httpClient := &http.Client{
		Transport: &mockTransport{server: server},
	}
	client := NewClientWithHTTPClient("datadoghq.com", httpClient)

	creds := &types.ClientCredentials{
		ClientID: "test-client-id",
		Site:     "datadoghq.com",
	}

	_, err := client.ExchangeCode("test-code", "http://127.0.0.1:8000/oauth/callback", "test-verifier", creds)
	if err == nil {
		t.Error("ExchangeCode() expected JSON parse error but got none")
	}
	if !strings.Contains(err.Error(), "failed to parse") {
		t.Errorf("Error = %v, want JSON parse error", err)
	}
}
