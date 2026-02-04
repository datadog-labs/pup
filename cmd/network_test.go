// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"testing"
)

func TestNetworkCmd(t *testing.T) {
	if networkCmd == nil {
		t.Fatal("networkCmd is nil")
	}

	if networkCmd.Use != "network" {
		t.Errorf("Use = %s, want network", networkCmd.Use)
	}

	if networkCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if networkCmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestNetworkCmd_Subcommands(t *testing.T) {
	expectedCommands := []string{"flows", "devices"}

	commands := networkCmd.Commands()

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

func TestNetworkFlowsCmd(t *testing.T) {
	if networkFlowsCmd == nil {
		t.Fatal("networkFlowsCmd is nil")
	}

	if networkFlowsCmd.Use != "flows" {
		t.Errorf("Use = %s, want flows", networkFlowsCmd.Use)
	}

	if networkFlowsCmd.Short == "" {
		t.Error("Short description is empty")
	}

	// Check for list subcommand
	commands := networkFlowsCmd.Commands()
	foundList := false
	for _, cmd := range commands {
		if cmd.Use == "list" {
			foundList = true
			if cmd.RunE == nil {
				t.Error("Flows list command RunE is nil")
			}
		}
	}
	if !foundList {
		t.Error("Missing flows list subcommand")
	}
}

func TestNetworkDevicesCmd(t *testing.T) {
	if networkDevicesCmd == nil {
		t.Fatal("networkDevicesCmd is nil")
	}

	if networkDevicesCmd.Use != "devices" {
		t.Errorf("Use = %s, want devices", networkDevicesCmd.Use)
	}

	if networkDevicesCmd.Short == "" {
		t.Error("Short description is empty")
	}

	// Check for list subcommand
	commands := networkDevicesCmd.Commands()
	foundList := false
	for _, cmd := range commands {
		if cmd.Use == "list" {
			foundList = true
			if cmd.RunE == nil {
				t.Error("Devices list command RunE is nil")
			}
		}
	}
	if !foundList {
		t.Error("Missing devices list subcommand")
	}
}

func TestNetworkFlowsListCmd(t *testing.T) {
	if networkFlowsListCmd == nil {
		t.Fatal("networkFlowsListCmd is nil")
	}

	if networkFlowsListCmd.Use != "list" {
		t.Errorf("Use = %s, want list", networkFlowsListCmd.Use)
	}

	if networkFlowsListCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if networkFlowsListCmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestNetworkDevicesListCmd(t *testing.T) {
	if networkDevicesListCmd == nil {
		t.Fatal("networkDevicesListCmd is nil")
	}

	if networkDevicesListCmd.Use != "list" {
		t.Errorf("Use = %s, want list", networkDevicesListCmd.Use)
	}

	if networkDevicesListCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if networkDevicesListCmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestNetworkCmd_CommandHierarchy(t *testing.T) {
	// Verify main subcommands
	commands := networkCmd.Commands()
	for _, cmd := range commands {
		if cmd.Parent() != networkCmd {
			t.Errorf("Command %s parent is not networkCmd", cmd.Use)
		}
	}

	// Verify flows subcommands
	flowsCommands := networkFlowsCmd.Commands()
	for _, cmd := range flowsCommands {
		if cmd.Parent() != networkFlowsCmd {
			t.Errorf("Command %s parent is not networkFlowsCmd", cmd.Use)
		}
	}

	// Verify devices subcommands
	devicesCommands := networkDevicesCmd.Commands()
	for _, cmd := range devicesCommands {
		if cmd.Parent() != networkDevicesCmd {
			t.Errorf("Command %s parent is not networkDevicesCmd", cmd.Use)
		}
	}
}
