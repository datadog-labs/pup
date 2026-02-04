// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"testing"
)

func TestObsPipelinesCmd(t *testing.T) {
	if obsPipelinesCmd == nil {
		t.Fatal("obsPipelinesCmd is nil")
	}

	if obsPipelinesCmd.Use != "obs-pipelines" {
		t.Errorf("Use = %s, want obs-pipelines", obsPipelinesCmd.Use)
	}

	if obsPipelinesCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if obsPipelinesCmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestObsPipelinesCmd_Subcommands(t *testing.T) {
	expectedCommands := []string{"list", "get"}

	commands := obsPipelinesCmd.Commands()

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

func TestObsPipelinesListCmd(t *testing.T) {
	if obsPipelinesListCmd == nil {
		t.Fatal("obsPipelinesListCmd is nil")
	}

	if obsPipelinesListCmd.Use != "list" {
		t.Errorf("Use = %s, want list", obsPipelinesListCmd.Use)
	}

	if obsPipelinesListCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if obsPipelinesListCmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestObsPipelinesGetCmd(t *testing.T) {
	if obsPipelinesGetCmd == nil {
		t.Fatal("obsPipelinesGetCmd is nil")
	}

	if obsPipelinesGetCmd.Use != "get [pipeline-id]" {
		t.Errorf("Use = %s, want 'get [pipeline-id]'", obsPipelinesGetCmd.Use)
	}

	if obsPipelinesGetCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if obsPipelinesGetCmd.RunE == nil {
		t.Error("RunE is nil")
	}

	if obsPipelinesGetCmd.Args == nil {
		t.Error("Args validator is nil")
	}
}

func TestObsPipelinesCmd_ParentChild(t *testing.T) {
	commands := obsPipelinesCmd.Commands()

	for _, cmd := range commands {
		if cmd.Parent() != obsPipelinesCmd {
			t.Errorf("Command %s parent is not obsPipelinesCmd", cmd.Use)
		}
	}
}
