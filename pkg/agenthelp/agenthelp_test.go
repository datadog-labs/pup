// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package agenthelp

import (
	"testing"

	"github.com/spf13/cobra"
)

func newTestRoot() *cobra.Command {
	root := &cobra.Command{Use: "pup", Short: "test CLI"}
	root.PersistentFlags().String("output", "json", "Output format")
	root.PersistentFlags().Bool("agent", false, "Agent mode")

	monitors := &cobra.Command{Use: "monitors", Short: "Manage monitors"}
	monitorsList := &cobra.Command{Use: "list", Short: "List monitors"}
	monitorsList.Flags().String("tags", "", "Filter by tags")
	monitorsList.Flags().Int("limit", 200, "Maximum results")
	monitorsGet := &cobra.Command{Use: "get", Short: "Get monitor details"}
	monitorsDelete := &cobra.Command{Use: "delete", Short: "Delete a monitor"}
	monitors.AddCommand(monitorsList, monitorsGet, monitorsDelete)

	logs := &cobra.Command{Use: "logs", Short: "Search and analyze logs"}
	logsSearch := &cobra.Command{Use: "search", Short: "Search logs"}
	logsSearch.Flags().String("query", "", "Search query")
	logsSearch.Flags().String("from", "1h", "Start time")
	logs.AddCommand(logsSearch)

	root.AddCommand(monitors, logs)
	return root
}

func TestGenerateSchema(t *testing.T) {
	root := newTestRoot()
	schema := GenerateSchema(root)

	if schema.Version == "" {
		t.Error("Schema.Version should not be empty")
	}
	if schema.Description == "" {
		t.Error("Schema.Description should not be empty")
	}
	if schema.Auth.OAuth == "" {
		t.Error("Schema.Auth.OAuth should not be empty")
	}
	if len(schema.GlobalFlags) == 0 {
		t.Error("Schema.GlobalFlags should not be empty")
	}
	if len(schema.Commands) == 0 {
		t.Error("Schema.Commands should not be empty")
	}
	if len(schema.QuerySyntax) == 0 {
		t.Error("Schema.QuerySyntax should not be empty")
	}
	if len(schema.TimeFormats.Relative) == 0 {
		t.Error("Schema.TimeFormats.Relative should not be empty")
	}
	if len(schema.Workflows) == 0 {
		t.Error("Schema.Workflows should not be empty")
	}
	if len(schema.BestPractices) == 0 {
		t.Error("Schema.BestPractices should not be empty")
	}
	if len(schema.AntiPatterns) == 0 {
		t.Error("Schema.AntiPatterns should not be empty")
	}
}

func TestGenerateSchema_CommandsIncludeSubcommands(t *testing.T) {
	root := newTestRoot()
	schema := GenerateSchema(root)

	var found bool
	for _, cmd := range schema.Commands {
		if cmd.Name == "monitors" {
			found = true
			if len(cmd.Subcommands) != 3 {
				t.Errorf("monitors should have 3 subcommands, got %d", len(cmd.Subcommands))
			}
			for _, sub := range cmd.Subcommands {
				if sub.Name == "list" {
					if len(sub.Flags) == 0 {
						t.Error("monitors list should have flags")
					}
				}
			}
		}
	}
	if !found {
		t.Error("monitors command not found in schema")
	}
}

func TestGenerateSchema_GlobalFlags(t *testing.T) {
	root := newTestRoot()
	schema := GenerateSchema(root)

	flagNames := make(map[string]bool)
	for _, f := range schema.GlobalFlags {
		flagNames[f.Name] = true
	}

	if !flagNames["--output"] {
		t.Error("Global flags should include --output")
	}
	if !flagNames["--agent"] {
		t.Error("Global flags should include --agent")
	}
}

func TestGenerateSubtreeSchema(t *testing.T) {
	root := newTestRoot()

	schema := GenerateSubtreeSchema(root, "monitors")
	if schema == nil {
		t.Fatal("GenerateSubtreeSchema should not return nil for 'monitors'")
	}
	if len(schema.Commands) != 1 {
		t.Errorf("Subtree schema should have 1 command, got %d", len(schema.Commands))
	}
	if schema.Commands[0].Name != "monitors" {
		t.Errorf("Subtree command should be 'monitors', got %q", schema.Commands[0].Name)
	}
}

func TestGenerateSubtreeSchema_NotFound(t *testing.T) {
	root := newTestRoot()

	schema := GenerateSubtreeSchema(root, "nonexistent")
	if schema != nil {
		t.Error("GenerateSubtreeSchema should return nil for nonexistent command")
	}
}

func TestGenerateCompactSchema(t *testing.T) {
	root := newTestRoot()
	compact := GenerateCompactSchema(root)

	if compact.Version == "" {
		t.Error("CompactSchema.Version should not be empty")
	}
	if len(compact.Commands) == 0 {
		t.Error("CompactSchema.Commands should not be empty")
	}

	for _, cmd := range compact.Commands {
		if cmd.Name == "monitors" {
			if len(cmd.Subcommands) != 3 {
				t.Errorf("monitors should have 3 subcommands, got %d", len(cmd.Subcommands))
			}
			for _, sub := range cmd.Subcommands {
				if sub.Name == "list" && len(sub.Flags) == 0 {
					t.Error("monitors list should have flags in compact schema")
				}
			}
		}
	}
}

func TestIsReadOnlyCommand(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"list", true},
		{"get", true},
		{"search", true},
		{"query", true},
		{"delete", false},
		{"create", false},
		{"update", false},
		{"set", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isReadOnlyCommand(tt.name)
			if got != tt.want {
				t.Errorf("isReadOnlyCommand(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestFilterQuerySyntax(t *testing.T) {
	// Known domain
	result := filterQuerySyntax("logs")
	if _, ok := result["logs"]; !ok {
		t.Error("filterQuerySyntax('logs') should contain 'logs' key")
	}
	if len(result) != 1 {
		t.Errorf("filterQuerySyntax('logs') should have 1 entry, got %d", len(result))
	}

	// Unknown domain returns all
	result = filterQuerySyntax("unknown")
	if len(result) < 2 {
		t.Error("filterQuerySyntax('unknown') should return all entries")
	}
}
