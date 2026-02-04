// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package client

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/pup/pkg/auth/storage"
	"github.com/DataDog/pup/pkg/config"
)

// Client wraps the Datadog API client
type Client struct {
	config *config.Config
	ctx    context.Context
	api    *datadog.APIClient
}

// New creates a new Datadog API client
// Authentication priority:
//  1. OAuth2 tokens (if available and valid)
//  2. API keys (DD_API_KEY and DD_APP_KEY)
func New(cfg *config.Config) (*Client, error) {
	var ctx context.Context

	// Try OAuth2 tokens first (preferred method)
	store, err := storage.GetStorage(nil)
	if err == nil {
		tokens, err := store.LoadTokens(cfg.Site)
		if err == nil && tokens != nil && !tokens.IsExpired() {
			// Use OAuth2 Bearer token authentication
			ctx = context.WithValue(
				context.Background(),
				datadog.ContextAccessToken,
				tokens.AccessToken,
			)
		}
	}

	// Fall back to API keys if OAuth not available
	if ctx == nil {
		if cfg.APIKey == "" || cfg.AppKey == "" {
			return nil, fmt.Errorf(
				"authentication required: either run 'fetch auth login' for OAuth2 or set DD_API_KEY and DD_APP_KEY environment variables",
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
	configuration.Host = fmt.Sprintf("api.%s", cfg.Site)
	configuration.SetUnstableOperationEnabled("v2.QueryTimeseriesData", true)
	configuration.SetUnstableOperationEnabled("v2.ListIncidents", true)
	configuration.SetUnstableOperationEnabled("v2.GetIncident", true)

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
