// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"testing"
)

func TestDataGovernanceCmd(t *testing.T) {
	if dataGovernanceCmd == nil {
		t.Fatal("dataGovernanceCmd is nil")
	}

	if dataGovernanceCmd.Use != "data-governance" {
		t.Errorf("Use = %s, want data-governance", dataGovernanceCmd.Use)
	}

	if dataGovernanceCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if dataGovernanceCmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestDataGovernanceCmd_Subcommands(t *testing.T) {
	// Check that scanner-rules subcommand exists
	commands := dataGovernanceCmd.Commands()

	foundScannerRules := false
	for _, cmd := range commands {
		if cmd.Use == "scanner-rules" {
			foundScannerRules = true
		}
	}

	if !foundScannerRules {
		t.Error("Missing scanner-rules subcommand")
	}
}

func TestDataGovernanceScannerRulesCmd(t *testing.T) {
	if dataGovernanceScannerRulesCmd == nil {
		t.Fatal("dataGovernanceScannerRulesCmd is nil")
	}

	if dataGovernanceScannerRulesCmd.Use != "scanner-rules" {
		t.Errorf("Use = %s, want scanner-rules", dataGovernanceScannerRulesCmd.Use)
	}

	if dataGovernanceScannerRulesCmd.Short == "" {
		t.Error("Short description is empty")
	}

	// Check for list subcommand
	commands := dataGovernanceScannerRulesCmd.Commands()
	foundList := false
	for _, cmd := range commands {
		if cmd.Use == "list" {
			foundList = true
			if cmd.RunE == nil {
				t.Error("Scanner rules list command RunE is nil")
			}
		}
	}
	if !foundList {
		t.Error("Missing scanner-rules list subcommand")
	}
}

func TestDataGovernanceScannerRulesListCmd(t *testing.T) {
	if dataGovernanceScannerRulesListCmd == nil {
		t.Fatal("dataGovernanceScannerRulesListCmd is nil")
	}

	if dataGovernanceScannerRulesListCmd.Use != "list" {
		t.Errorf("Use = %s, want list", dataGovernanceScannerRulesListCmd.Use)
	}

	if dataGovernanceScannerRulesListCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if dataGovernanceScannerRulesListCmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestDataGovernanceCmd_CommandHierarchy(t *testing.T) {
	// Verify scanner-rules is a subcommand of data-governance
	commands := dataGovernanceCmd.Commands()
	foundScannerRules := false
	for _, cmd := range commands {
		if cmd.Use == "scanner-rules" {
			foundScannerRules = true
			if cmd.Parent() != dataGovernanceCmd {
				t.Error("scanner-rules parent is not dataGovernanceCmd")
			}
		}
	}
	if !foundScannerRules {
		t.Error("scanner-rules subcommand not found in data-governance")
	}

	// Verify list is a subcommand of scanner-rules
	scannerRulesCommands := dataGovernanceScannerRulesCmd.Commands()
	for _, cmd := range scannerRulesCommands {
		if cmd.Parent() != dataGovernanceScannerRulesCmd {
			t.Errorf("Command %s parent is not dataGovernanceScannerRulesCmd", cmd.Use)
		}
	}
}
