// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"testing"
)

func TestOrganizationsCmd(t *testing.T) {
	if organizationsCmd == nil {
		t.Fatal("organizationsCmd is nil")
	}

	if organizationsCmd.Use != "organizations" {
		t.Errorf("Use = %s, want organizations", organizationsCmd.Use)
	}

	if organizationsCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if organizationsCmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestOrganizationsCmd_Subcommands(t *testing.T) {
	expectedCommands := []string{"get", "list"}

	commands := organizationsCmd.Commands()

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

func TestOrganizationsGetCmd(t *testing.T) {
	if organizationsGetCmd == nil {
		t.Fatal("organizationsGetCmd is nil")
	}

	if organizationsGetCmd.Use != "get" {
		t.Errorf("Use = %s, want get", organizationsGetCmd.Use)
	}

	if organizationsGetCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if organizationsGetCmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestOrganizationsListCmd(t *testing.T) {
	if organizationsListCmd == nil {
		t.Fatal("organizationsListCmd is nil")
	}

	if organizationsListCmd.Use != "list" {
		t.Errorf("Use = %s, want list", organizationsListCmd.Use)
	}

	if organizationsListCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if organizationsListCmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestOrganizationsCmd_ParentChild(t *testing.T) {
	commands := organizationsCmd.Commands()

	for _, cmd := range commands {
		if cmd.Parent() != organizationsCmd {
			t.Errorf("Command %s parent is not organizationsCmd", cmd.Use)
		}
	}
}
