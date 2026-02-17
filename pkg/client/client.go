// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/pup/pkg/auth/dcr"
	"github.com/DataDog/pup/pkg/auth/storage"
	"github.com/DataDog/pup/pkg/config"
	"github.com/DataDog/pup/pkg/useragent"
)

// Client wraps the Datadog API client
type Client struct {
	config *config.Config
	ctx    context.Context
	api    *datadog.APIClient
}

// Test hooks — overridden in tests to inject fakes
var (
	getStorageFunc   = func() (storage.Storage, error) { return storage.GetStorage(nil) }
	newDCRClientFunc = func(site string) *dcr.Client { return dcr.NewClient(site) }
)

// New creates a new Datadog API client
// Authentication priority:
//  1. OAuth2 tokens (if available and valid)
//  2. API keys (DD_API_KEY and DD_APP_KEY)
func New(cfg *config.Config) (*Client, error) {
	return NewWithOptions(cfg, false)
}

// NewWithAPIKeys creates a new Datadog API client forcing API key authentication
// This is used for endpoints that don't support OAuth2
func NewWithAPIKeys(cfg *config.Config) (*Client, error) {
	return NewWithOptions(cfg, true)
}

// NewWithOptions creates a new Datadog API client with authentication options
func NewWithOptions(cfg *config.Config, forceAPIKeys bool) (*Client, error) {
	var ctx context.Context

	if !forceAPIKeys {
		// Try OAuth2 tokens first (preferred method)
		store, err := getStorageFunc()
		if err == nil {
			tokens, err := store.LoadTokens(cfg.Site)
			if err == nil && tokens != nil {
				// Auto-refresh: if token is expired but refresh token is available, refresh it
				if tokens.IsExpired() && tokens.RefreshToken != "" {
					creds, credsErr := store.LoadClientCredentials(cfg.Site)
					if credsErr == nil && creds != nil {
						dcrClient := newDCRClientFunc(cfg.Site)
						newTokens, refreshErr := dcrClient.RefreshToken(tokens.RefreshToken, creds)
						if refreshErr == nil {
							_ = store.SaveTokens(cfg.Site, newTokens)
							tokens = newTokens
						}
					}
				}

				if !tokens.IsExpired() {
					// Use OAuth2 Bearer token authentication
					ctx = context.WithValue(
						context.Background(),
						datadog.ContextAccessToken,
						tokens.AccessToken,
					)
				}
			}
		}
	}

	// Fall back to API keys if OAuth not available or forced
	if ctx == nil {
		if cfg.APIKey == "" || cfg.AppKey == "" {
			return nil, fmt.Errorf(
				"authentication required: either run 'pup auth login' for OAuth2 or set DD_API_KEY and DD_APP_KEY environment variables",
			)
		}

		ctx = context.WithValue(
			context.Background(),
			datadog.ContextAPIKeys,
			map[string]datadog.APIKey{
				"apiKeyAuth": {
					Key: cfg.APIKey,
				},
				"appKeyAuth": {
					Key: cfg.AppKey,
				},
			},
		)
	}

	// Configure the API client
	configuration := datadog.NewConfiguration()
	configuration.Host = cfg.GetAPIHost()

	// Set custom user agent to identify requests as coming from pup CLI
	configuration.UserAgent = useragent.Get()

	// Enable all unstable operations to suppress warnings
	// These are beta/preview features that we want to use
	unstableOps := []string{
		"v2.ListIncidents",
		"v2.GetIncident",
		"v2.CreateIncident",
		"v2.UpdateIncident",
		"v2.DeleteIncident",
	}
	for _, op := range unstableOps {
		configuration.SetUnstableOperationEnabled(op, true)
	}

	api := datadog.NewAPIClient(configuration)

	return &Client{
		config: cfg,
		ctx:    ctx,
		api:    api,
	}, nil
}

// Context returns the client context
func (c *Client) Context() context.Context {
	return c.ctx
}

// V1 returns the v1 API client
func (c *Client) V1() *datadog.APIClient {
	return c.api
}

// V2 returns the v2 API client
func (c *Client) V2() *datadog.APIClient {
	return c.api
}

// API returns the API client
func (c *Client) API() *datadog.APIClient {
	return c.api
}

// Config returns the client configuration
func (c *Client) Config() *config.Config {
	return c.config
}

// ValidateEndpointAuth checks if the current authentication is compatible with the endpoint
func (c *Client) ValidateEndpointAuth(method, path string) error {
	return ValidateEndpointAuth(c.ctx, c.config, method, path)
}

// GetAuthType returns the type of authentication being used by this client
func (c *Client) GetAuthType() AuthType {
	return GetAuthType(c.ctx)
}

// RawRequest makes an HTTP request with proper authentication headers.
// This is used for APIs not covered by the typed datadog-api-client-go library.
func (c *Client) RawRequest(method, path string, body io.Reader) (*http.Response, error) {
	// Validate endpoint auth before making the request
	if err := c.ValidateEndpointAuth(method, path); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://%s%s", c.config.GetAPIHost(), path)

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", useragent.Get())

	// Set auth headers from context
	if token, ok := c.ctx.Value(datadog.ContextAccessToken).(string); ok && token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	} else if apiKeys, ok := c.ctx.Value(datadog.ContextAPIKeys).(map[string]datadog.APIKey); ok {
		if key, exists := apiKeys["apiKeyAuth"]; exists {
			req.Header.Set("DD-API-KEY", key.Key)
		}
		if key, exists := apiKeys["appKeyAuth"]; exists {
			req.Header.Set("DD-APPLICATION-KEY", key.Key)
		}
	}

	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}

	return resp, nil
}
