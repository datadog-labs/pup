// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"testing"
)

func TestAuditLogsCmd(t *testing.T) {
	if auditLogsCmd == nil {
		t.Fatal("auditLogsCmd is nil")
	}

	if auditLogsCmd.Use != "audit-logs" {
		t.Errorf("Use = %s, want audit-logs", auditLogsCmd.Use)
	}

	if auditLogsCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if auditLogsCmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestAuditLogsCmd_Subcommands(t *testing.T) {
	expectedCommands := []string{"list", "search"}

	commands := auditLogsCmd.Commands()

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

func TestAuditLogsListCmd(t *testing.T) {
	if auditLogsListCmd == nil {
		t.Fatal("auditLogsListCmd is nil")
	}

	if auditLogsListCmd.Use != "list" {
		t.Errorf("Use = %s, want list", auditLogsListCmd.Use)
	}

	if auditLogsListCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if auditLogsListCmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestAuditLogsSearchCmd(t *testing.T) {
	if auditLogsSearchCmd == nil {
		t.Fatal("auditLogsSearchCmd is nil")
	}

	if auditLogsSearchCmd.Use != "search" {
		t.Errorf("Use = %s, want search", auditLogsSearchCmd.Use)
	}

	if auditLogsSearchCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if auditLogsSearchCmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestAuditLogsCmd_ParentChild(t *testing.T) {
	commands := auditLogsCmd.Commands()

	for _, cmd := range commands {
		if cmd.Parent() != auditLogsCmd {
			t.Errorf("Command %s parent is not auditLogsCmd", cmd.Use)
		}
	}
}
