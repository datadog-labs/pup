// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"testing"
)

func TestErrorTrackingCmd(t *testing.T) {
	if errorTrackingCmd == nil {
		t.Fatal("errorTrackingCmd is nil")
	}

	if errorTrackingCmd.Use != "error-tracking" {
		t.Errorf("Use = %s, want error-tracking", errorTrackingCmd.Use)
	}

	if errorTrackingCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if errorTrackingCmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestErrorTrackingCmd_Subcommands(t *testing.T) {
	// Check that issues subcommand exists
	commands := errorTrackingCmd.Commands()

	foundIssues := false
	for _, cmd := range commands {
		if cmd.Use == "issues" {
			foundIssues = true
		}
	}

	if !foundIssues {
		t.Error("Missing issues subcommand")
	}
}

func TestErrorTrackingIssuesCmd(t *testing.T) {
	if errorTrackingIssuesCmd == nil {
		t.Fatal("errorTrackingIssuesCmd is nil")
	}

	if errorTrackingIssuesCmd.Use != "issues" {
		t.Errorf("Use = %s, want issues", errorTrackingIssuesCmd.Use)
	}

	if errorTrackingIssuesCmd.Short == "" {
		t.Error("Short description is empty")
	}

	// Check for list and get subcommands
	commands := errorTrackingIssuesCmd.Commands()
	commandMap := make(map[string]bool)
	for _, cmd := range commands {
		commandMap[cmd.Use] = true
	}

	if !commandMap["list"] {
		t.Error("Missing issues list subcommand")
	}

	// Check if get command exists
	foundGet := false
	for _, cmd := range commands {
		if cmd.Use == "get [issue-id]" || cmd.Use == "get" {
			foundGet = true
		}
	}
	if !foundGet {
		t.Error("Missing issues get subcommand")
	}
}

func TestErrorTrackingIssuesListCmd(t *testing.T) {
	if errorTrackingIssuesListCmd == nil {
		t.Fatal("errorTrackingIssuesListCmd is nil")
	}

	if errorTrackingIssuesListCmd.Use != "list" {
		t.Errorf("Use = %s, want list", errorTrackingIssuesListCmd.Use)
	}

	if errorTrackingIssuesListCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if errorTrackingIssuesListCmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestErrorTrackingIssuesGetCmd(t *testing.T) {
	if errorTrackingIssuesGetCmd == nil {
		t.Fatal("errorTrackingIssuesGetCmd is nil")
	}

	if errorTrackingIssuesGetCmd.Use != "get [issue-id]" {
		t.Errorf("Use = %s, want 'get [issue-id]'", errorTrackingIssuesGetCmd.Use)
	}

	if errorTrackingIssuesGetCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if errorTrackingIssuesGetCmd.RunE == nil {
		t.Error("RunE is nil")
	}

	if errorTrackingIssuesGetCmd.Args == nil {
		t.Error("Args validator is nil")
	}
}

func TestErrorTrackingCmd_CommandHierarchy(t *testing.T) {
	// Verify issues is a subcommand of error-tracking
	commands := errorTrackingCmd.Commands()
	foundIssues := false
	for _, cmd := range commands {
		if cmd.Use == "issues" {
			foundIssues = true
			if cmd.Parent() != errorTrackingCmd {
				t.Error("issues parent is not errorTrackingCmd")
			}
		}
	}
	if !foundIssues {
		t.Error("issues subcommand not found in error-tracking")
	}

	// Verify list and get are subcommands of issues
	issuesCommands := errorTrackingIssuesCmd.Commands()
	for _, cmd := range issuesCommands {
		if cmd.Parent() != errorTrackingIssuesCmd {
			t.Errorf("Command %s parent is not errorTrackingIssuesCmd", cmd.Use)
		}
	}
}
