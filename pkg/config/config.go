// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package config

import (
	"fmt"
	"os"
	"strings"
)

// Config holds the application configuration
type Config struct {
	APIKey      string
	AppKey      string
	AccessToken string
	Site        string
	AutoApprove bool
	AgentMode   bool
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		APIKey:      os.Getenv("DD_API_KEY"),
		AppKey:      os.Getenv("DD_APP_KEY"),
		AccessToken: os.Getenv("DD_ACCESS_TOKEN"),
		Site:        getEnvWithDefault("DD_SITE", "datadoghq.com"),
		AutoApprove: os.Getenv("DD_AUTO_APPROVE") == "true" || os.Getenv("DD_CLI_AUTO_APPROVE") == "true",
	}

	return cfg, nil
}

// Validate checks if required configuration is present
// Note: This only validates the site. Authentication can be via OAuth2 or API keys,
// which is checked in the client package.
func (c *Config) Validate() error {
	if c.Site == "" {
		return fmt.Errorf("DD_SITE is required")
	}
	return nil
}

// IsOnCallSite returns true if the site is an on-call domain (contains "oncall").
// On-call domains like navy.oncall.datadoghq.com are already fully-qualified
// and should not have "api." prepended.
func IsOnCallSite(site string) bool {
	return strings.Contains(site, "oncall")
}

// GetAPIURL returns the full API URL for the configured site.
// On-call domains are used as-is; all other sites get "https://api.{site}".
func (c *Config) GetAPIURL() string {
	return "https://" + c.GetAPIHost()
}

// GetAPIHost returns the host portion of the API URL (without scheme).
// On-call domains are returned as-is; all other sites get "api.{site}".
func (c *Config) GetAPIHost() string {
	if IsOnCallSite(c.Site) {
		return c.Site
	}
	return fmt.Sprintf("api.%s", c.Site)
}

// getEnvWithDefault returns environment variable value or default if not set
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
