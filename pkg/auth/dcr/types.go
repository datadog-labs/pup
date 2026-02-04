// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package dcr

import (
	"fmt"
)

// Constants from TypeScript PR #84 for compatibility
const (
	// DCRClientName is the client name used during registration
	// Matches datadog-api-claude-plugin for compatibility
	DCRClientName = "datadog-api-claude-plugin"
)

// DCRRedirectPorts are the specific ports to register for OAuth callbacks
// Must match TypeScript PR #84 for compatibility
var DCRRedirectPorts = []int{8000, 8080, 8888, 9000}

// GetRedirectURIs returns the standard redirect URIs for the given ports
func GetRedirectURIs() []string {
	uris := make([]string, len(DCRRedirectPorts))
	for i, port := range DCRRedirectPorts {
		uris[i] = fmt.Sprintf("http://127.0.0.1:%d/oauth/callback", port)
	}
	return uris
}

// RegistrationRequest represents a DCR registration request (RFC 7591)
// Matches TypeScript PR #84 format for compatibility
type RegistrationRequest struct {
	ClientName   string   `json:"client_name"`
	RedirectURIs []string `json:"redirect_uris"`
	GrantTypes   []string `json:"grant_types"`
}

// RegistrationResponse represents a DCR registration response
// Matches TypeScript PR #84 format for compatibility
// Note: Public clients (token_endpoint_auth_method: 'none') don't receive client_secret
type RegistrationResponse struct {
	ClientID                string   `json:"client_id"`
	ClientName              string   `json:"client_name"`
	RedirectURIs            []string `json:"redirect_uris"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
	GrantTypes              []string `json:"grant_types"`
	Scope                   string   `json:"scope,omitempty"`
}

// TokenRequest represents an OAuth2 token request
type TokenRequest struct {
	GrantType    string `json:"grant_type"`
	Code         string `json:"code,omitempty"`
	RedirectURI  string `json:"redirect_uri,omitempty"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	CodeVerifier string `json:"code_verifier,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// TokenResponse represents an OAuth2 token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}
