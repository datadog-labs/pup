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

	"github.com/DataDog/pup/pkg/client"
	"github.com/DataDog/pup/pkg/config"
)

func TestMonitorsCmd(t *testing.T) {
	if monitorsCmd == nil {
		t.Fatal("monitorsCmd is nil")
	}

	if monitorsCmd.Use != "monitors" {
		t.Errorf("Use = %s, want monitors", monitorsCmd.Use)
	}

	if monitorsCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if monitorsCmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestMonitorsCmd_Subcommands(t *testing.T) {
	expectedCommands := []string{"list", "get", "delete"}

	commands := monitorsCmd.Commands()

	commandMap := make(map[string]bool)
	for _, cmd := range commands {
		commandMap[cmd.Name()] = true
	}

	for _, expected := range expectedCommands {
		if !commandMap[expected] {
			t.Errorf("Missing subcommand: %s", expected)
		}
	}
}

func TestMonitorsListCmd(t *testing.T) {
	if monitorsListCmd == nil {
		t.Fatal("monitorsListCmd is nil")
	}

	if monitorsListCmd.Use != "list" {
		t.Errorf("Use = %s, want list", monitorsListCmd.Use)
	}

	if monitorsListCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if monitorsListCmd.RunE == nil {
		t.Error("RunE is nil")
	}

	// Check for flags
	flags := monitorsListCmd.Flags()
	if flags.Lookup("name") == nil {
		t.Error("Missing --name flag")
	}
	if flags.Lookup("tags") == nil {
		t.Error("Missing --tags flag")
	}
}

func TestMonitorsGetCmd(t *testing.T) {
	if monitorsGetCmd == nil {
		t.Fatal("monitorsGetCmd is nil")
	}

	if monitorsGetCmd.Use != "get [monitor-id]" {
		t.Errorf("Use = %s, want 'get [monitor-id]'", monitorsGetCmd.Use)
	}

	if monitorsGetCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if monitorsGetCmd.RunE == nil {
		t.Error("RunE is nil")
	}

	if monitorsGetCmd.Args == nil {
		t.Error("Args validator is nil")
	}
}

func TestMonitorsDeleteCmd(t *testing.T) {
	if monitorsDeleteCmd == nil {
		t.Fatal("monitorsDeleteCmd is nil")
	}

	if monitorsDeleteCmd.Use != "delete [monitor-id]" {
		t.Errorf("Use = %s, want 'delete [monitor-id]'", monitorsDeleteCmd.Use)
	}

	if monitorsDeleteCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if monitorsDeleteCmd.RunE == nil {
		t.Error("RunE is nil")
	}

	if monitorsDeleteCmd.Args == nil {
		t.Error("Args validator is nil")
	}
}

// Helper function to create a test client with mock data
func setupMonitorsTestClient(t *testing.T) func() {
	t.Helper()

	// Save original values
	origClient := ddClient
	origCfg := cfg
	origFactory := clientFactory

	// Create test config
	cfg = &config.Config{
		Site:        "datadoghq.com",
		APIKey:      "test-api-key-12345678",
		AppKey:      "test-app-key-12345678",
		AutoApprove: false,
	}

	// Mock the client factory to return an error immediately
	// This prevents keychain access attempts
	clientFactory = func(c *config.Config) (*client.Client, error) {
		return nil, fmt.Errorf("mock client: no real API connection in tests")
	}

	ddClient = nil

	// Return cleanup function
	return func() {
		ddClient = origClient
		cfg = origCfg
		clientFactory = origFactory
	}
}

func TestRunMonitorsList(t *testing.T) {
	cleanup := setupMonitorsTestClient(t)
	defer cleanup()

	tests := []struct {
		name        string
		nameFilter  string
		tagsFilter  string
		wantErr     bool
		wantErrType string
	}{
		{
			name:        "no filters",
			nameFilter:  "",
			tagsFilter:  "",
			wantErr:     true, // Will fail without real API
			wantErrType: "mock client",
		},
		{
			name:        "with name filter",
			nameFilter:  "CPU",
			tagsFilter:  "",
			wantErr:     true,
			wantErrType: "mock client",
		},
		{
			name:        "with tags filter",
			nameFilter:  "",
			tagsFilter:  "env:prod",
			wantErr:     true,
			wantErrType: "mock client",
		},
		{
			name:        "with both filters",
			nameFilter:  "CPU",
			tagsFilter:  "env:prod",
			wantErr:     true,
			wantErrType: "mock client",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set filter flags
			monitorName = tt.nameFilter
			monitorTags = tt.tagsFilter

			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			err := runMonitorsList(monitorsListCmd, []string{})

			if (err != nil) != tt.wantErr {
				t.Errorf("runMonitorsList() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && tt.wantErrType != "" && err != nil {
				if !strings.Contains(err.Error(), tt.wantErrType) {
					t.Errorf("runMonitorsList() error = %v, want error containing %v", err, tt.wantErrType)
				}
			}
		})
	}
}

func TestRunMonitorsGet(t *testing.T) {
	cleanup := setupMonitorsTestClient(t)
	defer cleanup()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "with valid monitor ID",
			args:    []string{"12345"},
			wantErr: true, // Will fail without real API
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			err := runMonitorsGet(monitorsGetCmd, tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("runMonitorsGet() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunMonitorsDelete_AutoApprove(t *testing.T) {
	cleanup := setupMonitorsTestClient(t)
	defer cleanup()

	// Set auto-approve
	cfg.AutoApprove = true

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "with auto-approve",
			args:    []string{"12345"},
			wantErr: true, // Will fail without real API
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			err := runMonitorsDelete(monitorsDeleteCmd, tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("runMonitorsDelete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunMonitorsDelete_WithConfirmation(t *testing.T) {
	cleanup := setupMonitorsTestClient(t)
	defer cleanup()

	// Disable auto-approve
	cfg.AutoApprove = false

	tests := []struct {
		name    string
		args    []string
		input   string
		wantErr bool
	}{
		{
			name:    "fails on client creation (mock)",
			args:    []string{"12345"},
			input:   "n\n",
			wantErr: true, // getClient() called before confirmation
		},
		{
			name:    "fails on client creation with yes (mock)",
			args:    []string{"12345"},
			input:   "yes\n",
			wantErr: true, // getClient() called before confirmation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			// Simulate input
			inputReader = strings.NewReader(tt.input)
			defer func() { inputReader = os.Stdin }()

			err := runMonitorsDelete(monitorsDeleteCmd, tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("runMonitorsDelete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMonitorsCmd_ParentChild(t *testing.T) {
	commands := monitorsCmd.Commands()

	for _, cmd := range commands {
		if cmd.Parent() != monitorsCmd {
			t.Errorf("Command %s parent is not monitorsCmd", cmd.Use)
		}
	}
}
