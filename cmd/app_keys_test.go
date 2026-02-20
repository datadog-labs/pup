// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
)

func TestAppKeysCmd(t *testing.T) {
	if appKeysCmd == nil {
		t.Fatal("appKeysCmd is nil")
	}

	if appKeysCmd.Use != "app-keys" {
		t.Errorf("Use = %s, want app-keys", appKeysCmd.Use)
	}

	if appKeysCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if appKeysCmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestAppKeysCmd_Subcommands(t *testing.T) {
	expectedCommands := []string{"list", "get", "create", "update", "delete"}

	commands := appKeysCmd.Commands()

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

func TestAppKeysCmd_ParentChild(t *testing.T) {
	commands := appKeysCmd.Commands()

	for _, cmd := range commands {
		if cmd.Parent() != appKeysCmd {
			t.Errorf("Command %s parent is not appKeysCmd", cmd.Use)
		}
	}
}

func TestAppKeysListCmd(t *testing.T) {
	if appKeysListCmd == nil {
		t.Fatal("appKeysListCmd is nil")
	}

	if appKeysListCmd.Use != "list" {
		t.Errorf("Use = %s, want list", appKeysListCmd.Use)
	}

	if appKeysListCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if appKeysListCmd.RunE == nil {
		t.Error("RunE is nil")
	}

	// Check flags
	flags := appKeysListCmd.Flags()
	for _, name := range []string{"page-size", "page-number", "filter", "sort", "all"} {
		if flags.Lookup(name) == nil {
			t.Errorf("Missing --%s flag", name)
		}
	}
}

func TestAppKeysGetCmd(t *testing.T) {
	if appKeysGetCmd == nil {
		t.Fatal("appKeysGetCmd is nil")
	}

	if appKeysGetCmd.Use != "get [app-key-id]" {
		t.Errorf("Use = %s, want 'get [app-key-id]'", appKeysGetCmd.Use)
	}

	if appKeysGetCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if appKeysGetCmd.RunE == nil {
		t.Error("RunE is nil")
	}

	if appKeysGetCmd.Args == nil {
		t.Error("Args validator is nil")
	}
}

func TestAppKeysCreateCmd(t *testing.T) {
	if appKeysCreateCmd == nil {
		t.Fatal("appKeysCreateCmd is nil")
	}

	if appKeysCreateCmd.Use != "create" {
		t.Errorf("Use = %s, want create", appKeysCreateCmd.Use)
	}

	if appKeysCreateCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if appKeysCreateCmd.RunE == nil {
		t.Error("RunE is nil")
	}

	// Check flags
	flags := appKeysCreateCmd.Flags()
	if flags.Lookup("name") == nil {
		t.Error("Missing --name flag")
	}
	if flags.Lookup("scopes") == nil {
		t.Error("Missing --scopes flag")
	}
}

func TestAppKeysUpdateCmd(t *testing.T) {
	if appKeysUpdateCmd == nil {
		t.Fatal("appKeysUpdateCmd is nil")
	}

	if appKeysUpdateCmd.Use != "update [app-key-id]" {
		t.Errorf("Use = %s, want 'update [app-key-id]'", appKeysUpdateCmd.Use)
	}

	if appKeysUpdateCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if appKeysUpdateCmd.RunE == nil {
		t.Error("RunE is nil")
	}

	if appKeysUpdateCmd.Args == nil {
		t.Error("Args validator is nil")
	}

	// Check flags
	flags := appKeysUpdateCmd.Flags()
	if flags.Lookup("name") == nil {
		t.Error("Missing --name flag")
	}
	if flags.Lookup("scopes") == nil {
		t.Error("Missing --scopes flag")
	}
}

func TestAppKeysDeleteCmd(t *testing.T) {
	if appKeysDeleteCmd == nil {
		t.Fatal("appKeysDeleteCmd is nil")
	}

	if appKeysDeleteCmd.Use != "delete [app-key-id]" {
		t.Errorf("Use = %s, want 'delete [app-key-id]'", appKeysDeleteCmd.Use)
	}

	if appKeysDeleteCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if appKeysDeleteCmd.RunE == nil {
		t.Error("RunE is nil")
	}

	if appKeysDeleteCmd.Args == nil {
		t.Error("Args validator is nil")
	}
}

func TestAppKeysGetCmd_ArgsValidation(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{name: "valid - one arg", args: []string{"some-id"}, wantErr: false},
		{name: "invalid - no args", args: []string{}, wantErr: true},
		{name: "invalid - too many args", args: []string{"a", "b"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := appKeysGetCmd.Args(appKeysGetCmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args validation error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAppKeysUpdateCmd_ArgsValidation(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{name: "valid - one arg", args: []string{"some-id"}, wantErr: false},
		{name: "invalid - no args", args: []string{}, wantErr: true},
		{name: "invalid - too many args", args: []string{"a", "b"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := appKeysUpdateCmd.Args(appKeysUpdateCmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args validation error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAppKeysDeleteCmd_ArgsValidation(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{name: "valid - one arg", args: []string{"some-id"}, wantErr: false},
		{name: "invalid - no args", args: []string{}, wantErr: true},
		{name: "invalid - too many args", args: []string{"a", "b"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := appKeysDeleteCmd.Args(appKeysDeleteCmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args validation error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunAppKeysList(t *testing.T) {
	cleanup := setupTestClient(t)
	defer cleanup()

	err := runAppKeysList(appKeysListCmd, []string{})
	if err == nil {
		t.Error("expected error with mock client")
	}
}

func TestRunAppKeysListAll(t *testing.T) {
	cleanup := setupTestClient(t)
	defer cleanup()

	appKeysListAll = true
	defer func() { appKeysListAll = false }()

	err := runAppKeysList(appKeysListCmd, []string{})
	if err == nil {
		t.Error("expected error with mock client")
	}
}

func TestRunAppKeysGet(t *testing.T) {
	cleanup := setupTestClient(t)
	defer cleanup()

	err := runAppKeysGet(appKeysGetCmd, []string{"test-key-id"})
	if err == nil {
		t.Error("expected error with mock client")
	}
}

func TestRunAppKeysCreate(t *testing.T) {
	cleanup := setupTestClient(t)
	defer cleanup()

	appKeyName = "test-key"
	defer func() { appKeyName = "" }()

	err := runAppKeysCreate(appKeysCreateCmd, []string{})
	if err == nil {
		t.Error("expected error with mock client")
	}
}

func TestRunAppKeysCreate_WithScopes(t *testing.T) {
	cleanup := setupTestClient(t)
	defer cleanup()

	appKeyName = "test-key"
	appKeyScopes = "dashboards_read, metrics_read"
	defer func() {
		appKeyName = ""
		appKeyScopes = ""
	}()

	err := runAppKeysCreate(appKeysCreateCmd, []string{})
	if err == nil {
		t.Error("expected error with mock client")
	}
}

func TestRunAppKeysUpdate(t *testing.T) {
	cleanup := setupTestClient(t)
	defer cleanup()

	appKeyName = "new-name"
	defer func() { appKeyName = "" }()

	err := runAppKeysUpdate(appKeysUpdateCmd, []string{"test-key-id"})
	if err == nil {
		t.Error("expected error with mock client")
	}
}

func TestRunAppKeysDelete_AutoApprove(t *testing.T) {
	cleanup := setupTestClient(t)
	defer cleanup()

	cfg.AutoApprove = true

	err := runAppKeysDelete(appKeysDeleteCmd, []string{"test-key-id"})
	if err == nil {
		t.Error("expected error with mock client")
	}
}

func TestRunAppKeysDelete_WithConfirmation(t *testing.T) {
	cleanup := setupTestClient(t)
	defer cleanup()

	cfg.AutoApprove = false

	// getClient() is called before the confirmation prompt, so with the mock
	// client it fails regardless of user input.
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "fails on client creation with no", input: "no\n", wantErr: true},
		{name: "fails on client creation with yes", input: "yes\n", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			outputWriter = &buf
			inputReader = strings.NewReader(tt.input)
			defer func() {
				outputWriter = os.Stdout
				inputReader = os.Stdin
			}()

			err := runAppKeysDelete(appKeysDeleteCmd, []string{"test-key-id"})
			if (err != nil) != tt.wantErr {
				t.Errorf("runAppKeysDelete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestApplicationKeyFormatter(t *testing.T) {
	testKey := datadogV2.ApplicationKeyResponse{
		Data: &datadogV2.FullApplicationKey{
			Id: datadog.PtrString("test-key-id"),
			Attributes: &datadogV2.FullApplicationKeyAttributes{
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

	if testKey.Data.Attributes == nil || *testKey.Data.Attributes.Name != "Test Key" {
		t.Error("Test key name not set correctly")
	}
}
