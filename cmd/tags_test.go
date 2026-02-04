// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"testing"
)

func TestTagsCmd(t *testing.T) {
	if tagsCmd == nil {
		t.Fatal("tagsCmd is nil")
	}

	if tagsCmd.Use != "tags" {
		t.Errorf("Use = %s, want tags", tagsCmd.Use)
	}

	if tagsCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if tagsCmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestTagsCmd_Subcommands(t *testing.T) {
	expectedCommands := []string{"list", "get", "add", "update", "delete"}

	commands := tagsCmd.Commands()
	if len(commands) != len(expectedCommands) {
		t.Errorf("Number of subcommands = %d, want %d", len(commands), len(expectedCommands))
	}

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

func TestTagsListCmd(t *testing.T) {
	if tagsListCmd == nil {
		t.Fatal("tagsListCmd is nil")
	}

	if tagsListCmd.Use != "list" {
		t.Errorf("Use = %s, want list", tagsListCmd.Use)
	}

	if tagsListCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if tagsListCmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestTagsGetCmd(t *testing.T) {
	if tagsGetCmd == nil {
		t.Fatal("tagsGetCmd is nil")
	}

	if tagsGetCmd.Use != "get [hostname]" {
		t.Errorf("Use = %s, want 'get [hostname]'", tagsGetCmd.Use)
	}

	if tagsGetCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if tagsGetCmd.RunE == nil {
		t.Error("RunE is nil")
	}

	if tagsGetCmd.Args == nil {
		t.Error("Args validator is nil")
	}
}

func TestTagsAddCmd(t *testing.T) {
	if tagsAddCmd == nil {
		t.Fatal("tagsAddCmd is nil")
	}

	if tagsAddCmd.Use != "add [hostname]" {
		t.Errorf("Use = %s, want 'add [hostname]'", tagsAddCmd.Use)
	}

	if tagsAddCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if tagsAddCmd.RunE == nil {
		t.Error("RunE is nil")
	}

	if tagsAddCmd.Args == nil {
		t.Error("Args validator is nil")
	}
}

func TestTagsUpdateCmd(t *testing.T) {
	if tagsUpdateCmd == nil {
		t.Fatal("tagsUpdateCmd is nil")
	}

	if tagsUpdateCmd.Use != "update [hostname]" {
		t.Errorf("Use = %s, want 'update [hostname]'", tagsUpdateCmd.Use)
	}

	if tagsUpdateCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if tagsUpdateCmd.RunE == nil {
		t.Error("RunE is nil")
	}

	if tagsUpdateCmd.Args == nil {
		t.Error("Args validator is nil")
	}
}

func TestTagsDeleteCmd(t *testing.T) {
	if tagsDeleteCmd == nil {
		t.Fatal("tagsDeleteCmd is nil")
	}

	if tagsDeleteCmd.Use != "delete [hostname]" {
		t.Errorf("Use = %s, want 'delete [hostname]'", tagsDeleteCmd.Use)
	}

	if tagsDeleteCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if tagsDeleteCmd.RunE == nil {
		t.Error("RunE is nil")
	}

	if tagsDeleteCmd.Args == nil {
		t.Error("Args validator is nil")
	}
}

func TestTagsCmd_ParentChild(t *testing.T) {
	commands := tagsCmd.Commands()

	for _, cmd := range commands {
		if cmd.Parent() != tagsCmd {
			t.Errorf("Command %s parent is not tagsCmd", cmd.Use)
		}
	}
}
