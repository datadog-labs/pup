// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestAPMCmd(t *testing.T) {
	tests := []struct {
		name     string
		cmd      *cobra.Command
		expected string
	}{
		{
			name:     "apm command",
			cmd:      apmCmd,
			expected: "apm",
		},
		{
			name:     "services command",
			cmd:      apmServicesCmd,
			expected: "services",
		},
		{
			name:     "services list command",
			cmd:      apmServicesListCmd,
			expected: "list",
		},
		{
			name:     "services stats command",
			cmd:      apmServicesStatsCmd,
			expected: "stats",
		},
		{
			name:     "services operations command",
			cmd:      apmServicesOperationsCmd,
			expected: "operations <service>",
		},
		{
			name:     "services resources command",
			cmd:      apmServicesResourcesCmd,
			expected: "resources <service>",
		},
		{
			name:     "entities command",
			cmd:      apmEntitiesCmd,
			expected: "entities",
		},
		{
			name:     "entities list command",
			cmd:      apmEntitiesListCmd,
			expected: "list",
		},
		{
			name:     "dependencies command",
			cmd:      apmDependenciesCmd,
			expected: "dependencies",
		},
		{
			name:     "dependencies list command",
			cmd:      apmDependenciesListCmd,
			expected: "list [service]",
		},
		{
			name:     "flow-map command",
			cmd:      apmFlowMapCmd,
			expected: "flow-map",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cmd.Use != tt.expected {
				t.Errorf("expected Use to be %q, got %q", tt.expected, tt.cmd.Use)
			}
		})
	}
}

func TestAPMCmdStructure(t *testing.T) {
	// Test root command
	if apmCmd.Use != "apm" {
		t.Errorf("expected apm command Use to be 'apm', got %q", apmCmd.Use)
	}
	if apmCmd.Short == "" {
		t.Error("apm command Short description is empty")
	}
	if apmCmd.Long == "" {
		t.Error("apm command Long description is empty")
	}

	// Test services command
	if apmServicesCmd.Use != "services" {
		t.Errorf("expected services command Use to be 'services', got %q", apmServicesCmd.Use)
	}
	if apmServicesCmd.Short == "" {
		t.Error("services command Short description is empty")
	}

	// Test entities command
	if apmEntitiesCmd.Use != "entities" {
		t.Errorf("expected entities command Use to be 'entities', got %q", apmEntitiesCmd.Use)
	}
	if apmEntitiesCmd.Short == "" {
		t.Error("entities command Short description is empty")
	}

	// Test dependencies command
	if apmDependenciesCmd.Use != "dependencies" {
		t.Errorf("expected dependencies command Use to be 'dependencies', got %q", apmDependenciesCmd.Use)
	}
	if apmDependenciesCmd.Short == "" {
		t.Error("dependencies command Short description is empty")
	}

	// Test flow-map command
	if apmFlowMapCmd.Use != "flow-map" {
		t.Errorf("expected flow-map command Use to be 'flow-map', got %q", apmFlowMapCmd.Use)
	}
	if apmFlowMapCmd.Short == "" {
		t.Error("flow-map command Short description is empty")
	}
}

func TestAPMSubcommands(t *testing.T) {
	// Test that apm has correct subcommands
	expectedSubcommands := []string{"services", "entities", "dependencies", "flow-map"}
	actualSubcommands := make(map[string]bool)
	for _, cmd := range apmCmd.Commands() {
		actualSubcommands[cmd.Use] = true
	}

	for _, expected := range expectedSubcommands {
		if !actualSubcommands[expected] {
			t.Errorf("apm command missing subcommand: %s", expected)
		}
	}

	// Test that services has correct subcommands
	expectedServicesSubcommands := []string{"list", "stats", "operations <service>", "resources <service>"}
	actualServicesSubcommands := make(map[string]bool)
	for _, cmd := range apmServicesCmd.Commands() {
		actualServicesSubcommands[cmd.Use] = true
	}

	for _, expected := range expectedServicesSubcommands {
		if !actualServicesSubcommands[expected] {
			t.Errorf("services command missing subcommand: %s", expected)
		}
	}

	// Test that entities has correct subcommands
	expectedEntitiesSubcommands := []string{"list"}
	actualEntitiesSubcommands := make(map[string]bool)
	for _, cmd := range apmEntitiesCmd.Commands() {
		actualEntitiesSubcommands[cmd.Use] = true
	}

	for _, expected := range expectedEntitiesSubcommands {
		if !actualEntitiesSubcommands[expected] {
			t.Errorf("entities command missing subcommand: %s", expected)
		}
	}

	// Test that dependencies has correct subcommands
	expectedDependenciesSubcommands := []string{"list [service]"}
	actualDependenciesSubcommands := make(map[string]bool)
	for _, cmd := range apmDependenciesCmd.Commands() {
		actualDependenciesSubcommands[cmd.Use] = true
	}

	for _, expected := range expectedDependenciesSubcommands {
		if !actualDependenciesSubcommands[expected] {
			t.Errorf("dependencies command missing subcommand: %s", expected)
		}
	}
}

