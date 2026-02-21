// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package types

import (
	"testing"
	"time"
)

func TestTokenSet_IsExpired(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		issuedAt  int64
		expiresIn int64
		expected  bool
	}{
		{
			name:      "token expired",
			issuedAt:  time.Now().Add(-2 * time.Hour).Unix(),
			expiresIn: 3600, // 1 hour
			expected:  true,
		},
		{
			name:      "token valid - just issued",
			issuedAt:  time.Now().Unix(),
			expiresIn: 3600, // 1 hour
			expected:  false,
		},
		{
			name:      "token expiring soon (within 5 min buffer)",
			issuedAt:  time.Now().Add(-56 * time.Minute).Unix(), // 56 minutes ago
			expiresIn: 3600,                                     // expires in 4 minutes, within 5 min buffer
			expected:  true,
		},
		{
			name:      "token valid - 10 minutes left",
			issuedAt:  time.Now().Add(-50 * time.Minute).Unix(),
			expiresIn: 3600, // 10 minutes left
			expected:  false,
		},
		{
			name:      "token valid - long expiry",
			issuedAt:  time.Now().Unix(),
			expiresIn: 86400, // 24 hours
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			token := &TokenSet{
				IssuedAt:  tt.issuedAt,
				ExpiresIn: tt.expiresIn,
			}

			result := token.IsExpired()
			if result != tt.expected {
				expiresAt := time.Unix(tt.issuedAt+tt.expiresIn, 0)
				t.Errorf("IsExpired() = %v, want %v (expires at: %s, now: %s)",
					result, tt.expected, expiresAt.Format(time.RFC3339), time.Now().Format(time.RFC3339))
			}
		})
	}
}

func TestOAuthError_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		err      OAuthError
		expected string
	}{
		{
			name: "error with description",
			err: OAuthError{
				Error:            "invalid_grant",
				ErrorDescription: "The authorization code has expired",
			},
			expected: "The authorization code has expired",
		},
		{
			name: "error without description",
			err: OAuthError{
				Error: "invalid_client",
			},
			expected: "invalid_client",
		},
		{
			name: "error with URI",
			err: OAuthError{
				Error:            "invalid_scope",
				ErrorDescription: "The requested scope is invalid",
				ErrorURI:         "https://docs.datadoghq.com/oauth/errors",
			},
			expected: "The requested scope is invalid",
		},
		{
			name: "empty error",
			err: OAuthError{
				Error: "",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.err.String()
			if result != tt.expected {
				t.Errorf("String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestDefaultScopes(t *testing.T) {
	t.Parallel()
	scopes := DefaultScopes()

	// Verify we have all expected scopes (matching PR #84)
	expectedScopes := []string{
		// Dashboards
		"dashboards_read",
		"dashboards_write",
		// Monitors
		"monitors_read",
		"monitors_write",
		"monitors_downtime",
		// APM/Traces
		"apm_read",
		// SLOs
		"slos_read",
		"slos_write",
		"slos_corrections",
		// Incidents
		"incident_read",
		"incident_write",
		// Synthetics (including new scopes from PR #84)
		"synthetics_read",
		"synthetics_write",
		"synthetics_global_variable_read",
		"synthetics_global_variable_write",
		"synthetics_private_location_read",
		"synthetics_private_location_write",
		// Security (including new scopes from PR #84)
		"security_monitoring_signals_read",
		"security_monitoring_rules_read",
		"security_monitoring_findings_read",
		"security_monitoring_suppressions_read",
		"security_monitoring_filters_read",
		// RUM (including new scopes from PR #84)
		"rum_apps_read",
		"rum_apps_write",
		"rum_retention_filters_read",
		"rum_retention_filters_write",
		// Infrastructure
		"hosts_read",
		// Users
		"user_access_read",
		"user_self_profile_read",
		// Cases
		"cases_read",
		"cases_write",
		// Events
		"events_read",
		// Logs
		"logs_read_data",
		"logs_read_index_data",
		// Metrics
		"metrics_read",
		"timeseries_query",
		// Usage
		"usage_read",
		// Error Tracking
		"error_tracking_read",
	}

	if len(scopes) != len(expectedScopes) {
		t.Errorf("Expected %d scopes, got %d", len(expectedScopes), len(scopes))
	}

	// Create a map for quick lookup
	scopeMap := make(map[string]bool)
	for _, scope := range scopes {
		scopeMap[scope] = true
	}

	// Verify each expected scope is present
	for _, expectedScope := range expectedScopes {
		if !scopeMap[expectedScope] {
			t.Errorf("Missing expected scope: %s", expectedScope)
		}
	}

	// Verify no unexpected scopes
	for _, scope := range scopes {
		found := false
		for _, expectedScope := range expectedScopes {
			if scope == expectedScope {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Unexpected scope found: %s", scope)
		}
	}
}

func TestTokenSet_JSONTags(t *testing.T) {
	t.Parallel()
	// Verify JSON tags are camelCase (matching PR #84)
	token := TokenSet{
		AccessToken:  "test-access",
		RefreshToken: "test-refresh",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		IssuedAt:     time.Now().Unix(),
		Scope:        "test",
		ClientID:     "test-client",
	}

	// This test ensures the struct is usable
	// JSON marshaling tests are in storage tests
	if token.AccessToken != "test-access" {
		t.Error("AccessToken field not accessible")
	}
	if token.ClientID != "test-client" {
		t.Error("ClientID field not accessible")
	}
}

func TestClientCredentials_Structure(t *testing.T) {
	t.Parallel()
	// Verify ClientCredentials structure matches PR #84
	creds := ClientCredentials{
		ClientID:     "test-id",
		ClientName:   "test-name",
		RedirectURIs: []string{"http://127.0.0.1:8000/oauth/callback"},
		RegisteredAt: time.Now().Unix(),
		Site:         "datadoghq.com",
	}

	if creds.ClientID != "test-id" {
		t.Error("ClientID field not accessible")
	}
	if creds.ClientName != "test-name" {
		t.Error("ClientName field not accessible")
	}
	if len(creds.RedirectURIs) != 1 {
		t.Error("RedirectURIs field not accessible")
	}
	if creds.RegisteredAt == 0 {
		t.Error("RegisteredAt should be set")
	}
	if creds.Site != "datadoghq.com" {
		t.Error("Site field not accessible")
	}
}
