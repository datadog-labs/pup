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

func TestDashboardsCmd(t *testing.T) {
	if dashboardsCmd == nil {
		t.Fatal("dashboardsCmd is nil")
	}

	if dashboardsCmd.Use != "dashboards" {
		t.Errorf("Use = %s, want dashboards", dashboardsCmd.Use)
	}

	if dashboardsCmd.Short == "" {
		t.Error("Short description is empty")
	}
}

func TestDashboardsCmd_Subcommands(t *testing.T) {
	expectedCommands := []string{"list", "get", "delete"}

	commands := dashboardsCmd.Commands()

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

// Helper function to setup dashboards test client
func setupDashboardsTestClient(t *testing.T) func() {
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

func TestRunDashboardsList(t *testing.T) {
	cleanup := setupDashboardsTestClient(t)
	defer cleanup()

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "fails on client creation",
			wantErr: true, // Mock client error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			err := runDashboardsList(dashboardsListCmd, []string{})

			if (err != nil) != tt.wantErr {
				t.Errorf("runDashboardsList() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunDashboardsGet(t *testing.T) {
	cleanup := setupDashboardsTestClient(t)
	defer cleanup()

	tests := []struct {
		name        string
		dashboardID string
		wantErr     bool
	}{
		{
			name:        "with valid dashboard ID",
			dashboardID: "abc-123-xyz",
			wantErr:     true, // Mock client error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			err := runDashboardsGet(dashboardsGetCmd, []string{tt.dashboardID})

			if (err != nil) != tt.wantErr {
				t.Errorf("runDashboardsGet() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunDashboardsDelete_AutoApprove(t *testing.T) {
	cleanup := setupDashboardsTestClient(t)
	defer cleanup()

	// Set auto-approve
	cfg.AutoApprove = true

	tests := []struct {
		name        string
		dashboardID string
		wantErr     bool
	}{
		{
			name:        "with auto-approve",
			dashboardID: "abc-123-xyz",
			wantErr:     true, // Mock client error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			err := runDashboardsDelete(dashboardsDeleteCmd, []string{tt.dashboardID})

			if (err != nil) != tt.wantErr {
				t.Errorf("runDashboardsDelete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunDashboardsDelete_WithConfirmation(t *testing.T) {
	cleanup := setupDashboardsTestClient(t)
	defer cleanup()

	// Disable auto-approve
	cfg.AutoApprove = false

	tests := []struct {
		name        string
		dashboardID string
		input       string
		wantErr     bool
	}{
		{
			name:        "fails on client creation (mock)",
			dashboardID: "abc-123-xyz",
			input:       "n\n",
			wantErr:     true, // getClient() called before confirmation
		},
		{
			name:        "fails on client creation with yes (mock)",
			dashboardID: "abc-123-xyz",
			input:       "yes\n",
			wantErr:     true, // getClient() called before confirmation
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

			err := runDashboardsDelete(dashboardsDeleteCmd, []string{tt.dashboardID})

			if (err != nil) != tt.wantErr {
				t.Errorf("runDashboardsDelete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDashboardsCmd_ParentChild(t *testing.T) {
	commands := dashboardsCmd.Commands()

	for _, cmd := range commands {
		if cmd.Parent() != dashboardsCmd {
			t.Errorf("Command %s parent is not dashboardsCmd", cmd.Use)
		}
	}
}
