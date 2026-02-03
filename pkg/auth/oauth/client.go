// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package oauth

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/DataDog/fetch/pkg/auth/types"
)

// Client handles OAuth2 authorization flow
type Client struct {
	site string
}

// NewClient creates a new OAuth client
func NewClient(site string) *Client {
	return &Client{
		site: site,
	}
}

// BuildAuthorizationURL builds the OAuth2 authorization URL with PKCE
func (c *Client) BuildAuthorizationURL(
	clientID string,
	redirectURI string,
	state string,
	challenge *PKCEChallenge,
	scopes []string,
) string {
	params := url.Values{
		"response_type":         {"code"},
		"client_id":             {clientID},
		"redirect_uri":          {redirectURI},
		"state":                 {state},
		"scope":                 {strings.Join(scopes, " ")},
		"code_challenge":        {challenge.Challenge},
		"code_challenge_method": {challenge.Method},
	}

	baseURL := fmt.Sprintf("https://%s/oauth2/v1/authorize", c.site)
	return fmt.Sprintf("%s?%s", baseURL, params.Encode())
}

// ValidateCallback validates the OAuth callback parameters
func (c *Client) ValidateCallback(code, state, expectedState string) error {
	if code == "" {
		return fmt.Errorf("missing authorization code")
	}

	if state == "" {
		return fmt.Errorf("missing state parameter")
	}

	if state != expectedState {
		return fmt.Errorf("state parameter mismatch (CSRF protection)")
	}

	return nil
}

// ParseCallbackError parses OAuth error from callback parameters
func (c *Client) ParseCallbackError(errorCode, errorDescription string) error {
	if errorCode == "" {
		return nil
	}

	if errorDescription != "" {
		return fmt.Errorf("OAuth error: %s - %s", errorCode, errorDescription)
	}

	return fmt.Errorf("OAuth error: %s", errorCode)
}

// GetAuthConfig returns the authentication configuration
func (c *Client) GetAuthConfig(scopes []string) *types.AuthConfig {
	if scopes == nil || len(scopes) == 0 {
		scopes = types.DefaultScopes()
	}

	return &types.AuthConfig{
		Site:         c.site,
		RedirectPort: 0, // Will be assigned dynamically
		Scopes:       scopes,
	}
}
