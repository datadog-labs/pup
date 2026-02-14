// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNotebooksCmd(t *testing.T) {
	if notebooksCmd == nil {
		t.Fatal("notebooksCmd is nil")
	}

	if notebooksCmd.Use != "notebooks" {
		t.Errorf("Use = %s, want notebooks", notebooksCmd.Use)
	}

	if notebooksCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if notebooksCmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestNotebooksCmd_Subcommands(t *testing.T) {
	expectedCommands := []string{"list", "get", "create", "update", "delete"}

	commands := notebooksCmd.Commands()

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

func TestNotebooksListCmd(t *testing.T) {
	if notebooksListCmd == nil {
		t.Fatal("notebooksListCmd is nil")
	}

	if notebooksListCmd.Use != "list" {
		t.Errorf("Use = %s, want list", notebooksListCmd.Use)
	}

	if notebooksListCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if notebooksListCmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestNotebooksGetCmd(t *testing.T) {
	if notebooksGetCmd == nil {
		t.Fatal("notebooksGetCmd is nil")
	}

	if notebooksGetCmd.Use != "get [notebook-id]" {
		t.Errorf("Use = %s, want 'get [notebook-id]'", notebooksGetCmd.Use)
	}

	if notebooksGetCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if notebooksGetCmd.RunE == nil {
		t.Error("RunE is nil")
	}

	if notebooksGetCmd.Args == nil {
		t.Error("Args validator is nil")
	}
}

func TestNotebooksDeleteCmd(t *testing.T) {
	if notebooksDeleteCmd == nil {
		t.Fatal("notebooksDeleteCmd is nil")
	}

	if notebooksDeleteCmd.Use != "delete [notebook-id]" {
		t.Errorf("Use = %s, want 'delete [notebook-id]'", notebooksDeleteCmd.Use)
	}

	if notebooksDeleteCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if notebooksDeleteCmd.RunE == nil {
		t.Error("RunE is nil")
	}

	if notebooksDeleteCmd.Args == nil {
		t.Error("Args validator is nil")
	}
}

func TestReadBody_File(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "body.json")
	content := []byte(`{"data":{"attributes":{"name":"test"}}}`)
	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		t.Fatal(err)
	}

	got, err := readBody("@" + tmpFile)
	if err != nil {
		t.Fatalf("readBody returned error: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("got %s, want %s", got, content)
	}
}

func TestReadBody_Stdin(t *testing.T) {
	content := `{"data":{"attributes":{"name":"test"}}}`
	origReader := inputReader
	inputReader = strings.NewReader(content)
	defer func() { inputReader = origReader }()

	got, err := readBody("-")
	if err != nil {
		t.Fatalf("readBody returned error: %v", err)
	}
	if string(got) != content {
		t.Errorf("got %s, want %s", got, content)
	}
}

func TestReadBody_MissingFile(t *testing.T) {
	_, err := readBody("@/nonexistent/path/body.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(err.Error(), "failed to read body file") {
		t.Errorf("error = %v, want 'failed to read body file'", err)
	}
}

func TestReadBody_InvalidJSON(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(tmpFile, []byte("not json"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := readBody("@" + tmpFile)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "invalid JSON in body") {
		t.Errorf("error = %v, want 'invalid JSON in body'", err)
	}
}

func TestReadBody_InvalidJSON_Stdin(t *testing.T) {
	origReader := inputReader
	inputReader = strings.NewReader("not json")
	defer func() { inputReader = origReader }()

	_, err := readBody("-")
	if err == nil {
		t.Fatal("expected error for invalid JSON from stdin")
	}
	if !strings.Contains(err.Error(), "invalid JSON in body") {
		t.Errorf("error = %v, want 'invalid JSON in body'", err)
	}
}

func TestReadBody_EmptyValue(t *testing.T) {
	_, err := readBody("")
	if err == nil {
		t.Fatal("expected error for empty body value")
	}
}

func TestNotebooksCreateCmd(t *testing.T) {
	if notebooksCreateCmd == nil {
		t.Fatal("notebooksCreateCmd is nil")
	}

	if notebooksCreateCmd.Use != "create" {
		t.Errorf("Use = %s, want create", notebooksCreateCmd.Use)
	}

	if notebooksCreateCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if notebooksCreateCmd.RunE == nil {
		t.Error("RunE is nil")
	}

	flags := notebooksCreateCmd.Flags()
	if flags.Lookup("body") == nil {
		t.Error("Missing --body flag")
	}
}

func TestNotebooksCreateCmd_BodyRequired(t *testing.T) {
	if notebooksCreateCmd.Flags().Lookup("body") == nil {
		t.Fatal("--body flag not found")
	}

	if err := notebooksCreateCmd.ValidateRequiredFlags(); err == nil {
		t.Error("expected --body to be required")
	}
}

func TestNotebooksUpdateCmd(t *testing.T) {
	if notebooksUpdateCmd == nil {
		t.Fatal("notebooksUpdateCmd is nil")
	}

	if notebooksUpdateCmd.Use != "update [notebook-id]" {
		t.Errorf("Use = %s, want 'update [notebook-id]'", notebooksUpdateCmd.Use)
	}

	if notebooksUpdateCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if notebooksUpdateCmd.RunE == nil {
		t.Error("RunE is nil")
	}

	if notebooksUpdateCmd.Args == nil {
		t.Error("Args validator is nil")
	}

	flags := notebooksUpdateCmd.Flags()
	if flags.Lookup("body") == nil {
		t.Error("Missing --body flag")
	}
}

func TestNotebooksCmd_ParentChild(t *testing.T) {
	commands := notebooksCmd.Commands()

	for _, cmd := range commands {
		if cmd.Parent() != notebooksCmd {
			t.Errorf("Command %s parent is not notebooksCmd", cmd.Use)
		}
	}
}
