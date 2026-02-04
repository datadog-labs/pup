// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package dcr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/DataDog/pup/pkg/auth/types"
)

// Client handles Dynamic Client Registration with Datadog
type Client struct {
	site       string
	httpClient *http.Client
}

// NewClient creates a new DCR client
func NewClient(site string) *Client {
	return &Client{
		site: site,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Register performs Dynamic Client Registration
// Registers all standard redirect URIs (ports 8000, 8080, 8888, 9000)
// Matches TypeScript PR #84 behavior for compatibility
func (c *Client) Register(redirectURI string, scopes []string) (*types.ClientCredentials, error) {
	// Build registration request - matches TypeScript PR #84
	// Only sends client_name, redirect_uris, and grant_types (no scope, auth method, or response_types)
	req := RegistrationRequest{
		ClientName:   DCRClientName,
		RedirectURIs: GetRedirectURIs(), // Register all standard ports
		GrantTypes:   []string{"authorization_code", "refresh_token"},
	}

	// Marshal request
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal registration request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("https://api.%s/api/v2/oauth2/register", c.site)
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send registration request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusCreated {
		var oauthErr types.OAuthError
		if err := json.Unmarshal(respBody, &oauthErr); err == nil {
			return nil, fmt.Errorf("DCR failed: %s", oauthErr.String())
		}
		return nil, fmt.Errorf("DCR failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var regResp RegistrationResponse
	if err := json.Unmarshal(respBody, &regResp); err != nil {
		return nil, fmt.Errorf("failed to parse registration response: %w", err)
	}

	// Return client credentials - matches TypeScript PR #84 format
	return &types.ClientCredentials{
		ClientID:     regResp.ClientID,
		ClientName:   regResp.ClientName,
		RedirectURIs: regResp.RedirectURIs,
		RegisteredAt: time.Now().Unix(),
		Site:         c.site,
	}, nil
}

// ExchangeCode exchanges an authorization code for tokens
// Matches TypeScript PR #84 - uses form-encoded data, no client_secret (public client)
func (c *Client) ExchangeCode(code, redirectURI, codeVerifier string, creds *types.ClientCredentials) (*types.TokenSet, error) {
	// Build form data
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", creds.ClientID)
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("code_verifier", codeVerifier)

	return c.requestTokens(data)
}

// RefreshToken refreshes an access token using a refresh token
// Matches TypeScript PR #84 - uses form-encoded data, no client_secret (public client)
func (c *Client) RefreshToken(refreshToken string, creds *types.ClientCredentials) (*types.TokenSet, error) {
	// Build form data
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", creds.ClientID)
	data.Set("refresh_token", refreshToken)

	return c.requestTokens(data)
}

// requestTokens makes a token request to the OAuth2 token endpoint
// Matches TypeScript PR #84 - uses application/x-www-form-urlencoded (not JSON)
func (c *Client) requestTokens(data url.Values) (*types.TokenSet, error) {
	// Create HTTP request with form-encoded body
	tokenURL := fmt.Sprintf("https://api.%s/oauth2/v1/token", c.site)
	httpReq, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept", "application/json")

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send token request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		var oauthErr types.OAuthError
		if err := json.Unmarshal(respBody, &oauthErr); err == nil {
			return nil, fmt.Errorf("token request failed: %s", oauthErr.String())
		}
		return nil, fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var tokenResp TokenResponse
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	// Get issued at timestamp (matches PR #84)
	issuedAt := time.Now().Unix()

	// Return token set with client ID
	return &types.TokenSet{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    tokenResp.TokenType,
		ExpiresIn:    tokenResp.ExpiresIn,
		IssuedAt:     issuedAt,
		Scope:        tokenResp.Scope,
		ClientID:     "", // Will be set by caller if needed
	}, nil
}
