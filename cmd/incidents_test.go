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

func TestIncidentsCmd(t *testing.T) {
	if incidentsCmd == nil {
		t.Fatal("incidentsCmd is nil")
	}
}

func setupIncidentsTestClient(t *testing.T) func() {
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

func TestRunIncidentsList(t *testing.T) {
	cleanup := setupIncidentsTestClient(t)
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

			err := runIncidentsList(incidentsListCmd, []string{})

			if (err != nil) != tt.wantErr {
				t.Errorf("runIncidentsList() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIncidentsCmd_NewSubcommands(t *testing.T) {
	expectedCommands := []string{"settings", "handles", "postmortem-templates"}
	commands := incidentsCmd.Commands()
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

func TestRunIncidentsSettingsGet(t *testing.T) {
	cleanup := setupIncidentsTestClient(t)
	defer cleanup()

	var buf bytes.Buffer
	outputWriter = &buf
	defer func() { outputWriter = os.Stdout }()

	err := runIncidentsSettingsGet(incidentsSettingsGetCmd, []string{})
	if err == nil {
		t.Error("expected error due to mock client, got nil")
	}
}

func TestRunIncidentsPostmortemTemplatesList(t *testing.T) {
	cleanup := setupIncidentsTestClient(t)
	defer cleanup()

	var buf bytes.Buffer
	outputWriter = &buf
	defer func() { outputWriter = os.Stdout }()

	err := runIncidentsPostmortemTemplatesList(incidentsPostmortemTemplatesListCmd, []string{})
	if err == nil {
		t.Error("expected error due to mock client, got nil")
	}
}

func TestRunIncidentsGet(t *testing.T) {
	cleanup := setupIncidentsTestClient(t)
	defer cleanup()

	tests := []struct {
		name       string
		incidentID string
		wantErr    bool
	}{
		{
			name:       "with valid incident ID",
			incidentID: "incident-123",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			err := runIncidentsGet(incidentsGetCmd, []string{tt.incidentID})

			if (err != nil) != tt.wantErr {
				t.Errorf("runIncidentsGet() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
