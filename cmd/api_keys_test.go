// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"testing"
)

func TestAPIKeysCmd(t *testing.T) {
	if apiKeysCmd == nil {
		t.Fatal("apiKeysCmd is nil")
	}

	if apiKeysCmd.Use != "api-keys" {
		t.Errorf("Use = %s, want api-keys", apiKeysCmd.Use)
	}

	if apiKeysCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if apiKeysCmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestAPIKeysCmd_Subcommands(t *testing.T) {
	expectedCommands := []string{"list", "get", "create", "delete"}

	commands := apiKeysCmd.Commands()

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

func TestAPIKeysListCmd(t *testing.T) {
	if apiKeysListCmd == nil {
		t.Fatal("apiKeysListCmd is nil")
	}

	if apiKeysListCmd.Use != "list" {
		t.Errorf("Use = %s, want list", apiKeysListCmd.Use)
	}

	if apiKeysListCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if apiKeysListCmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestAPIKeysGetCmd(t *testing.T) {
	if apiKeysGetCmd == nil {
		t.Fatal("apiKeysGetCmd is nil")
	}

	if apiKeysGetCmd.Use != "get [key-id]" {
		t.Errorf("Use = %s, want 'get [key-id]'", apiKeysGetCmd.Use)
	}

	if apiKeysGetCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if apiKeysGetCmd.RunE == nil {
		t.Error("RunE is nil")
	}

	if apiKeysGetCmd.Args == nil {
		t.Error("Args validator is nil")
	}
}

func TestAPIKeysCreateCmd(t *testing.T) {
	if apiKeysCreateCmd == nil {
		t.Fatal("apiKeysCreateCmd is nil")
	}

	if apiKeysCreateCmd.Use != "create" {
		t.Errorf("Use = %s, want create", apiKeysCreateCmd.Use)
	}

	if apiKeysCreateCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if apiKeysCreateCmd.RunE == nil {
		t.Error("RunE is nil")
	}

	// Check for flags
	flags := apiKeysCreateCmd.Flags()
	if flags.Lookup("name") == nil {
		t.Error("Missing --name flag")
	}
}

func TestAPIKeysDeleteCmd(t *testing.T) {
	if apiKeysDeleteCmd == nil {
		t.Fatal("apiKeysDeleteCmd is nil")
	}

	if apiKeysDeleteCmd.Use != "delete [key-id]" {
		t.Errorf("Use = %s, want 'delete [key-id]'", apiKeysDeleteCmd.Use)
	}

	if apiKeysDeleteCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if apiKeysDeleteCmd.RunE == nil {
		t.Error("RunE is nil")
	}

	if apiKeysDeleteCmd.Args == nil {
		t.Error("Args validator is nil")
	}
}

func TestAPIKeysCmd_ParentChild(t *testing.T) {
	commands := apiKeysCmd.Commands()

	for _, cmd := range commands {
		if cmd.Parent() != apiKeysCmd {
			t.Errorf("Command %s parent is not apiKeysCmd", cmd.Use)
		}
	}
}
