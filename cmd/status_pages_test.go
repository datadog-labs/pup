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

func TestStatusPagesCmd(t *testing.T) {
	if statusPagesCmd == nil {
		t.Fatal("statusPagesCmd is nil")
	}
	if statusPagesCmd.Use != "status-pages" {
		t.Errorf("Use = %s, want status-pages", statusPagesCmd.Use)
	}
	if statusPagesCmd.Short == "" {
		t.Error("Short description is empty")
	}
}

func TestStatusPagesCmd_Subcommands(t *testing.T) {
	expectedCommands := []string{"pages", "components", "degradations"}
	commands := statusPagesCmd.Commands()
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

func TestStatusPagesPagesCmd_Subcommands(t *testing.T) {
	expectedPrefixes := []string{"list", "get", "create", "update", "delete"}
	commands := statusPagesPagesCmd.Commands()
	for _, expected := range expectedPrefixes {
		found := false
		for _, cmd := range commands {
			if strings.HasPrefix(cmd.Use, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Missing pages subcommand starting with: %s", expected)
		}
	}
}

func setupStatusPagesTestClient(t *testing.T) func() {
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

func TestRunStatusPagesList(t *testing.T) {
	cleanup := setupStatusPagesTestClient(t)
	defer cleanup()

	var buf bytes.Buffer
	outputWriter = &buf
	defer func() { outputWriter = os.Stdout }()

	err := runStatusPagesList(statusPagesPagesListCmd, []string{})
	if err == nil {
		t.Error("expected error due to mock client, got nil")
	}
}

func TestRunStatusPagesGet(t *testing.T) {
	cleanup := setupStatusPagesTestClient(t)
	defer cleanup()

	var buf bytes.Buffer
	outputWriter = &buf
	defer func() { outputWriter = os.Stdout }()

	// Invalid UUID should fail
	err := runStatusPagesGet(statusPagesPagesGetCmd, []string{"not-a-uuid"})
	if err == nil {
		t.Error("expected error for invalid UUID, got nil")
	}
}

func TestParseUUID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "valid UUID", input: "123e4567-e89b-12d3-a456-426614174000", wantErr: false},
		{name: "invalid UUID", input: "not-a-uuid", wantErr: true},
		{name: "empty string", input: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseUUID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseUUID(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}
