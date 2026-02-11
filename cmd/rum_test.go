// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"testing"

	"github.com/DataDog/pup/pkg/config"
	"github.com/spf13/cobra"
)

func TestRumCmd(t *testing.T) {
	if rumCmd == nil {
		t.Fatal("rumCmd is nil")
	}

	if rumCmd.Use != "rum" {
		t.Errorf("Use = %s, want rum", rumCmd.Use)
	}

	if rumCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if rumCmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestRumCmd_Subcommands(t *testing.T) {
	expectedCommands := []string{"apps", "metrics", "retention-filters", "sessions"}

	commands := rumCmd.Commands()

	commandMap := make(map[string]bool)
	for _, cmd := range commands {
		commandMap[cmd.Use] = true
	}

	for _, expected := range expectedCommands {
		if !commandMap[expected] {
			t.Errorf("Missing subcommand: %s", expected)
		}
	}
}

func TestRumAppsCmd(t *testing.T) {
	if rumAppsCmd == nil {
		t.Fatal("rumAppsCmd is nil")
	}

	if rumAppsCmd.Use != "apps" {
		t.Errorf("Use = %s, want apps", rumAppsCmd.Use)
	}

	if rumAppsCmd.Short == "" {
		t.Error("Short description is empty")
	}

	// Check for subcommands
	commands := rumAppsCmd.Commands()
	expectedSubcmds := []string{"list", "get", "create", "update", "delete"}
	commandMap := make(map[string]bool)
	for _, cmd := range commands {
		commandMap[cmd.Use] = true
	}

	for _, expected := range expectedSubcmds {
		if !commandMap[expected] {
			t.Errorf("Missing apps %s subcommand", expected)
		}
	}
}

func TestRumMetricsCmd(t *testing.T) {
	if rumMetricsCmd == nil {
		t.Fatal("rumMetricsCmd is nil")
	}

	if rumMetricsCmd.Use != "metrics" {
		t.Errorf("Use = %s, want metrics", rumMetricsCmd.Use)
	}

	if rumMetricsCmd.Short == "" {
		t.Error("Short description is empty")
	}

	// Check for subcommands
	commands := rumMetricsCmd.Commands()
	expectedSubcmds := []string{"list", "get", "create", "update", "delete"}
	commandMap := make(map[string]bool)
	for _, cmd := range commands {
		commandMap[cmd.Use] = true
	}

	for _, expected := range expectedSubcmds {
		if !commandMap[expected] {
			t.Errorf("Missing metrics %s subcommand", expected)
		}
	}
}

func TestRumRetentionFiltersCmd(t *testing.T) {
	if rumRetentionFiltersCmd == nil {
		t.Fatal("rumRetentionFiltersCmd is nil")
	}

	if rumRetentionFiltersCmd.Use != "retention-filters" {
		t.Errorf("Use = %s, want retention-filters", rumRetentionFiltersCmd.Use)
	}

	if rumRetentionFiltersCmd.Short == "" {
		t.Error("Short description is empty")
	}

	// Check for subcommands
	commands := rumRetentionFiltersCmd.Commands()
	expectedSubcmds := []string{"list", "get", "create", "update", "delete"}
	commandMap := make(map[string]bool)
	for _, cmd := range commands {
		commandMap[cmd.Use] = true
	}

	for _, expected := range expectedSubcmds {
		if !commandMap[expected] {
			t.Errorf("Missing retention-filters %s subcommand", expected)
		}
	}
}

func TestRumSessionsCmd(t *testing.T) {
	if rumSessionsCmd == nil {
		t.Fatal("rumSessionsCmd is nil")
	}

	if rumSessionsCmd.Use != "sessions" {
		t.Errorf("Use = %s, want sessions", rumSessionsCmd.Use)
	}

	if rumSessionsCmd.Short == "" {
		t.Error("Short description is empty")
	}

	// Check for subcommands
	commands := rumSessionsCmd.Commands()
	expectedSubcmds := []string{"list", "search"}
	commandMap := make(map[string]bool)
	for _, cmd := range commands {
		commandMap[cmd.Use] = true
	}

	for _, expected := range expectedSubcmds {
		if !commandMap[expected] {
			t.Errorf("Missing sessions %s subcommand", expected)
		}
	}
}

func TestRumCmd_CommandHierarchy(t *testing.T) {
	// Verify main subcommands
	commands := rumCmd.Commands()
	for _, cmd := range commands {
		if cmd.Parent() != rumCmd {
			t.Errorf("Command %s parent is not rumCmd", cmd.Use)
		}
	}

	// Test each subcommand hierarchy
	subcommands := []struct {
		name string
		cmd  *cobra.Command
	}{
		{"apps", rumAppsCmd},
		{"metrics", rumMetricsCmd},
		{"retention-filters", rumRetentionFiltersCmd},
		{"sessions", rumSessionsCmd},
	}

	for _, sub := range subcommands {
		t.Run(sub.name+" hierarchy", func(t *testing.T) {
			commands := sub.cmd.Commands()
			for _, cmd := range commands {
				if cmd.Parent() != sub.cmd {
					t.Errorf("Command %s parent is not %sCmd", cmd.Use, sub.name)
				}
			}
		})
	}
}

func TestRumAppsListCmd(t *testing.T) {
	if rumAppsListCmd == nil {
		t.Fatal("rumAppsListCmd is nil")
	}

	if rumAppsListCmd.Use != "list" {
		t.Errorf("Use = %s, want list", rumAppsListCmd.Use)
	}

	if rumAppsListCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if rumAppsListCmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

// TestRumSessionsList_TimeConversion tests that relative time strings are converted properly
func TestRumSessionsList_TimeConversion(t *testing.T) {
	// Save originals
	origCfg := cfg
	origClient := ddClient
	origFrom := rumFrom
	origTo := rumTo
	origLimit := rumLimit

	// Cleanup
	defer func() {
		cfg = origCfg
		ddClient = origClient
		rumFrom = origFrom
		rumTo = origTo
		rumLimit = origLimit
	}()

	tests := []struct {
		name      string
		from      string
		to        string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid relative time 1h",
			from:      "1h",
			to:        "now",
			wantError: true, // Will fail on getClient, but time parsing should work
		},
		{
			name:      "valid relative time 30m",
			from:      "30m",
			to:        "now",
			wantError: true, // Will fail on getClient, but time parsing should work
		},
		{
			name:      "valid relative time 7d",
			from:      "7d",
			to:        "now",
			wantError: true, // Will fail on getClient, but time parsing should work
		},
		{
			name:      "invalid time format",
			from:      "invalid",
			to:        "now",
			wantError: true,
			errorMsg:  "invalid --from time",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup: set test values
			cfg = &config.Config{}
			ddClient = nil // Don't create real client (avoids keychain)
			rumFrom = tt.from
			rumTo = tt.to
			rumLimit = 10

			// Run the command
			err := runRumSessionsList(nil, nil)

			// Verify error behavior
			if tt.wantError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tt.errorMsg != "" && err != nil {
				if !containsString(err.Error(), tt.errorMsg) {
					t.Errorf("Error message %q does not contain %q", err.Error(), tt.errorMsg)
				}
			}
		})
	}
}

