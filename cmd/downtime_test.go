// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"testing"
)

func TestDowntimeCmd(t *testing.T) {
	if downtimeCmd == nil {
		t.Fatal("downtimeCmd is nil")
	}

	if downtimeCmd.Use != "downtime" {
		t.Errorf("Use = %s, want downtime", downtimeCmd.Use)
	}

	if downtimeCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if downtimeCmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestDowntimeCmd_Subcommands(t *testing.T) {
	expectedCommands := []string{"list", "get", "cancel"}

	commands := downtimeCmd.Commands()
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

func TestDowntimeListCmd(t *testing.T) {
	if downtimeListCmd == nil {
		t.Fatal("downtimeListCmd is nil")
	}

	if downtimeListCmd.Use != "list" {
		t.Errorf("Use = %s, want list", downtimeListCmd.Use)
	}

	if downtimeListCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if downtimeListCmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestDowntimeGetCmd(t *testing.T) {
	if downtimeGetCmd == nil {
		t.Fatal("downtimeGetCmd is nil")
	}

	if downtimeGetCmd.Use != "get [downtime-id]" {
		t.Errorf("Use = %s, want 'get [downtime-id]'", downtimeGetCmd.Use)
	}

	if downtimeGetCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if downtimeGetCmd.RunE == nil {
		t.Error("RunE is nil")
	}

	if downtimeGetCmd.Args == nil {
		t.Error("Args validator is nil")
	}
}

func TestDowntimeCancelCmd(t *testing.T) {
	if downtimeCancelCmd == nil {
		t.Fatal("downtimeCancelCmd is nil")
	}

	if downtimeCancelCmd.Use != "cancel [downtime-id]" {
		t.Errorf("Use = %s, want 'cancel [downtime-id]'", downtimeCancelCmd.Use)
	}

	if downtimeCancelCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if downtimeCancelCmd.RunE == nil {
		t.Error("RunE is nil")
	}

	if downtimeCancelCmd.Args == nil {
		t.Error("Args validator is nil")
	}
}

func TestDowntimeCmd_ParentChild(t *testing.T) {
	commands := downtimeCmd.Commands()

	for _, cmd := range commands {
		if cmd.Parent() != downtimeCmd {
			t.Errorf("Command %s parent is not downtimeCmd", cmd.Use)
		}
	}
}
