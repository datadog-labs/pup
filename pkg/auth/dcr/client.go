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
func (c *Client) Register(redirectURI string, scopes []string) (*types.ClientCredentials, error) {
	// Build registration request
	req := RegistrationRequest{
		ClientName:    "Datadog Pup CLI",
		RedirectURIs:  []string{redirectURI},
		GrantTypes:    []string{"authorization_code", "refresh_token"},
		ResponseTypes: []string{"code"},
		TokenEndpointAuthMethod: "client_secret_post",
		Scope: strings.Join(scopes, " "),
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

	// Return client credentials
	return &types.ClientCredentials{
		ClientID:     regResp.ClientID,
		ClientSecret: regResp.ClientSecret,
		CreatedAt:    time.Now(),
		Site:         c.site,
	}, nil
}

// ExchangeCode exchanges an authorization code for tokens
func (c *Client) ExchangeCode(code, redirectURI, codeVerifier string, creds *types.ClientCredentials) (*types.TokenSet, error) {
	// Build token request
	req := TokenRequest{
		GrantType:    "authorization_code",
		Code:         code,
		RedirectURI:  redirectURI,
		ClientID:     creds.ClientID,
		ClientSecret: creds.ClientSecret,
		CodeVerifier: codeVerifier,
	}

	return c.requestTokens(req)
}

// RefreshToken refreshes an access token using a refresh token
func (c *Client) RefreshToken(refreshToken string, creds *types.ClientCredentials) (*types.TokenSet, error) {
	// Build token request
	req := TokenRequest{
		GrantType:    "refresh_token",
		RefreshToken: refreshToken,
		ClientID:     creds.ClientID,
		ClientSecret: creds.ClientSecret,
	}

	return c.requestTokens(req)
}

// requestTokens makes a token request to the OAuth2 token endpoint
func (c *Client) requestTokens(req TokenRequest) (*types.TokenSet, error) {
	// Marshal request
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal token request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("https://api.%s/oauth2/v1/token", c.site)
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
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

	// Calculate expiration time
	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	// Return token set
	return &types.TokenSet{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    tokenResp.TokenType,
		ExpiresIn:    tokenResp.ExpiresIn,
		ExpiresAt:    expiresAt,
		Scope:        tokenResp.Scope,
	}, nil
}
