// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestMiscCmd(t *testing.T) {
	if miscCmd == nil {
		t.Fatal("miscCmd is nil")
	}

	if miscCmd.Use != "misc" {
		t.Errorf("Use = %s, want misc", miscCmd.Use)
	}

	if miscCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if miscCmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestMiscCmd_Subcommands(t *testing.T) {
	expectedCommands := []string{"ip-ranges", "status"}

	commands := miscCmd.Commands()
	if len(commands) != len(expectedCommands) {
		t.Errorf("Number of subcommands = %d, want %d", len(commands), len(expectedCommands))
	}

	commandMap := make(map[string]*cobra.Command)
	for _, cmd := range commands {
		commandMap[cmd.Use] = cmd
	}

	for _, expected := range expectedCommands {
		if _, ok := commandMap[expected]; !ok {
			t.Errorf("Missing subcommand: %s", expected)
		}
	}
}

func TestMiscIPRangesCmd(t *testing.T) {
	if miscIPRangesCmd == nil {
		t.Fatal("miscIPRangesCmd is nil")
	}

	if miscIPRangesCmd.Use != "ip-ranges" {
		t.Errorf("Use = %s, want ip-ranges", miscIPRangesCmd.Use)
	}

	if miscIPRangesCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if miscIPRangesCmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestMiscStatusCmd(t *testing.T) {
	if miscStatusCmd == nil {
		t.Fatal("miscStatusCmd is nil")
	}

	if miscStatusCmd.Use != "status" {
		t.Errorf("Use = %s, want status", miscStatusCmd.Use)
	}

	if miscStatusCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if miscStatusCmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestMiscCmd_CommandStructure(t *testing.T) {
	tests := []struct {
		name        string
		cmd         *cobra.Command
		wantUse     string
		wantShort   bool
		wantRunE    bool
		wantArgs    bool
	}{
		{
			name:      "ip-ranges command",
			cmd:       miscIPRangesCmd,
			wantUse:   "ip-ranges",
			wantShort: true,
			wantRunE:  true,
			wantArgs:  false,
		},
		{
			name:      "status command",
			cmd:       miscStatusCmd,
			wantUse:   "status",
			wantShort: true,
			wantRunE:  true,
			wantArgs:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cmd == nil {
				t.Fatal("Command is nil")
			}

			if tt.cmd.Use != tt.wantUse {
				t.Errorf("Use = %s, want %s", tt.cmd.Use, tt.wantUse)
			}

			if tt.wantShort && tt.cmd.Short == "" {
				t.Error("Short description is empty")
			}

			if tt.wantRunE && tt.cmd.RunE == nil {
				t.Error("RunE is nil")
			}

			if tt.wantArgs && tt.cmd.Args == nil {
				t.Error("Args validator is nil")
			}
		})
	}
}

func TestMiscCmd_ParentChild(t *testing.T) {
	// Verify parent-child relationships
	commands := miscCmd.Commands()

	for _, cmd := range commands {
		if cmd.Parent() != miscCmd {
			t.Errorf("Command %s parent is not miscCmd", cmd.Use)
		}
	}
}
