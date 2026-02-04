// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"testing"
)

func TestVulnerabilitiesCmd(t *testing.T) {
	if vulnerabilitiesCmd == nil {
		t.Fatal("vulnerabilitiesCmd is nil")
	}

	if vulnerabilitiesCmd.Use != "vulnerabilities" {
		t.Errorf("Use = %s, want vulnerabilities", vulnerabilitiesCmd.Use)
	}

	if vulnerabilitiesCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if vulnerabilitiesCmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestVulnerabilitiesCmd_Subcommands(t *testing.T) {
	expectedCommands := []string{"search", "list"}

	commands := vulnerabilitiesCmd.Commands()

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

func TestVulnerabilitiesSearchCmd(t *testing.T) {
	if vulnerabilitiesSearchCmd == nil {
		t.Fatal("vulnerabilitiesSearchCmd is nil")
	}

	if vulnerabilitiesSearchCmd.Use != "search" {
		t.Errorf("Use = %s, want search", vulnerabilitiesSearchCmd.Use)
	}

	if vulnerabilitiesSearchCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if vulnerabilitiesSearchCmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestVulnerabilitiesListCmd(t *testing.T) {
	if vulnerabilitiesListCmd == nil {
		t.Fatal("vulnerabilitiesListCmd is nil")
	}

	if vulnerabilitiesListCmd.Use != "list" {
		t.Errorf("Use = %s, want list", vulnerabilitiesListCmd.Use)
	}

	if vulnerabilitiesListCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if vulnerabilitiesListCmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestStaticAnalysisCmd(t *testing.T) {
	if staticAnalysisCmd == nil {
		t.Fatal("staticAnalysisCmd is nil")
	}

	if staticAnalysisCmd.Use != "static-analysis" {
		t.Errorf("Use = %s, want static-analysis", staticAnalysisCmd.Use)
	}

	if staticAnalysisCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if staticAnalysisCmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestStaticAnalysisCmd_Subcommands(t *testing.T) {
	expectedCommands := []string{"ast", "custom-rulesets", "sca", "coverage"}

	commands := staticAnalysisCmd.Commands()

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

func TestStaticAnalysisASTCmd(t *testing.T) {
	if staticAnalysisASTCmd == nil {
		t.Fatal("staticAnalysisASTCmd is nil")
	}

	if staticAnalysisASTCmd.Use != "ast" {
		t.Errorf("Use = %s, want ast", staticAnalysisASTCmd.Use)
	}

	if staticAnalysisASTCmd.Short == "" {
		t.Error("Short description is empty")
	}
}

func TestStaticAnalysisCustomRulesetsCmd(t *testing.T) {
	if staticAnalysisCustomRulesetsCmd == nil {
		t.Fatal("staticAnalysisCustomRulesetsCmd is nil")
	}

	if staticAnalysisCustomRulesetsCmd.Use != "custom-rulesets" {
		t.Errorf("Use = %s, want custom-rulesets", staticAnalysisCustomRulesetsCmd.Use)
	}

	if staticAnalysisCustomRulesetsCmd.Short == "" {
		t.Error("Short description is empty")
	}
}

func TestStaticAnalysisSCACmd(t *testing.T) {
	if staticAnalysisSCACmd == nil {
		t.Fatal("staticAnalysisSCACmd is nil")
	}

	if staticAnalysisSCACmd.Use != "sca" {
		t.Errorf("Use = %s, want sca", staticAnalysisSCACmd.Use)
	}

	if staticAnalysisSCACmd.Short == "" {
		t.Error("Short description is empty")
	}
}

func TestStaticAnalysisCoverageCmd(t *testing.T) {
	if staticAnalysisCoverageCmd == nil {
		t.Fatal("staticAnalysisCoverageCmd is nil")
	}

	if staticAnalysisCoverageCmd.Use != "coverage" {
		t.Errorf("Use = %s, want coverage", staticAnalysisCoverageCmd.Use)
	}

	if staticAnalysisCoverageCmd.Short == "" {
		t.Error("Short description is empty")
	}
}

func TestVulnerabilitiesCmd_ParentChild(t *testing.T) {
	commands := vulnerabilitiesCmd.Commands()

	for _, cmd := range commands {
		if cmd.Parent() != vulnerabilitiesCmd {
			t.Errorf("Command %s parent is not vulnerabilitiesCmd", cmd.Use)
		}
	}
}

func TestStaticAnalysisCmd_ParentChild(t *testing.T) {
	commands := staticAnalysisCmd.Commands()

	for _, cmd := range commands {
		if cmd.Parent() != staticAnalysisCmd {
			t.Errorf("Command %s parent is not staticAnalysisCmd", cmd.Use)
		}
	}
}
