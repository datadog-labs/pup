// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package oauth

import (
	"strings"
	"testing"

	"github.com/DataDog/pup/pkg/auth/types"
)

func TestNewClient(t *testing.T) {
	client := NewClient("datadoghq.com")
	if client == nil {
		t.Fatal("NewClient() returned nil")
	}
}

func TestClient_BuildAuthorizationURL(t *testing.T) {
	client := NewClient("datadoghq.com")

	challenge := &PKCEChallenge{
		Verifier:  "test-verifier",
		Challenge: "test-challenge",
		Method:    "S256",
	}

	scopes := []string{"dashboards_read", "monitors_read"}
	url := client.BuildAuthorizationURL(
		"test-client-id",
		"http://127.0.0.1:8000/oauth/callback",
		"test-state",
		challenge,
		scopes,
	)

	// Verify URL contains expected components
	expectedComponents := []string{
		"https://app.datadoghq.com/oauth2/v1/authorize",
		"response_type=code",
		"client_id=test-client-id",
		"redirect_uri=http",
		"state=test-state",
		"code_challenge=test-challenge",
		"code_challenge_method=S256",
	}

	for _, component := range expectedComponents {
		if !strings.Contains(url, component) {
			t.Errorf("URL missing component %q. Got: %s", component, url)
		}
	}

	// Verify scopes are space-separated in single parameter (OAuth2 spec)
	if !strings.Contains(url, "scope=dashboards_read") {
		t.Errorf("URL missing scope parameter. Got: %s", url)
	}
	if !strings.Contains(url, "monitors_read") {
		t.Errorf("URL missing monitors_read scope. Got: %s", url)
	}

	// Verify URL structure
	if !strings.HasPrefix(url, "https://app.datadoghq.com/oauth2/v1/authorize?") {
		t.Errorf("URL has incorrect base. Got: %s", url)
	}
}

func TestClient_BuildAuthorizationURL_DifferentSites(t *testing.T) {
	tests := []struct {
		name         string
		site         string
		expectedBase string
	}{
		{
			name:         "US1 site",
			site:         "datadoghq.com",
			expectedBase: "https://app.datadoghq.com/oauth2/v1/authorize",
		},
		{
			name:         "EU site",
			site:         "datadoghq.eu",
			expectedBase: "https://app.datadoghq.eu/oauth2/v1/authorize",
		},
		{
			name:         "US3 site",
			site:         "us3.datadoghq.com",
			expectedBase: "https://app.us3.datadoghq.com/oauth2/v1/authorize",
		},
		{
			name:         "Gov site",
			site:         "ddog-gov.com",
			expectedBase: "https://app.ddog-gov.com/oauth2/v1/authorize",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.site)
			challenge := &PKCEChallenge{
				Challenge: "test",
				Method:    "S256",
			}

			url := client.BuildAuthorizationURL(
				"client-id",
				"http://127.0.0.1:8000/oauth/callback",
				"state",
				challenge,
				[]string{"test"},
			)

			if !strings.HasPrefix(url, tt.expectedBase) {
				t.Errorf("URL = %s, want prefix %s", url, tt.expectedBase)
			}
		})
	}
}

func TestClient_BuildAuthorizationURL_WithDefaultScopes(t *testing.T) {
	client := NewClient("datadoghq.com")
	challenge := &PKCEChallenge{
		Challenge: "test-challenge",
		Method:    "S256",
	}

	url := client.BuildAuthorizationURL(
		"test-client-id",
		"http://127.0.0.1:8000/oauth/callback",
		"test-state",
		challenge,
		nil, // Empty scopes should use defaults
	)

	// Should contain some default scopes
	if !strings.Contains(url, "scope=") {
		t.Error("URL missing scope parameter")
	}
}

func TestClient_ValidateCallback(t *testing.T) {
	client := NewClient("datadoghq.com")

	tests := []struct {
		name          string
		code          string
		state         string
		expectedState string
		wantError     bool
		errorContains string
	}{
		{
			name:          "valid callback",
			code:          "auth-code-123",
			state:         "state-abc",
			expectedState: "state-abc",
			wantError:     false,
		},
		{
			name:          "missing code",
			code:          "",
			state:         "state-abc",
			expectedState: "state-abc",
			wantError:     true,
			errorContains: "authorization code",
		},
		{
			name:          "missing state",
			code:          "auth-code-123",
			state:         "",
			expectedState: "state-abc",
			wantError:     true,
			errorContains: "state parameter",
		},
		{
			name:          "state mismatch",
			code:          "auth-code-123",
			state:         "wrong-state",
			expectedState: "correct-state",
			wantError:     true,
			errorContains: "mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.ValidateCallback(tt.code, tt.state, tt.expectedState)

			if tt.wantError {
				if err == nil {
					t.Error("ValidateCallback() expected error but got none")
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Error = %v, want to contain %q", err, tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateCallback() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestClient_ParseCallbackError(t *testing.T) {
	client := NewClient("datadoghq.com")

	tests := []struct {
		name             string
		errorCode        string
		errorDescription string
		wantError        bool
		wantContains     string
	}{
		{
			name:             "no error",
			errorCode:        "",
			errorDescription: "",
			wantError:        false,
		},
		{
			name:             "error with description",
			errorCode:        "invalid_grant",
			errorDescription: "The authorization code has expired",
			wantError:        true,
			wantContains:     "The authorization code has expired",
		},
		{
			name:             "error without description",
			errorCode:        "access_denied",
			errorDescription: "",
			wantError:        true,
			wantContains:     "access_denied",
		},
		{
			name:             "invalid_client error",
			errorCode:        "invalid_client",
			errorDescription: "Client authentication failed",
			wantError:        true,
			wantContains:     "Client authentication failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.ParseCallbackError(tt.errorCode, tt.errorDescription)

			if tt.wantError {
				if err == nil {
					t.Error("ParseCallbackError() expected error but got none")
					return
				}
				if tt.wantContains != "" && !strings.Contains(err.Error(), tt.wantContains) {
					t.Errorf("Error = %v, want to contain %q", err, tt.wantContains)
				}
			} else {
				if err != nil {
					t.Errorf("ParseCallbackError() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestClient_GetAuthConfig(t *testing.T) {
	client := NewClient("datadoghq.com")

	tests := []struct {
		name        string
		scopes      []string
		wantDefault bool
	}{
		{
			name:        "custom scopes",
			scopes:      []string{"dashboards_read", "monitors_write"},
			wantDefault: false,
		},
		{
			name:        "nil scopes uses defaults",
			scopes:      nil,
			wantDefault: true,
		},
		{
			name:        "empty scopes uses defaults",
			scopes:      []string{},
			wantDefault: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := client.GetAuthConfig(tt.scopes)

			if config == nil {
				t.Fatal("GetAuthConfig() returned nil")
			}

			if config.Site != "datadoghq.com" {
				t.Errorf("Site = %s, want datadoghq.com", config.Site)
			}

			if tt.wantDefault {
				defaultScopes := types.DefaultScopes()
				if len(config.Scopes) != len(defaultScopes) {
					t.Errorf("Scopes length = %d, want %d (default)", len(config.Scopes), len(defaultScopes))
				}
			} else {
				if len(config.Scopes) != len(tt.scopes) {
					t.Errorf("Scopes length = %d, want %d", len(config.Scopes), len(tt.scopes))
				}
			}
		})
	}
}
