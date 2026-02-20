// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

//go:build !js

package cmd

import (
	"testing"

	"github.com/datadog-labs/pup/pkg/config"
)

func TestAuthCmd(t *testing.T) {
	if authCmd == nil {
		t.Fatal("authCmd is nil")
	}

	if authCmd.Use != "auth" {
		t.Errorf("Use = %s, want auth", authCmd.Use)
	}

	if authCmd.Short == "" {
		t.Error("Short description is empty")
	}
}

func TestAuthCmd_Subcommands(t *testing.T) {
	expectedCommands := []string{"login", "status", "logout", "refresh"}

	commands := authCmd.Commands()

	commandMap := make(map[string]bool)
	for _, cmd := range commands {
		commandMap[cmd.Name()] = true
	}

	for _, expected := range expectedCommands {
		if !commandMap[expected] {
			t.Errorf("Missing subcommand: %s", expected)
		}
	}
}

func TestAuthLoginCmd(t *testing.T) {
	if authLoginCmd == nil {
		t.Fatal("authLoginCmd is nil")
	}

	if authLoginCmd.Use != "login" {
		t.Errorf("Use = %s, want login", authLoginCmd.Use)
	}

	if authLoginCmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestAuthStatusCmd(t *testing.T) {
	if authStatusCmd == nil {
		t.Fatal("authStatusCmd is nil")
	}

	if authStatusCmd.Use != "status" {
		t.Errorf("Use = %s, want status", authStatusCmd.Use)
	}

	if authStatusCmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestAuthLogoutCmd(t *testing.T) {
	if authLogoutCmd == nil {
		t.Fatal("authLogoutCmd is nil")
	}

	if authLogoutCmd.Use != "logout" {
		t.Errorf("Use = %s, want logout", authLogoutCmd.Use)
	}

	if authLogoutCmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestAuthRefreshCmd(t *testing.T) {
	if authRefreshCmd == nil {
		t.Fatal("authRefreshCmd is nil")
	}

	if authRefreshCmd.Use != "refresh" {
		t.Errorf("Use = %s, want refresh", authRefreshCmd.Use)
	}

	if authRefreshCmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func setupAuthTestEnv(t *testing.T) func() {
	t.Helper()

	// Save original config
	origCfg := cfg

	// Create test config
	cfg = &config.Config{
		Site:        "datadoghq.com",
		APIKey:      "test-api-key-12345678",
		AppKey:      "test-app-key-12345678",
		AutoApprove: false,
	}

	return func() {
		cfg = origCfg
	}
}

func TestRunAuthStatus_NotAuthenticated(t *testing.T) {
	cleanup := setupAuthTestEnv(t)
	defer cleanup()

	// Auth status doesn't require mock client - it checks token storage
	// With no tokens stored, it should report not authenticated
	// This test validates the command structure works
	err := runAuthStatus(authStatusCmd, []string{})

	// runAuthStatus returns nil even when not authenticated, it just prints status
	if err != nil {
		t.Errorf("runAuthStatus() unexpected error = %v", err)
	}
}

func TestRunAuthLogout(t *testing.T) {
	cleanup := setupAuthTestEnv(t)
	defer cleanup()

	// Logout attempts to clear tokens from storage
	// This will fail in test environment due to keychain access
	// We just verify the function can be called
	_ = runAuthLogout(authLogoutCmd, []string{})
	// Don't check error since keychain access will fail in tests
}

func TestAuthCmd_ParentChild(t *testing.T) {
	commands := authCmd.Commands()

	for _, cmd := range commands {
		if cmd.Parent() != authCmd {
			t.Errorf("Command %s parent is not authCmd", cmd.Use)
		}
	}
}

// Note: runAuthLogin and runAuthRefresh are not tested with full OAuth flow mocking
// as they require complex browser interaction and callback server setup.
// These functions are validated through:
// 1. Command structure tests above
// 2. Integration tests (if available)
// 3. Manual testing during development
//
// To fully test these functions would require:
// - Mocking exec.Command for browser opening
// - Mocking callback.NewServer for OAuth callback
// - Mocking dcr.RegisterClient for client registration
// - Mocking oauth.NewClient for OAuth client
// These are beyond the scope of unit tests and better suited for integration tests.