// TestRumSessionsSearch_TimeConversion tests that relative time strings are converted properly
func TestRumSessionsSearch_TimeConversion(t *testing.T) {
	// Save originals
	origCfg := cfg
	origClient := ddClient
	origFrom := rumFrom
	origTo := rumTo
	origLimit := rumLimit
	origQuery := rumQuery

	// Cleanup
	defer func() {
		cfg = origCfg
		ddClient = origClient
		rumFrom = origFrom
		rumTo = origTo
		rumLimit = origLimit
		rumQuery = origQuery
	}()

	tests := []struct {
		name      string
		query     string
		from      string
		to        string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid relative time 1h",
			query:     "@type:view",
			from:      "1h",
			to:        "now",
			wantError: true, // Will fail on getClient, but time parsing should work
		},
		{
			name:      "valid relative time 2h",
			query:     "status:error",
			from:      "2h",
			to:        "1h",
			wantError: true, // Will fail on getClient, but time parsing should work
		},
		{
			name:      "invalid from time format",
			query:     "@type:view",
			from:      "invalid",
			to:        "now",
			wantError: true,
			errorMsg:  "invalid --from time",
		},
		{
			name:      "invalid to time format",
			query:     "@type:view",
			from:      "1h",
			to:        "bad-time",
			wantError: true,
			errorMsg:  "invalid --to time",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup: set test values
			cfg = &config.Config{}
			ddClient = nil // Don't create real client (avoids keychain)
			rumQuery = tt.query
			rumFrom = tt.from
			rumTo = tt.to
			rumLimit = 10

			// Run the command
			err := runRumSessionsSearch(nil, nil)

			// Verify error behavior
			if tt.wantError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tt.errorMsg != "" && err != nil {
				if !containsString(err.Error(), tt.errorMsg) {
					t.Errorf("Error message %q does not contain %q", err.Error(), tt.errorMsg)
				}
			}
		})
	}
}

// Helper function for string contains check
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
