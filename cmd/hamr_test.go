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

func TestHamrCmd(t *testing.T) {
	if hamrCmd == nil {
		t.Fatal("hamrCmd is nil")
	}
	if hamrCmd.Use != "hamr" {
		t.Errorf("Use = %s, want hamr", hamrCmd.Use)
	}
	if hamrCmd.Short == "" {
		t.Error("Short description is empty")
	}
}

func TestHamrCmd_Subcommands(t *testing.T) {
	expectedCommands := []string{"connections"}
	commands := hamrCmd.Commands()
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

func TestHamrConnectionsCmd_Subcommands(t *testing.T) {
	expectedCommands := []string{"get", "create"}
	commands := hamrConnectionsCmd.Commands()
	commandMap := make(map[string]bool)
	for _, cmd := range commands {
		commandMap[cmd.Use] = true
	}
	for _, expected := range expectedCommands {
		if !commandMap[expected] {
			t.Errorf("Missing connections subcommand: %s", expected)
		}
	}
}

func setupHamrTestClient(t *testing.T) func() {
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

func TestRunHamrConnectionsGet(t *testing.T) {
	cleanup := setupHamrTestClient(t)
	defer cleanup()

	var buf bytes.Buffer
	outputWriter = &buf
	defer func() { outputWriter = os.Stdout }()

	err := runHamrConnectionsGet(hamrConnectionsGetCmd, []string{})
	if err == nil {
		t.Error("expected error due to mock client, got nil")
	}
}
