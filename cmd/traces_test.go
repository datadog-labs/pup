// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"testing"
)

func TestTracesCmd(t *testing.T) {
	if tracesCmd == nil {
		t.Fatal("tracesCmd is nil")
	}
	if tracesCmd.Use != "traces" {
		t.Errorf("Use = %s, want traces", tracesCmd.Use)
	}
}

func TestTracesSubcommands(t *testing.T) {
	subcommands := map[string]bool{
		"search":    false,
		"aggregate": false,
	}
	for _, cmd := range tracesCmd.Commands() {
		if _, ok := subcommands[cmd.Name()]; ok {
			subcommands[cmd.Name()] = true
		}
	}
	for name, found := range subcommands {
		if !found {
			t.Errorf("missing subcommand: %s", name)
		}
	}
}

func TestTracesSearchFlags(t *testing.T) {
	flags := tracesSearchCmd.Flags()

	tests := []struct {
		name         string
		defaultValue string
	}{
		{"query", "*"},
		{"from", "1h"},
		{"to", "now"},
		{"sort", "-timestamp"},
	}

	for _, tt := range tests {
		f := flags.Lookup(tt.name)
		if f == nil {
			t.Errorf("search command missing --%s flag", tt.name)
			continue
		}
		if f.DefValue != tt.defaultValue {
			t.Errorf("--%s default = %q, want %q", tt.name, f.DefValue, tt.defaultValue)
		}
	}

	// Check limit flag separately (int type)
	limitFlag := flags.Lookup("limit")
	if limitFlag == nil {
		t.Error("search command missing --limit flag")
	} else if limitFlag.DefValue != "50" {
		t.Errorf("--limit default = %q, want %q", limitFlag.DefValue, "50")
	}
}

func TestTracesAggregateFlags(t *testing.T) {
	flags := tracesAggregateCmd.Flags()

	tests := []struct {
		name         string
		defaultValue string
	}{
		{"query", "*"},
		{"from", "1h"},
		{"to", "now"},
		{"compute", ""},
		{"group-by", ""},
	}

	for _, tt := range tests {
		f := flags.Lookup(tt.name)
		if f == nil {
			t.Errorf("aggregate command missing --%s flag", tt.name)
			continue
		}
		if f.DefValue != tt.defaultValue {
			t.Errorf("--%s default = %q, want %q", tt.name, f.DefValue, tt.defaultValue)
		}
	}
}
