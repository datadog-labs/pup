// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"testing"
)

func TestInfrastructureCmd(t *testing.T) {
	if infrastructureCmd == nil {
		t.Fatal("infrastructureCmd is nil")
	}

	if infrastructureCmd.Use != "infrastructure" {
		t.Errorf("Use = %s, want infrastructure", infrastructureCmd.Use)
	}

	if infrastructureCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if infrastructureCmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestInfrastructureCmd_Subcommands(t *testing.T) {
	// Check that hosts subcommand exists
	commands := infrastructureCmd.Commands()

	foundHosts := false
	for _, cmd := range commands {
		if cmd.Use == "hosts" {
			foundHosts = true
		}
	}

	if !foundHosts {
		t.Error("Missing hosts subcommand")
	}
}

func TestInfrastructureHostsCmd(t *testing.T) {
	if infrastructureHostsCmd == nil {
		t.Fatal("infrastructureHostsCmd is nil")
	}

	if infrastructureHostsCmd.Use != "hosts" {
		t.Errorf("Use = %s, want hosts", infrastructureHostsCmd.Use)
	}

	if infrastructureHostsCmd.Short == "" {
		t.Error("Short description is empty")
	}

	// Check for list and get subcommands
	commands := infrastructureHostsCmd.Commands()
	commandMap := make(map[string]bool)
	for _, cmd := range commands {
		commandMap[cmd.Use] = true
	}

	if !commandMap["list"] {
		t.Error("Missing hosts list subcommand")
	}

	// Check if get command exists (format could be "get" or "get [hostname]")
	foundGet := false
	for _, cmd := range commands {
		if cmd.Use == "get [hostname]" || cmd.Use == "get" {
			foundGet = true
		}
	}
	if !foundGet {
		t.Error("Missing hosts get subcommand")
	}
}

func TestInfrastructureHostsListCmd(t *testing.T) {
	if infrastructureHostsListCmd == nil {
		t.Fatal("infrastructureHostsListCmd is nil")
	}

	if infrastructureHostsListCmd.Use != "list" {
		t.Errorf("Use = %s, want list", infrastructureHostsListCmd.Use)
	}

	if infrastructureHostsListCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if infrastructureHostsListCmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestInfrastructureHostsGetCmd(t *testing.T) {
	if infrastructureHostsGetCmd == nil {
		t.Fatal("infrastructureHostsGetCmd is nil")
	}

	if infrastructureHostsGetCmd.Use != "get [hostname]" {
		t.Errorf("Use = %s, want 'get [hostname]'", infrastructureHostsGetCmd.Use)
	}

	if infrastructureHostsGetCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if infrastructureHostsGetCmd.RunE == nil {
		t.Error("RunE is nil")
	}

	if infrastructureHostsGetCmd.Args == nil {
		t.Error("Args validator is nil")
	}
}

func TestInfrastructureCmd_CommandHierarchy(t *testing.T) {
	// Verify hosts is a subcommand of infrastructure
	commands := infrastructureCmd.Commands()
	foundHosts := false
	for _, cmd := range commands {
		if cmd.Use == "hosts" {
			foundHosts = true
			if cmd.Parent() != infrastructureCmd {
				t.Error("hosts parent is not infrastructureCmd")
			}
		}
	}
	if !foundHosts {
		t.Error("hosts subcommand not found in infrastructure")
	}

	// Verify list and get are subcommands of hosts
	hostsCommands := infrastructureHostsCmd.Commands()
	for _, cmd := range hostsCommands {
		if cmd.Parent() != infrastructureHostsCmd {
			t.Errorf("Command %s parent is not infrastructureHostsCmd", cmd.Use)
		}
	}
}
