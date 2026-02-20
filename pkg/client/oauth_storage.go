// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

//go:build !js

package client

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/datadog-labs/pup/pkg/auth/dcr"
	"github.com/datadog-labs/pup/pkg/auth/storage"
	"github.com/datadog-labs/pup/pkg/config"
)

// Test hooks — overridden in tests to inject fakes
var (
	getStorageFunc   = func() (storage.Storage, error) { return storage.GetStorage(nil) }
	newDCRClientFunc = func(site string) *dcr.Client { return dcr.NewClient(site) }
)

// tryOAuthFromStorage attempts to load OAuth tokens from local storage and
// returns a context with an access token set. If the stored token is expired
// and a refresh token is available, it attempts to auto-refresh. Returns nil
// if no valid token could be obtained.
func tryOAuthFromStorage(cfg *config.Config) context.Context {
	store, err := getStorageFunc()
	if err != nil {
		return nil
	}
	tokens, err := store.LoadTokens(cfg.Site)
	if err != nil || tokens == nil {
		return nil
	}

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
		return context.WithValue(
			context.Background(),
			datadog.ContextAccessToken,
			tokens.AccessToken,
		)
	}
	return nil
}
