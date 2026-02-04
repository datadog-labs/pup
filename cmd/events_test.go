// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"testing"
)

func TestEventsCmd(t *testing.T) {
	if eventsCmd == nil {
		t.Fatal("eventsCmd is nil")
	}

	if eventsCmd.Use != "events" {
		t.Errorf("Use = %s, want events", eventsCmd.Use)
	}

	if eventsCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if eventsCmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestEventsCmd_Subcommands(t *testing.T) {
	expectedCommands := []string{"list", "search", "get"}

	commands := eventsCmd.Commands()
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

func TestEventsListCmd(t *testing.T) {
	if eventsListCmd == nil {
		t.Fatal("eventsListCmd is nil")
	}

	if eventsListCmd.Use != "list" {
		t.Errorf("Use = %s, want list", eventsListCmd.Use)
	}

	if eventsListCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if eventsListCmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestEventsSearchCmd(t *testing.T) {
	if eventsSearchCmd == nil {
		t.Fatal("eventsSearchCmd is nil")
	}

	if eventsSearchCmd.Use != "search" {
		t.Errorf("Use = %s, want search", eventsSearchCmd.Use)
	}

	if eventsSearchCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if eventsSearchCmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestEventsGetCmd(t *testing.T) {
	if eventsGetCmd == nil {
		t.Fatal("eventsGetCmd is nil")
	}

	if eventsGetCmd.Use != "get [event-id]" {
		t.Errorf("Use = %s, want 'get [event-id]'", eventsGetCmd.Use)
	}

	if eventsGetCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if eventsGetCmd.RunE == nil {
		t.Error("RunE is nil")
	}

	if eventsGetCmd.Args == nil {
		t.Error("Args validator is nil")
	}
}

func TestEventsCmd_ParentChild(t *testing.T) {
	commands := eventsCmd.Commands()

	for _, cmd := range commands {
		if cmd.Parent() != eventsCmd {
			t.Errorf("Command %s parent is not eventsCmd", cmd.Use)
		}
	}
}
