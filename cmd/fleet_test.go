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

func TestFleetCmd(t *testing.T) {
	if fleetCmd == nil {
		t.Fatal("fleetCmd is nil")
	}

	if fleetCmd.Use != "fleet" {
		t.Errorf("Use = %s, want fleet", fleetCmd.Use)
	}

	if fleetCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if fleetCmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestFleetCmd_Subcommands(t *testing.T) {
	expectedCommands := []string{"agents", "deployments", "schedules"}

	commands := fleetCmd.Commands()
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

func TestFleetAgentsCmd_Subcommands(t *testing.T) {
	expectedCommands := []string{"list", "get", "versions"}

	commands := fleetAgentsCmd.Commands()
	commandMap := make(map[string]bool)
	for _, cmd := range commands {
		commandMap[cmd.Name()] = true
	}

	for _, expected := range expectedCommands {
		if !commandMap[expected] {
			t.Errorf("Missing agents subcommand: %s", expected)
		}
	}
}

func TestFleetDeploymentsCmd_Subcommands(t *testing.T) {
	expectedCommands := []string{"list", "get", "configure", "upgrade", "cancel"}

	commands := fleetDeploymentsCmd.Commands()
	commandMap := make(map[string]bool)
	for _, cmd := range commands {
		commandMap[cmd.Name()] = true
	}

	for _, expected := range expectedCommands {
		if !commandMap[expected] {
			t.Errorf("Missing deployments subcommand: %s", expected)
		}
	}
}

func TestFleetSchedulesCmd_Subcommands(t *testing.T) {
	expectedCommands := []string{"list", "get", "create", "update", "delete", "trigger"}

	commands := fleetSchedulesCmd.Commands()
	commandMap := make(map[string]bool)
	for _, cmd := range commands {
		commandMap[cmd.Name()] = true
	}

	for _, expected := range expectedCommands {
		if !commandMap[expected] {
			t.Errorf("Missing schedules subcommand: %s", expected)
		}
	}
}

func TestFleetAgentsListCmd_Flags(t *testing.T) {
	flags := fleetAgentsListCmd.Flags()

	expectedFlags := []string{"page-size", "tags", "filter", "sort", "desc"}
	for _, flag := range expectedFlags {
		if flags.Lookup(flag) == nil {
			t.Errorf("Missing --%s flag on agents list", flag)
		}
	}
}

func TestFleetDeploymentsListCmd_Flags(t *testing.T) {
	flags := fleetDeploymentsListCmd.Flags()

	if flags.Lookup("page-size") == nil {
		t.Error("Missing --page-size flag on deployments list")
	}
}

func TestFleetDeploymentsConfigureCmd_Flags(t *testing.T) {
	flags := fleetDeploymentsConfigureCmd.Flags()

	if flags.Lookup("file") == nil {
		t.Error("Missing --file flag on deployments configure")
	}
}

func TestFleetDeploymentsUpgradeCmd_Flags(t *testing.T) {
	flags := fleetDeploymentsUpgradeCmd.Flags()

	if flags.Lookup("file") == nil {
		t.Error("Missing --file flag on deployments upgrade")
	}
}

func TestFleetSchedulesCreateCmd_Flags(t *testing.T) {
	flags := fleetSchedulesCreateCmd.Flags()

	if flags.Lookup("file") == nil {
		t.Error("Missing --file flag on schedules create")
	}
}

func TestFleetSchedulesUpdateCmd_Flags(t *testing.T) {
	flags := fleetSchedulesUpdateCmd.Flags()

	if flags.Lookup("file") == nil {
		t.Error("Missing --file flag on schedules update")
	}
}

func TestFleetAgentsGetCmd_Args(t *testing.T) {
	if fleetAgentsGetCmd.Args == nil {
		t.Error("Args validator is nil")
	}
}

func TestFleetDeploymentsGetCmd_Args(t *testing.T) {
	if fleetDeploymentsGetCmd.Args == nil {
		t.Error("Args validator is nil")
	}
}

func TestFleetDeploymentsCancelCmd_Args(t *testing.T) {
	if fleetDeploymentsCancelCmd.Args == nil {
		t.Error("Args validator is nil")
	}
}

func TestFleetSchedulesGetCmd_Args(t *testing.T) {
	if fleetSchedulesGetCmd.Args == nil {
		t.Error("Args validator is nil")
	}
}

func TestFleetSchedulesUpdateCmd_Args(t *testing.T) {
	if fleetSchedulesUpdateCmd.Args == nil {
		t.Error("Args validator is nil")
	}
}

func TestFleetSchedulesDeleteCmd_Args(t *testing.T) {
	if fleetSchedulesDeleteCmd.Args == nil {
		t.Error("Args validator is nil")
	}
}

func TestFleetSchedulesTriggerCmd_Args(t *testing.T) {
	if fleetSchedulesTriggerCmd.Args == nil {
		t.Error("Args validator is nil")
	}
}

func setupFleetTestClient(t *testing.T) func() {
	t.Helper()

	origClient := ddClient
	origCfg := cfg
	origFactory := clientFactory
	origAPIKeyFactory := apiKeyClientFactory

	cfg = &config.Config{
		Site:        "datadoghq.com",
		APIKey:      "test-api-key-12345678",
		AppKey:      "test-app-key-12345678",
		AutoApprove: false,
	}

	mockErr := func(c *config.Config) (*client.Client, error) {
		return nil, fmt.Errorf("mock client: no real API connection in tests")
	}
	clientFactory = mockErr
	apiKeyClientFactory = mockErr

	ddClient = nil

	return func() {
		ddClient = origClient
		cfg = origCfg
		clientFactory = origFactory
		apiKeyClientFactory = origAPIKeyFactory
	}
}

func TestRunFleetAgentsList(t *testing.T) {
	cleanup := setupFleetTestClient(t)
	defer cleanup()

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "fails on client creation",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			err := runFleetAgentsList(fleetAgentsListCmd, []string{})

			if (err != nil) != tt.wantErr {
				t.Errorf("runFleetAgentsList() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunFleetAgentsGet(t *testing.T) {
	cleanup := setupFleetTestClient(t)
	defer cleanup()

	var buf bytes.Buffer
	outputWriter = &buf
	defer func() { outputWriter = os.Stdout }()

	err := runFleetAgentsGet(fleetAgentsGetCmd, []string{"agent-key-123"})
	if err == nil {
		t.Error("expected error due to mock client, got nil")
	}
}

func TestRunFleetAgentsVersions(t *testing.T) {
	cleanup := setupFleetTestClient(t)
	defer cleanup()

	var buf bytes.Buffer
	outputWriter = &buf
	defer func() { outputWriter = os.Stdout }()

	err := runFleetAgentsVersions(fleetAgentsVersionsCmd, []string{})
	if err == nil {
		t.Error("expected error due to mock client, got nil")
	}
}

func TestRunFleetDeploymentsList(t *testing.T) {
	cleanup := setupFleetTestClient(t)
	defer cleanup()

	var buf bytes.Buffer
	outputWriter = &buf
	defer func() { outputWriter = os.Stdout }()

	err := runFleetDeploymentsList(fleetDeploymentsListCmd, []string{})
	if err == nil {
		t.Error("expected error due to mock client, got nil")
	}
}

func TestRunFleetDeploymentsGet(t *testing.T) {
	cleanup := setupFleetTestClient(t)
	defer cleanup()

	var buf bytes.Buffer
	outputWriter = &buf
	defer func() { outputWriter = os.Stdout }()

	err := runFleetDeploymentsGet(fleetDeploymentsGetCmd, []string{"deploy-123"})
	if err == nil {
		t.Error("expected error due to mock client, got nil")
	}
}

func TestRunFleetDeploymentsCancel_AutoApprove(t *testing.T) {
	cleanup := setupFleetTestClient(t)
	defer cleanup()

	cfg.AutoApprove = true

	var buf bytes.Buffer
	outputWriter = &buf
	defer func() { outputWriter = os.Stdout }()

	err := runFleetDeploymentsCancel(fleetDeploymentsCancelCmd, []string{"deploy-123"})
	if err == nil {
		t.Error("expected error due to mock client, got nil")
	}
}

func TestRunFleetDeploymentsCancel_WithConfirmation(t *testing.T) {
	cleanup := setupFleetTestClient(t)
	defer cleanup()

	cfg.AutoApprove = false

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "user declines",
			input:   "n\n",
			wantErr: false,
		},
		{
			name:    "user confirms - fails on client creation",
			input:   "yes\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			inputReader = strings.NewReader(tt.input)
			defer func() { inputReader = os.Stdin }()

			err := runFleetDeploymentsCancel(fleetDeploymentsCancelCmd, []string{"deploy-123"})

			if (err != nil) != tt.wantErr {
				t.Errorf("runFleetDeploymentsCancel() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunFleetSchedulesList(t *testing.T) {
	cleanup := setupFleetTestClient(t)
	defer cleanup()

	var buf bytes.Buffer
	outputWriter = &buf
	defer func() { outputWriter = os.Stdout }()

	err := runFleetSchedulesList(fleetSchedulesListCmd, []string{})
	if err == nil {
		t.Error("expected error due to mock client, got nil")
	}
}

func TestRunFleetSchedulesGet(t *testing.T) {
	cleanup := setupFleetTestClient(t)
	defer cleanup()

	var buf bytes.Buffer
	outputWriter = &buf
	defer func() { outputWriter = os.Stdout }()

	err := runFleetSchedulesGet(fleetSchedulesGetCmd, []string{"sched-123"})
	if err == nil {
		t.Error("expected error due to mock client, got nil")
	}
}

func TestRunFleetSchedulesDelete_AutoApprove(t *testing.T) {
	cleanup := setupFleetTestClient(t)
	defer cleanup()

	cfg.AutoApprove = true

	var buf bytes.Buffer
	outputWriter = &buf
	defer func() { outputWriter = os.Stdout }()

	err := runFleetSchedulesDelete(fleetSchedulesDeleteCmd, []string{"sched-123"})
	if err == nil {
		t.Error("expected error due to mock client, got nil")
	}
}

func TestRunFleetSchedulesDelete_WithConfirmation(t *testing.T) {
	cleanup := setupFleetTestClient(t)
	defer cleanup()

	cfg.AutoApprove = false

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "user declines",
			input:   "n\n",
			wantErr: false,
		},
		{
			name:    "user confirms - fails on client creation",
			input:   "yes\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			inputReader = strings.NewReader(tt.input)
			defer func() { inputReader = os.Stdin }()

			err := runFleetSchedulesDelete(fleetSchedulesDeleteCmd, []string{"sched-123"})

			if (err != nil) != tt.wantErr {
				t.Errorf("runFleetSchedulesDelete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunFleetSchedulesTrigger(t *testing.T) {
	cleanup := setupFleetTestClient(t)
	defer cleanup()

	var buf bytes.Buffer
	outputWriter = &buf
	defer func() { outputWriter = os.Stdout }()

	err := runFleetSchedulesTrigger(fleetSchedulesTriggerCmd, []string{"sched-123"})
	if err == nil {
		t.Error("expected error due to mock client, got nil")
	}
}

func TestFleetCmd_ParentChild(t *testing.T) {
	commands := fleetCmd.Commands()

	for _, cmd := range commands {
		if cmd.Parent() != fleetCmd {
			t.Errorf("Command %s parent is not fleetCmd", cmd.Use)
		}
	}
}
