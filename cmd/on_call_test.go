// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"testing"
)

func TestOnCallCmd(t *testing.T) {
	if onCallCmd == nil {
		t.Fatal("onCallCmd is nil")
	}

	if onCallCmd.Use != "on-call" {
		t.Errorf("Use = %s, want on-call", onCallCmd.Use)
	}

	if onCallCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if onCallCmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestOnCallCmd_Subcommands(t *testing.T) {
	// Check that teams subcommand exists
	commands := onCallCmd.Commands()

	foundTeams := false
	for _, cmd := range commands {
		if cmd.Use == "teams" {
			foundTeams = true
		}
	}

	if !foundTeams {
		t.Error("Missing teams subcommand")
	}
}

func TestOnCallTeamsCmd(t *testing.T) {
	if onCallTeamsCmd == nil {
		t.Fatal("onCallTeamsCmd is nil")
	}

	if onCallTeamsCmd.Use != "teams" {
		t.Errorf("Use = %s, want teams", onCallTeamsCmd.Use)
	}

	if onCallTeamsCmd.Short == "" {
		t.Error("Short description is empty")
	}

	// Check for list and get subcommands
	commands := onCallTeamsCmd.Commands()
	commandMap := make(map[string]bool)
	for _, cmd := range commands {
		commandMap[cmd.Use] = true
	}

	if !commandMap["list"] {
		t.Error("Missing teams list subcommand")
	}

	// Check if get command exists
	foundGet := false
	for _, cmd := range commands {
		if cmd.Use == "get [team-id]" || cmd.Use == "get" {
			foundGet = true
		}
	}
	if !foundGet {
		t.Error("Missing teams get subcommand")
	}
}

func TestOnCallTeamsListCmd(t *testing.T) {
	if onCallTeamsListCmd == nil {
		t.Fatal("onCallTeamsListCmd is nil")
	}

	if onCallTeamsListCmd.Use != "list" {
		t.Errorf("Use = %s, want list", onCallTeamsListCmd.Use)
	}

	if onCallTeamsListCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if onCallTeamsListCmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestOnCallTeamsGetCmd(t *testing.T) {
	if onCallTeamsGetCmd == nil {
		t.Fatal("onCallTeamsGetCmd is nil")
	}

	if onCallTeamsGetCmd.Use != "get [team-id]" {
		t.Errorf("Use = %s, want 'get [team-id]'", onCallTeamsGetCmd.Use)
	}

	if onCallTeamsGetCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if onCallTeamsGetCmd.RunE == nil {
		t.Error("RunE is nil")
	}

	if onCallTeamsGetCmd.Args == nil {
		t.Error("Args validator is nil")
	}
}

func TestOnCallCmd_CommandHierarchy(t *testing.T) {
	// Verify teams is a subcommand of on-call
	commands := onCallCmd.Commands()
	foundTeams := false
	for _, cmd := range commands {
		if cmd.Use == "teams" {
			foundTeams = true
			if cmd.Parent() != onCallCmd {
				t.Error("teams parent is not onCallCmd")
			}
		}
	}
	if !foundTeams {
		t.Error("teams subcommand not found in on-call")
	}

	// Verify list and get are subcommands of teams
	teamsCommands := onCallTeamsCmd.Commands()
	for _, cmd := range teamsCommands {
		if cmd.Parent() != onCallTeamsCmd {
			t.Errorf("Command %s parent is not onCallTeamsCmd", cmd.Use)
		}
	}
}
