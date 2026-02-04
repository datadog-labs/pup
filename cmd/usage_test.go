// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"testing"
)

func TestUsageCmd(t *testing.T) {
	if usageCmd == nil {
		t.Fatal("usageCmd is nil")
	}

	if usageCmd.Use != "usage" {
		t.Errorf("Use = %s, want usage", usageCmd.Use)
	}

	if usageCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if usageCmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestUsageCmd_Subcommands(t *testing.T) {
	expectedCommands := []string{"summary", "hourly"}

	commands := usageCmd.Commands()

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

func TestUsageSummaryCmd(t *testing.T) {
	if usageSummaryCmd == nil {
		t.Fatal("usageSummaryCmd is nil")
	}

	if usageSummaryCmd.Use != "summary" {
		t.Errorf("Use = %s, want summary", usageSummaryCmd.Use)
	}

	if usageSummaryCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if usageSummaryCmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestUsageHourlyCmd(t *testing.T) {
	if usageHourlyCmd == nil {
		t.Fatal("usageHourlyCmd is nil")
	}

	if usageHourlyCmd.Use != "hourly" {
		t.Errorf("Use = %s, want hourly", usageHourlyCmd.Use)
	}

	if usageHourlyCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if usageHourlyCmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestUsageCmd_ParentChild(t *testing.T) {
	commands := usageCmd.Commands()

	for _, cmd := range commands {
		if cmd.Parent() != usageCmd {
			t.Errorf("Command %s parent is not usageCmd", cmd.Use)
		}
	}
}
