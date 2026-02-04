// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"testing"

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
