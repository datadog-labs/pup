// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"testing"
)

func TestScorecardsCmd(t *testing.T) {
	if scorecardsCmd == nil {
		t.Fatal("scorecardsCmd is nil")
	}

	if scorecardsCmd.Use != "scorecards" {
		t.Errorf("Use = %s, want scorecards", scorecardsCmd.Use)
	}

	if scorecardsCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if scorecardsCmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestScorecardsCmd_Subcommands(t *testing.T) {
	expectedCommands := []string{"list", "get"}

	commands := scorecardsCmd.Commands()

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

func TestScorecardsListCmd(t *testing.T) {
	if scorecardsListCmd == nil {
		t.Fatal("scorecardsListCmd is nil")
	}

	if scorecardsListCmd.Use != "list" {
		t.Errorf("Use = %s, want list", scorecardsListCmd.Use)
	}

	if scorecardsListCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if scorecardsListCmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestScorecardsGetCmd(t *testing.T) {
	if scorecardsGetCmd == nil {
		t.Fatal("scorecardsGetCmd is nil")
	}

	if scorecardsGetCmd.Use != "get [scorecard-id]" {
		t.Errorf("Use = %s, want 'get [scorecard-id]'", scorecardsGetCmd.Use)
	}

	if scorecardsGetCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if scorecardsGetCmd.RunE == nil {
		t.Error("RunE is nil")
	}

	if scorecardsGetCmd.Args == nil {
		t.Error("Args validator is nil")
	}
}

func TestScorecardsCmd_ParentChild(t *testing.T) {
	commands := scorecardsCmd.Commands()

	for _, cmd := range commands {
		if cmd.Parent() != scorecardsCmd {
			t.Errorf("Command %s parent is not scorecardsCmd", cmd.Use)
		}
	}
}
