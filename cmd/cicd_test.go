// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"testing"
)

func TestCICDCmd(t *testing.T) {
	if cicdCmd == nil {
		t.Fatal("cicdCmd is nil")
	}

	if cicdCmd.Use != "cicd" {
		t.Errorf("Use = %s, want cicd", cicdCmd.Use)
	}

	if cicdCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cicdCmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestCICDCmd_Subcommands(t *testing.T) {
	expectedCommands := []string{"pipelines", "events"}

	commands := cicdCmd.Commands()

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

func TestCICDPipelinesCmd(t *testing.T) {
	if cicdPipelinesCmd == nil {
		t.Fatal("cicdPipelinesCmd is nil")
	}

	if cicdPipelinesCmd.Use != "pipelines" {
		t.Errorf("Use = %s, want pipelines", cicdPipelinesCmd.Use)
	}

	if cicdPipelinesCmd.Short == "" {
		t.Error("Short description is empty")
	}

	// Check for list and get subcommands
	commands := cicdPipelinesCmd.Commands()
	commandMap := make(map[string]bool)
	for _, cmd := range commands {
		commandMap[cmd.Use] = true
	}

	if !commandMap["list"] {
		t.Error("Missing pipelines list subcommand")
	}

	// Check if get command exists
	foundGet := false
	for _, cmd := range commands {
		if cmd.Use == "get [pipeline-id]" || cmd.Use == "get" {
			foundGet = true
		}
	}
	if !foundGet {
		t.Error("Missing pipelines get subcommand")
	}
}

func TestCICDEventsCmd(t *testing.T) {
	if cicdEventsCmd == nil {
		t.Fatal("cicdEventsCmd is nil")
	}

	if cicdEventsCmd.Use != "events" {
		t.Errorf("Use = %s, want events", cicdEventsCmd.Use)
	}

	if cicdEventsCmd.Short == "" {
		t.Error("Short description is empty")
	}

	// Check for search and aggregate subcommands
	commands := cicdEventsCmd.Commands()
	commandMap := make(map[string]bool)
	for _, cmd := range commands {
		commandMap[cmd.Use] = true
	}

	if !commandMap["search"] {
		t.Error("Missing events search subcommand")
	}

	if !commandMap["aggregate"] {
		t.Error("Missing events aggregate subcommand")
	}
}

func TestCICDPipelinesListCmd(t *testing.T) {
	if cicdPipelinesListCmd == nil {
		t.Fatal("cicdPipelinesListCmd is nil")
	}

	if cicdPipelinesListCmd.Use != "list" {
		t.Errorf("Use = %s, want list", cicdPipelinesListCmd.Use)
	}

	if cicdPipelinesListCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cicdPipelinesListCmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestCICDEventsSearchCmd(t *testing.T) {
	if cicdEventsSearchCmd == nil {
		t.Fatal("cicdEventsSearchCmd is nil")
	}

	if cicdEventsSearchCmd.Use != "search" {
		t.Errorf("Use = %s, want search", cicdEventsSearchCmd.Use)
	}

	if cicdEventsSearchCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cicdEventsSearchCmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestCICDEventsAggregateCmd(t *testing.T) {
	if cicdEventsAggregateCmd == nil {
		t.Fatal("cicdEventsAggregateCmd is nil")
	}

	if cicdEventsAggregateCmd.Use != "aggregate" {
		t.Errorf("Use = %s, want aggregate", cicdEventsAggregateCmd.Use)
	}

	if cicdEventsAggregateCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cicdEventsAggregateCmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestCICDCmd_CommandHierarchy(t *testing.T) {
	// Verify main subcommands
	commands := cicdCmd.Commands()
	for _, cmd := range commands {
		if cmd.Parent() != cicdCmd {
			t.Errorf("Command %s parent is not cicdCmd", cmd.Use)
		}
	}

	// Verify pipelines subcommands
	pipelinesCommands := cicdPipelinesCmd.Commands()
	for _, cmd := range pipelinesCommands {
		if cmd.Parent() != cicdPipelinesCmd {
			t.Errorf("Command %s parent is not cicdPipelinesCmd", cmd.Use)
		}
	}

	// Verify events subcommands
	eventsCommands := cicdEventsCmd.Commands()
	for _, cmd := range eventsCommands {
		if cmd.Parent() != cicdEventsCmd {
			t.Errorf("Command %s parent is not cicdEventsCmd", cmd.Use)
		}
	}
}
