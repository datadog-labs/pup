// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/datadog-labs/pup/pkg/config"
)

// EndpointAuthRequirement defines authentication requirements for API endpoints
type EndpointAuthRequirement struct {
	Path            string
	Method          string
	SupportsOAuth   bool
	RequiresAPIKeys bool
	Reason          string
}

// endpointsWithoutOAuth defines endpoints that do NOT support OAuth and require API/App keys
// Based on analysis of datadog-api-spec repository
var endpointsWithoutOAuth = []EndpointAuthRequirement{
	// Logs API - missing logs_read_data/logs_write_data OAuth scopes
	{Path: "/api/v2/logs/events", Method: "POST", SupportsOAuth: false, RequiresAPIKeys: true, Reason: "Logs API missing OAuth implementation in spec"},
	{Path: "/api/v2/logs/events/search", Method: "POST", SupportsOAuth: false, RequiresAPIKeys: true, Reason: "Logs API missing OAuth implementation in spec"},
	{Path: "/api/v2/logs/analytics/aggregate", Method: "POST", SupportsOAuth: false, RequiresAPIKeys: true, Reason: "Logs API missing OAuth implementation in spec"},
	{Path: "/api/v2/logs/config/archives", Method: "GET", SupportsOAuth: false, RequiresAPIKeys: true, Reason: "Logs config API missing OAuth implementation in spec"},
	{Path: "/api/v2/logs/config/archives/", Method: "GET", SupportsOAuth: false, RequiresAPIKeys: true, Reason: "Logs config API missing OAuth implementation in spec"},
	{Path: "/api/v2/logs/config/archives/", Method: "DELETE", SupportsOAuth: false, RequiresAPIKeys: true, Reason: "Logs config API missing OAuth implementation in spec"},
	{Path: "/api/v2/logs/config/custom_destinations", Method: "GET", SupportsOAuth: false, RequiresAPIKeys: true, Reason: "Logs config API missing OAuth implementation in spec"},
	{Path: "/api/v2/logs/config/custom_destinations/", Method: "GET", SupportsOAuth: false, RequiresAPIKeys: true, Reason: "Logs config API missing OAuth implementation in spec"},
	{Path: "/api/v2/logs/config/metrics", Method: "GET", SupportsOAuth: false, RequiresAPIKeys: true, Reason: "Logs config API missing OAuth implementation in spec"},
	{Path: "/api/v2/logs/config/metrics/", Method: "GET", SupportsOAuth: false, RequiresAPIKeys: true, Reason: "Logs config API missing OAuth implementation in spec"},
	{Path: "/api/v2/logs/config/metrics/", Method: "DELETE", SupportsOAuth: false, RequiresAPIKeys: true, Reason: "Logs config API missing OAuth implementation in spec"},

	// RUM API - missing rum_apps_read/rum_apps_write OAuth scopes
	{Path: "/api/v2/rum/applications", Method: "GET", SupportsOAuth: false, RequiresAPIKeys: true, Reason: "RUM API missing OAuth implementation in spec"},
	{Path: "/api/v2/rum/applications/", Method: "GET", SupportsOAuth: false, RequiresAPIKeys: true, Reason: "RUM API missing OAuth implementation in spec"},
	{Path: "/api/v2/rum/applications", Method: "POST", SupportsOAuth: false, RequiresAPIKeys: true, Reason: "RUM API missing OAuth implementation in spec"},
	{Path: "/api/v2/rum/applications/", Method: "PATCH", SupportsOAuth: false, RequiresAPIKeys: true, Reason: "RUM API missing OAuth implementation in spec"},
	{Path: "/api/v2/rum/applications/", Method: "DELETE", SupportsOAuth: false, RequiresAPIKeys: true, Reason: "RUM API missing OAuth implementation in spec"},
	{Path: "/api/v2/rum/metrics", Method: "GET", SupportsOAuth: false, RequiresAPIKeys: true, Reason: "RUM metrics API missing OAuth implementation in spec"},
	{Path: "/api/v2/rum/metrics/", Method: "GET", SupportsOAuth: false, RequiresAPIKeys: true, Reason: "RUM metrics API missing OAuth implementation in spec"},
	{Path: "/api/v2/rum/retention_filters", Method: "GET", SupportsOAuth: false, RequiresAPIKeys: true, Reason: "RUM retention filters API missing OAuth implementation in spec"},
	{Path: "/api/v2/rum/retention_filters/", Method: "GET", SupportsOAuth: false, RequiresAPIKeys: true, Reason: "RUM retention filters API missing OAuth implementation in spec"},
	{Path: "/api/v2/rum/events/search", Method: "POST", SupportsOAuth: false, RequiresAPIKeys: true, Reason: "RUM events API missing OAuth implementation in spec"},

	// API/App Keys Management - missing api_keys_read/write OAuth scopes
	{Path: "/api/v2/api_keys", Method: "GET", SupportsOAuth: false, RequiresAPIKeys: true, Reason: "API Keys management missing OAuth implementation in spec"},
	{Path: "/api/v2/api_keys/", Method: "GET", SupportsOAuth: false, RequiresAPIKeys: true, Reason: "API Keys management missing OAuth implementation in spec"},
	{Path: "/api/v2/api_keys", Method: "POST", SupportsOAuth: false, RequiresAPIKeys: true, Reason: "API Keys management missing OAuth implementation in spec"},
	{Path: "/api/v2/api_keys/", Method: "DELETE", SupportsOAuth: false, RequiresAPIKeys: true, Reason: "API Keys management missing OAuth implementation in spec"},
	{Path: "/api/v2/application_keys", Method: "GET", SupportsOAuth: false, RequiresAPIKeys: true, Reason: "Application Keys management missing OAuth implementation in spec"},
	{Path: "/api/v2/application_keys/", Method: "GET", SupportsOAuth: false, RequiresAPIKeys: true, Reason: "Application Keys management missing OAuth implementation in spec"},
	{Path: "/api/v2/application_keys/", Method: "PATCH", SupportsOAuth: false, RequiresAPIKeys: true, Reason: "Application Keys management missing OAuth implementation in spec"},
	{Path: "/api/v2/application_keys/", Method: "DELETE", SupportsOAuth: false, RequiresAPIKeys: true, Reason: "Application Keys management missing OAuth implementation in spec"},

	// Error Tracking API - OAuth not working in practice
	{Path: "/api/v2/error_tracking/issues/search", Method: "POST", SupportsOAuth: false, RequiresAPIKeys: true, Reason: "Error Tracking API requires API keys"},
	{Path: "/api/v2/error_tracking/issues/", Method: "GET", SupportsOAuth: false, RequiresAPIKeys: true, Reason: "Error Tracking API requires API keys"},

	// Notebooks API (V1) - missing OAuth scopes in spec
	{Path: "/api/v1/notebooks", Method: "GET", SupportsOAuth: false, RequiresAPIKeys: true, Reason: "Notebooks API missing OAuth implementation in spec"},
	{Path: "/api/v1/notebooks", Method: "POST", SupportsOAuth: false, RequiresAPIKeys: true, Reason: "Notebooks API missing OAuth implementation in spec"},
	{Path: "/api/v1/notebooks/", Method: "GET", SupportsOAuth: false, RequiresAPIKeys: true, Reason: "Notebooks API missing OAuth implementation in spec"},
	{Path: "/api/v1/notebooks/", Method: "PUT", SupportsOAuth: false, RequiresAPIKeys: true, Reason: "Notebooks API missing OAuth implementation in spec"},
	{Path: "/api/v1/notebooks/", Method: "DELETE", SupportsOAuth: false, RequiresAPIKeys: true, Reason: "Notebooks API missing OAuth implementation in spec"},
}

