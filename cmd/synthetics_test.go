// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"testing"
)

func TestSyntheticsCmd(t *testing.T) {
	if syntheticsCmd == nil {
		t.Fatal("syntheticsCmd is nil")
	}

	if syntheticsCmd.Use != "synthetics" {
		t.Errorf("Use = %s, want synthetics", syntheticsCmd.Use)
	}

	if syntheticsCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if syntheticsCmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestSyntheticsCmd_Subcommands(t *testing.T) {
	expectedCommands := []string{"tests", "locations"}

	commands := syntheticsCmd.Commands()

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

func TestSyntheticsTestsCmd(t *testing.T) {
	if syntheticsTestsCmd == nil {
		t.Fatal("syntheticsTestsCmd is nil")
	}

	if syntheticsTestsCmd.Use != "tests" {
		t.Errorf("Use = %s, want tests", syntheticsTestsCmd.Use)
	}

	if syntheticsTestsCmd.Short == "" {
		t.Error("Short description is empty")
	}

	// Check for list and get subcommands
	commands := syntheticsTestsCmd.Commands()
	commandMap := make(map[string]bool)
	for _, cmd := range commands {
		commandMap[cmd.Use] = true
	}

	if !commandMap["list"] {
		t.Error("Missing tests list subcommand")
	}

	// Check if get command exists
	foundGet := false
	for _, cmd := range commands {
		if cmd.Use == "get [test-id]" || cmd.Use == "get" {
			foundGet = true
		}
	}
	if !foundGet {
		t.Error("Missing tests get subcommand")
	}
}

func TestSyntheticsLocationsCmd(t *testing.T) {
	if syntheticsLocationsCmd == nil {
		t.Fatal("syntheticsLocationsCmd is nil")
	}

	if syntheticsLocationsCmd.Use != "locations" {
		t.Errorf("Use = %s, want locations", syntheticsLocationsCmd.Use)
	}

	if syntheticsLocationsCmd.Short == "" {
		t.Error("Short description is empty")
	}

	// Check for list subcommand
	commands := syntheticsLocationsCmd.Commands()
	foundList := false
	for _, cmd := range commands {
		if cmd.Use == "list" {
			foundList = true
		}
	}
	if !foundList {
		t.Error("Missing locations list subcommand")
	}
}

func TestSyntheticsTestsListCmd(t *testing.T) {
	if syntheticsTestsListCmd == nil {
		t.Fatal("syntheticsTestsListCmd is nil")
	}

	if syntheticsTestsListCmd.Use != "list" {
		t.Errorf("Use = %s, want list", syntheticsTestsListCmd.Use)
	}

	if syntheticsTestsListCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if syntheticsTestsListCmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestSyntheticsTestsGetCmd(t *testing.T) {
	if syntheticsTestsGetCmd == nil {
		t.Fatal("syntheticsTestsGetCmd is nil")
	}

	if syntheticsTestsGetCmd.Use != "get [test-id]" {
		t.Errorf("Use = %s, want 'get [test-id]'", syntheticsTestsGetCmd.Use)
	}

	if syntheticsTestsGetCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if syntheticsTestsGetCmd.RunE == nil {
		t.Error("RunE is nil")
	}

	if syntheticsTestsGetCmd.Args == nil {
		t.Error("Args validator is nil")
	}
}

func TestSyntheticsLocationsListCmd(t *testing.T) {
	if syntheticsLocationsListCmd == nil {
		t.Fatal("syntheticsLocationsListCmd is nil")
	}

	if syntheticsLocationsListCmd.Use != "list" {
		t.Errorf("Use = %s, want list", syntheticsLocationsListCmd.Use)
	}

	if syntheticsLocationsListCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if syntheticsLocationsListCmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestSyntheticsCmd_CommandHierarchy(t *testing.T) {
	// Verify parent-child relationships
	commands := syntheticsCmd.Commands()
	for _, cmd := range commands {
		if cmd.Parent() != syntheticsCmd {
			t.Errorf("Command %s parent is not syntheticsCmd", cmd.Use)
		}
	}

	// Verify tests subcommands
	testsCommands := syntheticsTestsCmd.Commands()
	for _, cmd := range testsCommands {
		if cmd.Parent() != syntheticsTestsCmd {
			t.Errorf("Command %s parent is not syntheticsTestsCmd", cmd.Use)
		}
	}

	// Verify locations subcommands
	locationsCommands := syntheticsLocationsCmd.Commands()
	for _, cmd := range locationsCommands {
		if cmd.Parent() != syntheticsLocationsCmd {
			t.Errorf("Command %s parent is not syntheticsLocationsCmd", cmd.Use)
		}
	}
}
