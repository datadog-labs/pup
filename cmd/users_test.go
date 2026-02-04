// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"testing"
)

func TestUsersCmd(t *testing.T) {
	if usersCmd == nil {
		t.Fatal("usersCmd is nil")
	}

	if usersCmd.Use != "users" {
		t.Errorf("Use = %s, want users", usersCmd.Use)
	}

	if usersCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if usersCmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestUsersCmd_Subcommands(t *testing.T) {
	expectedCommands := []string{"list", "get", "roles"}

	commands := usersCmd.Commands()

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

func TestUsersListCmd(t *testing.T) {
	if usersListCmd == nil {
		t.Fatal("usersListCmd is nil")
	}

	if usersListCmd.Use != "list" {
		t.Errorf("Use = %s, want list", usersListCmd.Use)
	}

	if usersListCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if usersListCmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestUsersGetCmd(t *testing.T) {
	if usersGetCmd == nil {
		t.Fatal("usersGetCmd is nil")
	}

	if usersGetCmd.Use != "get [user-id]" {
		t.Errorf("Use = %s, want 'get [user-id]'", usersGetCmd.Use)
	}

	if usersGetCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if usersGetCmd.RunE == nil {
		t.Error("RunE is nil")
	}

	if usersGetCmd.Args == nil {
		t.Error("Args validator is nil")
	}
}

func TestUsersRolesCmd(t *testing.T) {
	if usersRolesCmd == nil {
		t.Fatal("usersRolesCmd is nil")
	}

	if usersRolesCmd.Use != "roles" {
		t.Errorf("Use = %s, want roles", usersRolesCmd.Use)
	}

	if usersRolesCmd.Short == "" {
		t.Error("Short description is empty")
	}

	// Check for list subcommand
	commands := usersRolesCmd.Commands()
	foundList := false
	for _, cmd := range commands {
		if cmd.Use == "list" {
			foundList = true
			if cmd.RunE == nil {
				t.Error("Roles list command RunE is nil")
			}
		}
	}
	if !foundList {
		t.Error("Missing roles list subcommand")
	}
}

func TestUsersCmd_ParentChild(t *testing.T) {
	// Verify parent-child relationships
	commands := usersCmd.Commands()
	for _, cmd := range commands {
		if cmd.Parent() != usersCmd {
			t.Errorf("Command %s parent is not usersCmd", cmd.Use)
		}
	}

	// Verify roles subcommands
	rolesCommands := usersRolesCmd.Commands()
	for _, cmd := range rolesCommands {
		if cmd.Parent() != usersRolesCmd {
			t.Errorf("Command %s parent is not usersRolesCmd", cmd.Use)
		}
	}
}
