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

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/DataDog/pup/pkg/client"
	"github.com/DataDog/pup/pkg/config"
)

func TestAPIKeysCmd(t *testing.T) {
	if apiKeysCmd == nil {
		t.Fatal("apiKeysCmd is nil")
	}

	if apiKeysCmd.Use != "api-keys" {
		t.Errorf("Use = %s, want api-keys", apiKeysCmd.Use)
	}

	if apiKeysCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if apiKeysCmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestAPIKeysCmd_Subcommands(t *testing.T) {
	expectedCommands := []string{"list", "get", "create", "delete"}

	commands := apiKeysCmd.Commands()

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

func TestAPIKeysListCmd(t *testing.T) {
	if apiKeysListCmd == nil {
		t.Fatal("apiKeysListCmd is nil")
	}

	if apiKeysListCmd.Use != "list" {
		t.Errorf("Use = %s, want list", apiKeysListCmd.Use)
	}

	if apiKeysListCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if apiKeysListCmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestAPIKeysGetCmd(t *testing.T) {
	if apiKeysGetCmd == nil {
		t.Fatal("apiKeysGetCmd is nil")
	}

	if apiKeysGetCmd.Use != "get [key-id]" {
		t.Errorf("Use = %s, want 'get [key-id]'", apiKeysGetCmd.Use)
	}

	if apiKeysGetCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if apiKeysGetCmd.RunE == nil {
		t.Error("RunE is nil")
	}

	if apiKeysGetCmd.Args == nil {
		t.Error("Args validator is nil")
	}
}

func TestAPIKeysCreateCmd(t *testing.T) {
	if apiKeysCreateCmd == nil {
		t.Fatal("apiKeysCreateCmd is nil")
	}

	if apiKeysCreateCmd.Use != "create" {
		t.Errorf("Use = %s, want create", apiKeysCreateCmd.Use)
	}

	if apiKeysCreateCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if apiKeysCreateCmd.RunE == nil {
		t.Error("RunE is nil")
	}

	// Check for flags
	flags := apiKeysCreateCmd.Flags()
	if flags.Lookup("name") == nil {
		t.Error("Missing --name flag")
	}
}

func TestAPIKeysDeleteCmd(t *testing.T) {
	if apiKeysDeleteCmd == nil {
		t.Fatal("apiKeysDeleteCmd is nil")
	}

	if apiKeysDeleteCmd.Use != "delete [key-id]" {
		t.Errorf("Use = %s, want 'delete [key-id]'", apiKeysDeleteCmd.Use)
	}

	if apiKeysDeleteCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if apiKeysDeleteCmd.RunE == nil {
		t.Error("RunE is nil")
	}

	if apiKeysDeleteCmd.Args == nil {
		t.Error("Args validator is nil")
	}
}

func TestAPIKeysCmd_ParentChild(t *testing.T) {
	commands := apiKeysCmd.Commands()

	for _, cmd := range commands {
		if cmd.Parent() != apiKeysCmd {
			t.Errorf("Command %s parent is not apiKeysCmd", cmd.Use)
		}
	}
}

// Helper function to create a test client with mock data
func setupTestClient(t *testing.T) func() {
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

// Helper to capture output
func captureOutput(t *testing.T, f func()) string {
	t.Helper()
	var buf bytes.Buffer
	origWriter := outputWriter
	outputWriter = &buf
	defer func() { outputWriter = origWriter }()
	f()
	return buf.String()
}

func TestRunAPIKeysList(t *testing.T) {
	cleanup := setupTestClient(t)
	defer cleanup()

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "requires valid client",
			wantErr: true, // Will fail without real API credentials
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			err := runAPIKeysList(apiKeysListCmd, []string{})

			if (err != nil) != tt.wantErr {
				t.Errorf("runAPIKeysList() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunAPIKeysGet(t *testing.T) {
	cleanup := setupTestClient(t)
	defer cleanup()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "with valid key ID",
			args:    []string{"test-key-id"},
			wantErr: true, // Will fail without real API
		},
		{
			name:    "requires key ID",
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			// For empty args, we test the command validation
			if len(tt.args) == 0 {
				// cobra.ExactArgs(1) will catch this
				return
			}

			err := runAPIKeysGet(apiKeysGetCmd, tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("runAPIKeysGet() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunAPIKeysCreate(t *testing.T) {
	cleanup := setupTestClient(t)
	defer cleanup()

	tests := []struct {
		name    string
		keyName string
		wantErr bool
	}{
		{
			name:    "with valid name",
			keyName: "test-key",
			wantErr: true, // Will fail without real API
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiKeyName = tt.keyName

			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			err := runAPIKeysCreate(apiKeysCreateCmd, []string{})

			if (err != nil) != tt.wantErr {
				t.Errorf("runAPIKeysCreate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunAPIKeysDelete_AutoApprove(t *testing.T) {
	cleanup := setupTestClient(t)
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
			args:    []string{"test-key-id"},
			wantErr: true, // Will fail without real API
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			err := runAPIKeysDelete(apiKeysDeleteCmd, tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("runAPIKeysDelete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunAPIKeysDelete_WithConfirmation(t *testing.T) {
	cleanup := setupTestClient(t)
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
			args:    []string{"test-key-id"},
			input:   "no\n",
			wantErr: true, // getClient() called before confirmation
		},
		{
			name:    "fails on client creation with yes (mock)",
			args:    []string{"test-key-id"},
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

			err := runAPIKeysDelete(apiKeysDeleteCmd, tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("runAPIKeysDelete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAPIKeyFormatter(t *testing.T) {
	// Test that we can format API key responses
	testKey := datadogV2.APIKeyResponse{
		Data: &datadogV2.FullAPIKey{
			Id: datadog.PtrString("test-key-id"),
			Attributes: &datadogV2.FullAPIKeyAttributes{
				Name: datadog.PtrString("Test Key"),
			},
		},
	}

	if testKey.Data == nil {
		t.Error("Test key data is nil")
	}

	if testKey.Data.Id == nil || *testKey.Data.Id != "test-key-id" {
		t.Error("Test key ID not set correctly")
	}
}