// AuthType represents the type of authentication being used
type AuthType int

const (
	AuthTypeNone AuthType = iota
	AuthTypeOAuth
	AuthTypeAPIKeys
)

// GetAuthType returns the authentication type from the context
func GetAuthType(ctx context.Context) AuthType {
	if token, ok := ctx.Value(datadog.ContextAccessToken).(string); ok && token != "" {
		return AuthTypeOAuth
	}
	if apiKeys, ok := ctx.Value(datadog.ContextAPIKeys).(map[string]datadog.APIKey); ok && len(apiKeys) > 0 {
		return AuthTypeAPIKeys
	}
	return AuthTypeNone
}

// ValidateEndpointAuth checks if the endpoint can be accessed with the current authentication
// Returns an error if:
// 1. The endpoint doesn't support OAuth but only OAuth is available
// 2. The endpoint requires API keys but they're not configured
func ValidateEndpointAuth(ctx context.Context, cfg *config.Config, method, path string) error {
	authType := GetAuthType(ctx)

	// Check if this endpoint requires special handling
	requirement := getEndpointRequirement(method, path)
	if requirement == nil {
		// Endpoint supports OAuth, no special validation needed
		return nil
	}

	// Endpoint doesn't support OAuth
	if !requirement.SupportsOAuth {
		if authType == AuthTypeOAuth {
			// User authenticated with OAuth but endpoint doesn't support it
			// Check if API keys are available as fallback
			if cfg.APIKey == "" || cfg.AppKey == "" {
				return fmt.Errorf(
					"endpoint %s %s does not support OAuth authentication. "+
						"Please set DD_API_KEY and DD_APP_KEY environment variables. "+
						"Reason: %s",
					method, path, requirement.Reason,
				)
			}
			// API keys are available, the client will need to be recreated with API keys
			// This is handled at the command level
		} else if authType == AuthTypeAPIKeys {
			// Already using API keys, all good
			return nil
		} else {
			// No authentication available
			return fmt.Errorf(
				"endpoint %s %s requires API key authentication. "+
					"Please set DD_API_KEY and DD_APP_KEY environment variables. "+
					"Reason: %s",
				method, path, requirement.Reason,
			)
		}
	}

	return nil
}

// RequiresAPIKeyFallback returns true if the endpoint doesn't support OAuth
// and we need to fallback to API keys even if OAuth is available
func RequiresAPIKeyFallback(method, path string) bool {
	requirement := getEndpointRequirement(method, path)
	return requirement != nil && !requirement.SupportsOAuth
}

// getEndpointRequirement finds the auth requirement for an endpoint
func getEndpointRequirement(method, path string) *EndpointAuthRequirement {
	for _, req := range endpointsWithoutOAuth {
		// Handle paths with IDs (e.g., /api/v2/rum/applications/{id})
		if strings.HasSuffix(req.Path, "/") {
			if strings.HasPrefix(path, req.Path[:len(req.Path)-1]) && req.Method == method {
				return &req
			}
		} else if req.Path == path && req.Method == method {
			return &req
		}
	}
	return nil
}

// GetAuthTypeDescription returns a human-readable description of the auth type
func GetAuthTypeDescription(authType AuthType) string {
	switch authType {
	case AuthTypeOAuth:
		return "OAuth2 Bearer Token"
	case AuthTypeAPIKeys:
		return "API Keys (DD_API_KEY + DD_APP_KEY)"
	default:
		return "None"
	}
}
