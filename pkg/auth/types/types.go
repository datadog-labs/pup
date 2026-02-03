// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package types

import "time"

// TokenSet represents OAuth2 tokens
type TokenSet struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
	Scope        string    `json:"scope"`
}

// IsExpired checks if the access token is expired
func (t *TokenSet) IsExpired() bool {
	// Consider token expired 5 minutes before actual expiration
	return time.Now().Add(5 * time.Minute).After(t.ExpiresAt)
}

// ClientCredentials represents DCR client credentials
type ClientCredentials struct {
	ClientID     string    `json:"client_id"`
	ClientSecret string    `json:"client_secret"`
	CreatedAt    time.Time `json:"created_at"`
	Site         string    `json:"site"`
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
		// Security
		"security_monitoring_signals_read",
		"security_monitoring_rules_read",
		"security_monitoring_findings_read",
		// RUM
		"rum_apps_read",
		"rum_apps_write",
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
