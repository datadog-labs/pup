// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestCloudCmd(t *testing.T) {
	if cloudCmd == nil {
		t.Fatal("cloudCmd is nil")
	}

	if cloudCmd.Use != "cloud" {
		t.Errorf("Use = %s, want cloud", cloudCmd.Use)
	}

	if cloudCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cloudCmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestCloudCmd_Subcommands(t *testing.T) {
	expectedCommands := []string{"aws", "gcp", "azure"}

	commands := cloudCmd.Commands()
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

func TestCloudAWSCmd(t *testing.T) {
	if cloudAWSCmd == nil {
		t.Fatal("cloudAWSCmd is nil")
	}

	if cloudAWSCmd.Use != "aws" {
		t.Errorf("Use = %s, want aws", cloudAWSCmd.Use)
	}

	if cloudAWSCmd.Short == "" {
		t.Error("Short description is empty")
	}

	// Check for list subcommand
	commands := cloudAWSCmd.Commands()
	foundList := false
	for _, cmd := range commands {
		if cmd.Use == "list" {
			foundList = true
			if cmd.RunE == nil {
				t.Error("AWS list command RunE is nil")
			}
		}
	}
	if !foundList {
		t.Error("AWS list subcommand not found")
	}
}

func TestCloudGCPCmd(t *testing.T) {
	if cloudGCPCmd == nil {
		t.Fatal("cloudGCPCmd is nil")
	}

	if cloudGCPCmd.Use != "gcp" {
		t.Errorf("Use = %s, want gcp", cloudGCPCmd.Use)
	}

	if cloudGCPCmd.Short == "" {
		t.Error("Short description is empty")
	}

	// Check for list subcommand
	commands := cloudGCPCmd.Commands()
	foundList := false
	for _, cmd := range commands {
		if cmd.Use == "list" {
			foundList = true
			if cmd.RunE == nil {
				t.Error("GCP list command RunE is nil")
			}
		}
	}
	if !foundList {
		t.Error("GCP list subcommand not found")
	}
}

func TestCloudAzureCmd(t *testing.T) {
	if cloudAzureCmd == nil {
		t.Fatal("cloudAzureCmd is nil")
	}

	if cloudAzureCmd.Use != "azure" {
		t.Errorf("Use = %s, want azure", cloudAzureCmd.Use)
	}

	if cloudAzureCmd.Short == "" {
		t.Error("Short description is empty")
	}

	// Check for list subcommand
	commands := cloudAzureCmd.Commands()
	foundList := false
	for _, cmd := range commands {
		if cmd.Use == "list" {
			foundList = true
			if cmd.RunE == nil {
				t.Error("Azure list command RunE is nil")
			}
		}
	}
	if !foundList {
		t.Error("Azure list subcommand not found")
	}
}

func TestCloudCmd_CommandHierarchy(t *testing.T) {
	tests := []struct {
		name           string
		parentCmd      *cobra.Command
		parentUse      string
		subcommandUse  string
	}{
		{
			name:          "AWS subcommand",
			parentCmd:     cloudCmd,
			parentUse:     "cloud",
			subcommandUse: "aws",
		},
		{
			name:          "GCP subcommand",
			parentCmd:     cloudCmd,
			parentUse:     "cloud",
			subcommandUse: "gcp",
		},
		{
			name:          "Azure subcommand",
			parentCmd:     cloudCmd,
			parentUse:     "cloud",
			subcommandUse: "azure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commands := tt.parentCmd.Commands()
			found := false
			for _, cmd := range commands {
				if cmd.Use == tt.subcommandUse {
					found = true
					if cmd.Parent() != tt.parentCmd {
						t.Errorf("Parent of %s is not %s", tt.subcommandUse, tt.parentUse)
					}
				}
			}
			if !found {
				t.Errorf("Subcommand %s not found in %s", tt.subcommandUse, tt.parentUse)
			}
		})
	}
}

func TestCloudCmd_ListCommands(t *testing.T) {
	cloudProviders := []struct {
		name string
		cmd  *cobra.Command
	}{
		{"AWS", cloudAWSCmd},
		{"GCP", cloudGCPCmd},
		{"Azure", cloudAzureCmd},
	}

	for _, provider := range cloudProviders {
		t.Run(provider.name+" list command", func(t *testing.T) {
			commands := provider.cmd.Commands()
			foundList := false
			for _, cmd := range commands {
				if cmd.Use == "list" {
					foundList = true
					if cmd.Short == "" {
						t.Errorf("%s list command has no short description", provider.name)
					}
					if cmd.RunE == nil {
						t.Errorf("%s list command has no RunE", provider.name)
					}
				}
			}
			if !foundList {
				t.Errorf("%s has no list subcommand", provider.name)
			}
		})
	}
}