func TestAPMFlags(t *testing.T) {
	// Test services stats flags
	statsFlags := apmServicesStatsCmd.Flags()
	if statsFlags.Lookup("env") == nil {
		t.Error("services stats command missing --env flag")
	}
	if statsFlags.Lookup("primary-tag") == nil {
		t.Error("services stats command missing --primary-tag flag")
	}
	if statsFlags.Lookup("start") == nil {
		t.Error("services stats command missing --start flag")
	}
	if statsFlags.Lookup("end") == nil {
		t.Error("services stats command missing --end flag")
	}

	// Test services operations flags
	opsFlags := apmServicesOperationsCmd.Flags()
	if opsFlags.Lookup("env") == nil {
		t.Error("services operations command missing --env flag")
	}
	if opsFlags.Lookup("primary-tag") == nil {
		t.Error("services operations command missing --primary-tag flag")
	}
	if opsFlags.Lookup("primary-only") == nil {
		t.Error("services operations command missing --primary-only flag")
	}
	if opsFlags.Lookup("start") == nil {
		t.Error("services operations command missing --start flag")
	}
	if opsFlags.Lookup("end") == nil {
		t.Error("services operations command missing --end flag")
	}

	// Test services resources flags
	resourcesFlags := apmServicesResourcesCmd.Flags()
	if resourcesFlags.Lookup("operation") == nil {
		t.Error("services resources command missing --operation flag")
	}
	if resourcesFlags.Lookup("env") == nil {
		t.Error("services resources command missing --env flag")
	}
	if resourcesFlags.Lookup("from") == nil {
		t.Error("services resources command missing --from flag")
	}
	if resourcesFlags.Lookup("to") == nil {
		t.Error("services resources command missing --to flag")
	}

	// Test entities list flags
	entitiesFlags := apmEntitiesListCmd.Flags()
	if entitiesFlags.Lookup("env") == nil {
		t.Error("entities list command missing --env flag")
	}
	if entitiesFlags.Lookup("types") == nil {
		t.Error("entities list command missing --types flag")
	}
	if entitiesFlags.Lookup("include") == nil {
		t.Error("entities list command missing --include flag")
	}
	if entitiesFlags.Lookup("limit") == nil {
		t.Error("entities list command missing --limit flag")
	}
	if entitiesFlags.Lookup("offset") == nil {
		t.Error("entities list command missing --offset flag")
	}
	if entitiesFlags.Lookup("start") == nil {
		t.Error("entities list command missing --start flag")
	}
	if entitiesFlags.Lookup("end") == nil {
		t.Error("entities list command missing --end flag")
	}

	// Test dependencies list flags
	depsFlags := apmDependenciesListCmd.Flags()
	if depsFlags.Lookup("env") == nil {
		t.Error("dependencies list command missing --env flag")
	}
	if depsFlags.Lookup("primary-tag") == nil {
		t.Error("dependencies list command missing --primary-tag flag")
	}
	if depsFlags.Lookup("start") == nil {
		t.Error("dependencies list command missing --start flag")
	}
	if depsFlags.Lookup("end") == nil {
		t.Error("dependencies list command missing --end flag")
	}

	// Test flow-map flags
	flowMapFlags := apmFlowMapCmd.Flags()
	if flowMapFlags.Lookup("query") == nil {
		t.Error("flow-map command missing --query flag")
	}
	if flowMapFlags.Lookup("limit") == nil {
		t.Error("flow-map command missing --limit flag")
	}
	if flowMapFlags.Lookup("from") == nil {
		t.Error("flow-map command missing --from flag")
	}
	if flowMapFlags.Lookup("to") == nil {
		t.Error("flow-map command missing --to flag")
	}
}

func TestAPMRequiredFlags(t *testing.T) {
	// Test that required flags are marked as required
	tests := []struct {
		name         string
		cmd          *cobra.Command
		requiredFlag string
	}{
		{
			name:         "services stats start flag required",
			cmd:          apmServicesStatsCmd,
			requiredFlag: "start",
		},
		{
			name:         "services stats end flag required",
			cmd:          apmServicesStatsCmd,
			requiredFlag: "end",
		},
		{
			name:         "services operations start flag required",
			cmd:          apmServicesOperationsCmd,
			requiredFlag: "start",
		},
		{
			name:         "services operations end flag required",
			cmd:          apmServicesOperationsCmd,
			requiredFlag: "end",
		},
		{
			name:         "services resources operation flag required",
			cmd:          apmServicesResourcesCmd,
			requiredFlag: "operation",
		},
		{
			name:         "services resources from flag required",
			cmd:          apmServicesResourcesCmd,
			requiredFlag: "from",
		},
		{
			name:         "services resources to flag required",
			cmd:          apmServicesResourcesCmd,
			requiredFlag: "to",
		},
		{
			name:         "entities list start flag required",
			cmd:          apmEntitiesListCmd,
			requiredFlag: "start",
		},
		{
			name:         "entities list end flag required",
			cmd:          apmEntitiesListCmd,
			requiredFlag: "end",
		},
		{
			name:         "dependencies list env flag required",
			cmd:          apmDependenciesListCmd,
			requiredFlag: "env",
		},
		{
			name:         "dependencies list start flag required",
			cmd:          apmDependenciesListCmd,
			requiredFlag: "start",
		},
		{
			name:         "dependencies list end flag required",
			cmd:          apmDependenciesListCmd,
			requiredFlag: "end",
		},
		{
			name:         "flow-map query flag required",
			cmd:          apmFlowMapCmd,
			requiredFlag: "query",
		},
		{
			name:         "flow-map from flag required",
			cmd:          apmFlowMapCmd,
			requiredFlag: "from",
		},
		{
			name:         "flow-map to flag required",
			cmd:          apmFlowMapCmd,
			requiredFlag: "to",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := tt.cmd.Flags().Lookup(tt.requiredFlag)
			if flag == nil {
				t.Fatalf("flag %s not found", tt.requiredFlag)
			}
			annotations := flag.Annotations
			if annotations == nil {
				t.Fatalf("flag %s has no annotations", tt.requiredFlag)
			}
			required, ok := annotations["cobra_annotation_bash_completion_one_required_flag"]
			if !ok || len(required) == 0 {
				t.Errorf("flag %s is not marked as required", tt.requiredFlag)
			}
		})
	}
}
