// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package types

import "time"

// TokenSet represents OAuth2 tokens
// JSON format matches TypeScript PR #84 for cross-compatibility
type TokenSet struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	TokenType    string `json:"tokenType"`
	ExpiresIn    int64  `json:"expiresIn"`
	IssuedAt     int64  `json:"issuedAt"` // Unix timestamp in seconds
	Scope        string `json:"scope,omitempty"`
	ClientID     string `json:"clientId,omitempty"` // Client ID used for this token
}

// IsExpired checks if the access token is expired
func (t *TokenSet) IsExpired() bool {
	// Consider token expired 5 minutes before actual expiration (matches PR #84)
	expiresAt := time.Unix(t.IssuedAt+t.ExpiresIn, 0)
	return time.Now().Add(5 * time.Minute).After(expiresAt)
}

// ClientCredentials represents DCR client credentials
// JSON format matches TypeScript PR #84 for cross-compatibility
// Note: Public clients don't receive a client_secret
type ClientCredentials struct {
	ClientID     string   `json:"clientId"`
	ClientName   string   `json:"clientName"`
	RedirectURIs []string `json:"redirectUris"`
	RegisteredAt int64    `json:"registeredAt"` // Unix timestamp in seconds
	Site         string   `json:"site"`
}

// AuthConfig represents authentication configuration
type AuthConfig struct {
	Site         string
	RedirectPort int
	Scopes       []string
}

// DefaultScopes returns the default OAuth2 scopes based on PR #84
func DefaultScopes() []string {
	return []string{
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
		// Synthetics
		"synthetics_read",
		"synthetics_write",
		"synthetics_global_variable_read",
		"synthetics_global_variable_write",
		"synthetics_private_location_read",
		"synthetics_private_location_write",
		// Security
		"security_monitoring_signals_read",
		"security_monitoring_rules_read",
		"security_monitoring_findings_read",
		"security_monitoring_suppressions_read",
		"security_monitoring_filters_read",
		// RUM
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
		// CI Visibility / Test Optimization
		"ci_visibility_read",
		"test_optimization_read",
		// Usage
		"usage_read",
	}
}

// OAuthError represents an OAuth error response
type OAuthError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	ErrorURI         string `json:"error_uri,omitempty"`
}

func (e *OAuthError) String() string {
	if e.ErrorDescription != "" {
		return e.ErrorDescription
	}
	return e.Error
}
