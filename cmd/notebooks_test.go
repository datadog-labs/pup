// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"testing"
)

func TestNotebooksCmd(t *testing.T) {
	if notebooksCmd == nil {
		t.Fatal("notebooksCmd is nil")
	}

	if notebooksCmd.Use != "notebooks" {
		t.Errorf("Use = %s, want notebooks", notebooksCmd.Use)
	}

	if notebooksCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if notebooksCmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestNotebooksCmd_Subcommands(t *testing.T) {
	expectedCommands := []string{"list", "get", "delete"}

	commands := notebooksCmd.Commands()

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

func TestNotebooksListCmd(t *testing.T) {
	if notebooksListCmd == nil {
		t.Fatal("notebooksListCmd is nil")
	}

	if notebooksListCmd.Use != "list" {
		t.Errorf("Use = %s, want list", notebooksListCmd.Use)
	}

	if notebooksListCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if notebooksListCmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestNotebooksGetCmd(t *testing.T) {
	if notebooksGetCmd == nil {
		t.Fatal("notebooksGetCmd is nil")
	}

	if notebooksGetCmd.Use != "get [notebook-id]" {
		t.Errorf("Use = %s, want 'get [notebook-id]'", notebooksGetCmd.Use)
	}

	if notebooksGetCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if notebooksGetCmd.RunE == nil {
		t.Error("RunE is nil")
	}

	if notebooksGetCmd.Args == nil {
		t.Error("Args validator is nil")
	}
}

func TestNotebooksDeleteCmd(t *testing.T) {
	if notebooksDeleteCmd == nil {
		t.Fatal("notebooksDeleteCmd is nil")
	}

	if notebooksDeleteCmd.Use != "delete [notebook-id]" {
		t.Errorf("Use = %s, want 'delete [notebook-id]'", notebooksDeleteCmd.Use)
	}

	if notebooksDeleteCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if notebooksDeleteCmd.RunE == nil {
		t.Error("RunE is nil")
	}

	if notebooksDeleteCmd.Args == nil {
		t.Error("Args validator is nil")
	}
}

func TestNotebooksCmd_ParentChild(t *testing.T) {
	commands := notebooksCmd.Commands()

	for _, cmd := range commands {
		if cmd.Parent() != notebooksCmd {
			t.Errorf("Command %s parent is not notebooksCmd", cmd.Use)
		}
	}
}
