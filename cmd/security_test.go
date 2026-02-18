// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"testing"
)

func TestSecurityCmd(t *testing.T) {
	if securityCmd == nil {
		t.Fatal("securityCmd is nil")
	}

	if securityCmd.Use != "security" {
		t.Errorf("Use = %s, want security", securityCmd.Use)
	}

	if securityCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if securityCmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestSecurityCmd_Subcommands(t *testing.T) {
	expectedCommands := []string{"rules", "signals", "findings", "content-packs", "risk-scores"}

	commands := securityCmd.Commands()

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

func TestSecurityRulesCmd(t *testing.T) {
	if securityRulesCmd == nil {
		t.Fatal("securityRulesCmd is nil")
	}

	if securityRulesCmd.Use != "rules" {
		t.Errorf("Use = %s, want rules", securityRulesCmd.Use)
	}

	if securityRulesCmd.Short == "" {
		t.Error("Short description is empty")
	}

	// Check for list and get subcommands
	commands := securityRulesCmd.Commands()
	commandMap := make(map[string]bool)
	for _, cmd := range commands {
		commandMap[cmd.Use] = true
	}

	if !commandMap["list"] {
		t.Error("Missing rules list subcommand")
	}

	// Check if get command exists
	foundGet := false
	for _, cmd := range commands {
		if cmd.Use == "get [rule-id]" || cmd.Use == "get" {
			foundGet = true
		}
	}
	if !foundGet {
		t.Error("Missing rules get subcommand")
	}
}

func TestSecuritySignalsCmd(t *testing.T) {
	if securitySignalsCmd == nil {
		t.Fatal("securitySignalsCmd is nil")
	}

	if securitySignalsCmd.Use != "signals" {
		t.Errorf("Use = %s, want signals", securitySignalsCmd.Use)
	}

	if securitySignalsCmd.Short == "" {
		t.Error("Short description is empty")
	}

	// Check for list subcommand
	commands := securitySignalsCmd.Commands()
	foundList := false
	for _, cmd := range commands {
		if cmd.Use == "list" {
			foundList = true
		}
	}
	if !foundList {
		t.Error("Missing signals list subcommand")
	}
}

func TestSecurityFindingsCmd(t *testing.T) {
	if securityFindingsCmd == nil {
		t.Fatal("securityFindingsCmd is nil")
	}

	if securityFindingsCmd.Use != "findings" {
		t.Errorf("Use = %s, want findings", securityFindingsCmd.Use)
	}

	if securityFindingsCmd.Short == "" {
		t.Error("Short description is empty")
	}

	// Check for search subcommand
	commands := securityFindingsCmd.Commands()
	foundSearch := false
	for _, cmd := range commands {
		if cmd.Use == "search" {
			foundSearch = true
		}
	}
	if !foundSearch {
		t.Error("Missing findings search subcommand")
	}
}

func TestSecurityCmd_CommandHierarchy(t *testing.T) {
	// Verify parent-child relationships
	commands := securityCmd.Commands()
	for _, cmd := range commands {
		if cmd.Parent() != securityCmd {
			t.Errorf("Command %s parent is not securityCmd", cmd.Use)
		}
	}

	// Verify rules subcommands
	rulesCommands := securityRulesCmd.Commands()
	for _, cmd := range rulesCommands {
		if cmd.Parent() != securityRulesCmd {
			t.Errorf("Command %s parent is not securityRulesCmd", cmd.Use)
		}
	}

	// Verify signals subcommands
	signalsCommands := securitySignalsCmd.Commands()
	for _, cmd := range signalsCommands {
		if cmd.Parent() != securitySignalsCmd {
			t.Errorf("Command %s parent is not securitySignalsCmd", cmd.Use)
		}
	}

	// Verify findings subcommands
	findingsCommands := securityFindingsCmd.Commands()
	for _, cmd := range findingsCommands {
		if cmd.Parent() != securityFindingsCmd {
			t.Errorf("Command %s parent is not securityFindingsCmd", cmd.Use)
		}
	}
}
