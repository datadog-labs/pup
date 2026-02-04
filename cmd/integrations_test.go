// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestIntegrationsCmd(t *testing.T) {
	if integrationsCmd == nil {
		t.Fatal("integrationsCmd is nil")
	}

	if integrationsCmd.Use != "integrations" {
		t.Errorf("Use = %s, want integrations", integrationsCmd.Use)
	}

	if integrationsCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if integrationsCmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestIntegrationsCmd_Subcommands(t *testing.T) {
	expectedCommands := []string{"slack", "pagerduty", "webhooks"}

	commands := integrationsCmd.Commands()
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

func TestIntegrationsSlackCmd(t *testing.T) {
	if integrationsSlackCmd == nil {
		t.Fatal("integrationsSlackCmd is nil")
	}

	if integrationsSlackCmd.Use != "slack" {
		t.Errorf("Use = %s, want slack", integrationsSlackCmd.Use)
	}

	if integrationsSlackCmd.Short == "" {
		t.Error("Short description is empty")
	}

	// Check for list subcommand
	commands := integrationsSlackCmd.Commands()
	foundList := false
	for _, cmd := range commands {
		if cmd.Use == "list" {
			foundList = true
			if cmd.RunE == nil {
				t.Error("Slack list command RunE is nil")
			}
		}
	}
	if !foundList {
		t.Error("Slack list subcommand not found")
	}
}

func TestIntegrationsPagerDutyCmd(t *testing.T) {
	if integrationsPagerDutyCmd == nil {
		t.Fatal("integrationsPagerDutyCmd is nil")
	}

	if integrationsPagerDutyCmd.Use != "pagerduty" {
		t.Errorf("Use = %s, want pagerduty", integrationsPagerDutyCmd.Use)
	}

	if integrationsPagerDutyCmd.Short == "" {
		t.Error("Short description is empty")
	}

	// Check for list subcommand
	commands := integrationsPagerDutyCmd.Commands()
	foundList := false
	for _, cmd := range commands {
		if cmd.Use == "list" {
			foundList = true
			if cmd.RunE == nil {
				t.Error("PagerDuty list command RunE is nil")
			}
		}
	}
	if !foundList {
		t.Error("PagerDuty list subcommand not found")
	}
}

func TestIntegrationsWebhooksCmd(t *testing.T) {
	if integrationsWebhooksCmd == nil {
		t.Fatal("integrationsWebhooksCmd is nil")
	}

	if integrationsWebhooksCmd.Use != "webhooks" {
		t.Errorf("Use = %s, want webhooks", integrationsWebhooksCmd.Use)
	}

	if integrationsWebhooksCmd.Short == "" {
		t.Error("Short description is empty")
	}

	// Check for list subcommand
	commands := integrationsWebhooksCmd.Commands()
	foundList := false
	for _, cmd := range commands {
		if cmd.Use == "list" {
			foundList = true
			if cmd.RunE == nil {
				t.Error("Webhooks list command RunE is nil")
			}
		}
	}
	if !foundList {
		t.Error("Webhooks list subcommand not found")
	}
}

func TestIntegrationsCmd_CommandHierarchy(t *testing.T) {
	tests := []struct {
		name           string
		parentCmd      *cobra.Command
		parentUse      string
		subcommandUse  string
	}{
		{
			name:          "Slack subcommand",
			parentCmd:     integrationsCmd,
			parentUse:     "integrations",
			subcommandUse: "slack",
		},
		{
			name:          "PagerDuty subcommand",
			parentCmd:     integrationsCmd,
			parentUse:     "integrations",
			subcommandUse: "pagerduty",
		},
		{
			name:          "Webhooks subcommand",
			parentCmd:     integrationsCmd,
			parentUse:     "integrations",
			subcommandUse: "webhooks",
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

func TestIntegrationsCmd_ListCommands(t *testing.T) {
	integrations := []struct {
		name string
		cmd  *cobra.Command
	}{
		{"Slack", integrationsSlackCmd},
		{"PagerDuty", integrationsPagerDutyCmd},
		{"Webhooks", integrationsWebhooksCmd},
	}

	for _, integration := range integrations {
		t.Run(integration.name+" list command", func(t *testing.T) {
			commands := integration.cmd.Commands()
			foundList := false
			for _, cmd := range commands {
				if cmd.Use == "list" {
					foundList = true
					if cmd.Short == "" {
						t.Errorf("%s list command has no short description", integration.name)
					}
					if cmd.RunE == nil {
						t.Errorf("%s list command has no RunE", integration.name)
					}
				}
			}
			if !foundList {
				t.Errorf("%s has no list subcommand", integration.name)
			}
		})
	}
}
