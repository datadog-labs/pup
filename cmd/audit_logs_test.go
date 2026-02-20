// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/datadog-labs/pup/pkg/client"
	"github.com/datadog-labs/pup/pkg/config"
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

func setupAuditLogsTestClient(t *testing.T) func() {
	t.Helper()

	origClient := ddClient
	origCfg := cfg
	origFactory := clientFactory

	cfg = &config.Config{
		Site:        "datadoghq.com",
		APIKey:      "test-api-key-12345678",
		AppKey:      "test-app-key-12345678",
		AutoApprove: false,
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

func TestRunAuditLogsList(t *testing.T) {
	cleanup := setupAuditLogsTestClient(t)
	defer cleanup()

	tests := []struct {
		name      string
		from      string
		to        string
		wantErr   bool
		errSubstr string // if set, error must contain this substring
	}{
		{
			name:    "valid defaults reach API client",
			from:    "1h",
			to:      "now",
			wantErr: true,
		},
		{
			name:    "valid relative times",
			from:    "30m",
			to:      "now",
			wantErr: true,
		},
		{
			name:    "valid RFC3339 times",
			from:    "2024-01-01T00:00:00Z",
			to:      "2024-01-02T00:00:00Z",
			wantErr: true,
		},
		{
			name:      "invalid from time",
			from:      "notadate",
			to:        "now",
			wantErr:   true,
			errSubstr: "invalid --from time",
		},
		{
			name:      "invalid to time",
			from:      "1h",
			to:        "notadate",
			wantErr:   true,
			errSubstr: "invalid --to time",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auditLogsFrom = tt.from
			auditLogsTo = tt.to

			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			err := runAuditLogsList(auditLogsListCmd, []string{})

			if (err != nil) != tt.wantErr {
				t.Errorf("runAuditLogsList() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.errSubstr != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("runAuditLogsList() error = %q, want substring %q", err.Error(), tt.errSubstr)
				}
			}
		})
	}
}

func TestRunAuditLogsSearch(t *testing.T) {
	cleanup := setupAuditLogsTestClient(t)
	defer cleanup()

	tests := []struct {
		name      string
		query     string
		from      string
		to        string
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "valid defaults reach API client",
			query:   "*",
			from:    "1h",
			to:      "now",
			wantErr: true,
		},
		{
			name:    "valid query with relative times",
			query:   "@evt.outcome:error",
			from:    "30m",
			to:      "now",
			wantErr: true,
		},
		{
			name:    "valid query with RFC3339 times",
			query:   "*",
			from:    "2024-01-01T00:00:00Z",
			to:      "2024-01-02T00:00:00Z",
			wantErr: true,
		},
		{
			name:      "invalid from time",
			query:     "*",
			from:      "notadate",
			to:        "now",
			wantErr:   true,
			errSubstr: "invalid --from time",
		},
		{
			name:      "invalid to time",
			query:     "*",
			from:      "1h",
			to:        "notadate",
			wantErr:   true,
			errSubstr: "invalid --to time",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auditLogsQuery = tt.query
			auditLogsFrom = tt.from
			auditLogsTo = tt.to

			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			err := runAuditLogsSearch(auditLogsSearchCmd, []string{})

			if (err != nil) != tt.wantErr {
				t.Errorf("runAuditLogsSearch() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.errSubstr != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("runAuditLogsSearch() error = %q, want substring %q", err.Error(), tt.errSubstr)
				}
			}
		})
	}
}
