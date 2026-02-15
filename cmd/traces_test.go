// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/DataDog/pup/pkg/client"
	"github.com/DataDog/pup/pkg/config"
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

func setupTracesTestClient(t *testing.T) func() {
	t.Helper()

	origClient := ddClient
	origCfg := cfg
	origFactory := clientFactory

	cfg = &config.Config{
		Site:   "datadoghq.com",
		APIKey: "test-api-key-12345678",
		AppKey: "test-app-key-12345678",
	}

	clientFactory = func(c *config.Config) (*client.Client, error) {
		return nil, fmt.Errorf("mock client: no real API connection in tests")
	}

	ddClient = nil

	return func() {
		ddClient = origClient
		cfg = origCfg
		clientFactory = origFactory
	}
}

func TestRunTracesSearch(t *testing.T) {
	cleanup := setupTracesTestClient(t)
	defer cleanup()

	tests := []struct {
		name    string
		query   string
		from    string
		to      string
		limit   int
		sort    string
		wantErr bool
	}{
		{
			name:    "valid query returns error from mock client",
			query:   "service:web-server",
			from:    "1h",
			to:      "now",
			limit:   50,
			sort:    "-timestamp",
			wantErr: true,
		},
		{
			name:    "invalid from time",
			query:   "*",
			from:    "invalid-time",
			to:      "now",
			limit:   50,
			sort:    "-timestamp",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracesQuery = tt.query
			tracesFrom = tt.from
			tracesTo = tt.to
			tracesLimit = tt.limit
			tracesSort = tt.sort

			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			err := runTracesSearch(tracesSearchCmd, []string{})

			if (err != nil) != tt.wantErr {
				t.Errorf("runTracesSearch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
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
