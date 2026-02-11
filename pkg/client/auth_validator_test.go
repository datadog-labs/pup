// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package client

import (
	"context"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/pup/pkg/config"
)

func TestGetAuthType(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected AuthType
	}{
		{
			name:     "no auth",
			ctx:      context.Background(),
			expected: AuthTypeNone,
		},
		{
			name:     "oauth token",
			ctx:      context.WithValue(context.Background(), datadog.ContextAccessToken, "test-token"),
			expected: AuthTypeOAuth,
		},
		{
			name: "api keys",
			ctx: context.WithValue(context.Background(), datadog.ContextAPIKeys, map[string]datadog.APIKey{
				"apiKeyAuth": {Key: "test-api-key"},
				"appKeyAuth": {Key: "test-app-key"},
			}),
			expected: AuthTypeAPIKeys,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetAuthType(tt.ctx)
			if got != tt.expected {
				t.Errorf("GetAuthType() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRequiresAPIKeyFallback(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		path     string
		expected bool
	}{
		{
			name:     "logs search requires API keys",
			method:   "POST",
			path:     "/api/v2/logs/events/search",
			expected: true,
		},
		{
			name:     "rum apps list requires API keys",
			method:   "GET",
			path:     "/api/v2/rum/applications",
			expected: true,
		},
		{
			name:     "api keys list requires API keys",
			method:   "GET",
			path:     "/api/v2/api_keys",
			expected: true,
		},
		{
			name:     "error tracking search requires API keys",
			method:   "POST",
			path:     "/api/v2/error_tracking/issues/search",
			expected: true,
		},
		{
			name:     "error tracking get requires API keys",
			method:   "GET",
			path:     "/api/v2/error_tracking/issues/abc123",
			expected: true,
		},
		{
			name:     "monitors list supports OAuth",
			method:   "GET",
			path:     "/api/v1/monitor",
			expected: false,
		},
		{
			name:     "dashboards list supports OAuth",
			method:   "GET",
			path:     "/api/v1/dashboard",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RequiresAPIKeyFallback(tt.method, tt.path)
			if got != tt.expected {
				t.Errorf("RequiresAPIKeyFallback(%s, %s) = %v, want %v", tt.method, tt.path, got, tt.expected)
			}
		})
	}
}

func TestValidateEndpointAuth(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		cfg       *config.Config
		method    string
		path      string
		wantError bool
	}{
		{
			name:      "OAuth with OAuth-supported endpoint - success",
			ctx:       context.WithValue(context.Background(), datadog.ContextAccessToken, "test-token"),
			cfg:       &config.Config{APIKey: "", AppKey: ""},
			method:    "GET",
			path:      "/api/v1/monitor",
			wantError: false,
		},
		{
			name:      "OAuth with non-OAuth endpoint but API keys available - success",
			ctx:       context.WithValue(context.Background(), datadog.ContextAccessToken, "test-token"),
			cfg:       &config.Config{APIKey: "key", AppKey: "app"},
			method:    "POST",
			path:      "/api/v2/logs/events/search",
			wantError: false,
		},
		{
			name:      "OAuth with non-OAuth endpoint and no API keys - error",
			ctx:       context.WithValue(context.Background(), datadog.ContextAccessToken, "test-token"),
			cfg:       &config.Config{APIKey: "", AppKey: ""},
			method:    "POST",
			path:      "/api/v2/logs/events/search",
			wantError: true,
		},
		{
			name: "API keys with non-OAuth endpoint - success",
			ctx: context.WithValue(context.Background(), datadog.ContextAPIKeys, map[string]datadog.APIKey{
				"apiKeyAuth": {Key: "test-api-key"},
			}),
			cfg:       &config.Config{APIKey: "key", AppKey: "app"},
			method:    "POST",
			path:      "/api/v2/logs/events/search",
			wantError: false,
		},
		{
			name:      "No auth with non-OAuth endpoint - error",
			ctx:       context.Background(),
			cfg:       &config.Config{APIKey: "", AppKey: ""},
			method:    "GET",
			path:      "/api/v2/rum/applications",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEndpointAuth(tt.ctx, tt.cfg, tt.method, tt.path)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateEndpointAuth() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestGetEndpointRequirement(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		path         string
		wantNil      bool
		wantOAuth    bool
		wantAPIKeys  bool
	}{
		{
			name:        "logs endpoint",
			method:      "POST",
			path:        "/api/v2/logs/events/search",
			wantNil:     false,
			wantOAuth:   false,
			wantAPIKeys: true,
		},
		{
			name:        "rum endpoint with ID",
			method:      "GET",
			path:        "/api/v2/rum/applications/abc123",
			wantNil:     false,
			wantOAuth:   false,
			wantAPIKeys: true,
		},
		{
			name:    "monitors endpoint (OAuth supported)",
			method:  "GET",
			path:    "/api/v1/monitor",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := getEndpointRequirement(tt.method, tt.path)
			if tt.wantNil {
				if req != nil {
					t.Errorf("getEndpointRequirement() = %v, want nil", req)
				}
			} else {
				if req == nil {
					t.Errorf("getEndpointRequirement() = nil, want non-nil")
					return
				}
				if req.SupportsOAuth != tt.wantOAuth {
					t.Errorf("SupportsOAuth = %v, want %v", req.SupportsOAuth, tt.wantOAuth)
				}
				if req.RequiresAPIKeys != tt.wantAPIKeys {
					t.Errorf("RequiresAPIKeys = %v, want %v", req.RequiresAPIKeys, tt.wantAPIKeys)
				}
			}
		})
	}
}

func TestGetAuthTypeDescription(t *testing.T) {
	tests := []struct {
		authType AuthType
		expected string
	}{
		{AuthTypeNone, "None"},
		{AuthTypeOAuth, "OAuth2 Bearer Token"},
		{AuthTypeAPIKeys, "API Keys (DD_API_KEY + DD_APP_KEY)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := GetAuthTypeDescription(tt.authType)
			if got != tt.expected {
				t.Errorf("GetAuthTypeDescription(%v) = %v, want %v", tt.authType, got, tt.expected)
			}
		})
	}
}
