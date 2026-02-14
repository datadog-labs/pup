// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DataDog/pup/pkg/client"
	"github.com/DataDog/pup/pkg/config"
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
	expectedCommands := []string{"list", "get", "create", "update", "delete", "cells"}

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

func TestNotebooksUpdateCmd_BodyRequired(t *testing.T) {
	if notebooksUpdateCmd.Flags().Lookup("body") == nil {
		t.Fatal("--body flag not found")
	}

	if err := notebooksUpdateCmd.ValidateRequiredFlags(); err == nil {
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

func setupNotebooksTestClient(t *testing.T) func() {
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

	clientFactory = func(c *config.Config) (*client.Client, error) {
		return nil, fmt.Errorf("mock client: no real API connection in tests")
	}

	apiKeyClientFactory = func(c *config.Config) (*client.Client, error) {
		return nil, fmt.Errorf("mock api-key client: no real API connection in tests")
	}

	ddClient = nil

	return func() {
		ddClient = origClient
		cfg = origCfg
		clientFactory = origFactory
		apiKeyClientFactory = origAPIKeyFactory
	}
}

func TestNotebooksCellsCmd(t *testing.T) {
	if notebooksCellsCmd == nil {
		t.Fatal("notebooksCellsCmd is nil")
	}

	if notebooksCellsCmd.Use != "cells" {
		t.Errorf("Use = %s, want cells", notebooksCellsCmd.Use)
	}

	commands := notebooksCellsCmd.Commands()
	commandMap := make(map[string]bool)
	for _, cmd := range commands {
		commandMap[cmd.Name()] = true
	}
	if !commandMap["append"] {
		t.Error("Missing subcommand: append")
	}
}

func TestNotebooksCellsAppendCmd(t *testing.T) {
	if notebooksCellsAppendCmd == nil {
		t.Fatal("notebooksCellsAppendCmd is nil")
	}

	if notebooksCellsAppendCmd.Use != "append [notebook-id]" {
		t.Errorf("Use = %s, want 'append [notebook-id]'", notebooksCellsAppendCmd.Use)
	}

	if notebooksCellsAppendCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if notebooksCellsAppendCmd.RunE == nil {
		t.Error("RunE is nil")
	}

	if notebooksCellsAppendCmd.Args == nil {
		t.Error("Args validator is nil")
	}

	flags := notebooksCellsAppendCmd.Flags()
	if flags.Lookup("body") == nil {
		t.Error("Missing --body flag")
	}
}

func TestNotebooksCellsAppendCmd_BodyRequired(t *testing.T) {
	if err := notebooksCellsAppendCmd.ValidateRequiredFlags(); err == nil {
		t.Error("expected --body to be required")
	}
}

func TestRunNotebooksCellsAppend(t *testing.T) {
	cleanup := setupNotebooksTestClient(t)
	defer cleanup()

	tests := []struct {
		name        string
		args        []string
		body        string
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid markdown cell fails on client creation",
			args:        []string{"12345"},
			body:        `{"cells":[{"type":"markdown","data":"# Hello"}]}`,
			wantErr:     true,
			errContains: "no real API connection in tests",
		},
		{
			name:        "valid metric cell fails on client creation",
			args:        []string{"12345"},
			body:        `{"cells":[{"type":"metric","data":"avg:system.cpu.user{*}","title":"CPU"}]}`,
			wantErr:     true,
			errContains: "no real API connection in tests",
		},
		{
			name:        "valid logs cell fails on client creation",
			args:        []string{"12345"},
			body:        `{"cells":[{"type":"logs","data":"service:api","title":"API Logs"}]}`,
			wantErr:     true,
			errContains: "no real API connection in tests",
		},
		{
			name:        "empty cells",
			args:        []string{"12345"},
			body:        `{"cells":[]}`,
			wantErr:     true,
			errContains: "no cells provided",
		},
		{
			name:        "unknown cell type",
			args:        []string{"12345"},
			body:        `{"cells":[{"type":"unknown","data":"test"}]}`,
			wantErr:     true,
			errContains: "unknown cell type",
		},
		{
			name:        "multiple cells fails on client creation",
			args:        []string{"12345"},
			body:        `{"cells":[{"type":"markdown","data":"# Title"},{"type":"metric","data":"avg:system.cpu.user{*}"}]}`,
			wantErr:     true,
			errContains: "no real API connection in tests",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			origReader := inputReader
			inputReader = strings.NewReader(tt.body)
			defer func() { inputReader = origReader }()

			cmd := notebooksCellsAppendCmd
			cmd.Flags().Set("body", "-")

			err := runNotebooksCellsAppend(cmd, tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("runNotebooksCellsAppend() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.errContains != "" && err != nil && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("error = %v, want to contain %q", err, tt.errContains)
			}
		})
	}
}

func TestRunNotebooksCellsAppend_InvalidJSON(t *testing.T) {
	cleanup := setupNotebooksTestClient(t)
	defer cleanup()

	var buf bytes.Buffer
	outputWriter = &buf
	defer func() { outputWriter = os.Stdout }()

	origReader := inputReader
	inputReader = strings.NewReader("not json")
	defer func() { inputReader = origReader }()

	cmd := notebooksCellsAppendCmd
	cmd.Flags().Set("body", "-")

	err := runNotebooksCellsAppend(cmd, []string{"12345"})
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "invalid JSON in body") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBuildCellRequest_Markdown(t *testing.T) {
	cell := simpleCell{Type: simpleCellMarkdown, Data: "# Hello World"}
	result, err := buildCellRequest(cell)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.NotebookCellCreateRequest == nil {
		t.Fatal("expected NotebookCellCreateRequest to be set")
	}
	if result.NotebookCellCreateRequest.Attributes.NotebookMarkdownCellAttributes == nil {
		t.Fatal("expected markdown attributes to be set")
	}
	if result.NotebookCellCreateRequest.Attributes.NotebookMarkdownCellAttributes.Definition.Text != "# Hello World" {
		t.Errorf("text = %q, want '# Hello World'", result.NotebookCellCreateRequest.Attributes.NotebookMarkdownCellAttributes.Definition.Text)
	}
}

func TestBuildCellRequest_Metric(t *testing.T) {
	cell := simpleCell{
		Type:  simpleCellMetric,
		Data:  "avg:system.cpu.user{*}",
		Title: "CPU Usage",
		Start: "2025-01-01T00:00:00Z",
		End:   "2025-01-01T01:00:00Z",
	}
	result, err := buildCellRequest(cell)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.NotebookCellCreateRequest == nil {
		t.Fatal("expected NotebookCellCreateRequest to be set")
	}
	if result.NotebookCellCreateRequest.Attributes.NotebookTimeseriesCellAttributes == nil {
		t.Fatal("expected timeseries attributes to be set")
	}

	// Verify it marshals to valid JSON (can be sent to API)
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	if !json.Valid(data) {
		t.Error("marshaled result is not valid JSON")
	}
}

func TestBuildCellRequest_Logs(t *testing.T) {
	cell := simpleCell{
		Type:  simpleCellLogs,
		Data:  "service:api status:error",
		Title: "API Errors",
	}
	result, err := buildCellRequest(cell)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.NotebookCellCreateRequest == nil {
		t.Fatal("expected NotebookCellCreateRequest to be set")
	}
	if result.NotebookCellCreateRequest.Attributes.NotebookLogStreamCellAttributes == nil {
		t.Fatal("expected log stream attributes to be set")
	}
}

func TestBuildCellRequest_LogsDefaultTitle(t *testing.T) {
	cell := simpleCell{Type: simpleCellLogs, Data: "service:api"}
	result, err := buildCellRequest(cell)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	title := result.NotebookCellCreateRequest.Attributes.NotebookLogStreamCellAttributes.Definition.Title
	if title == nil || *title != "logs" {
		t.Errorf("default title = %v, want 'logs'", title)
	}
}

func TestBuildCellRequest_UnknownType(t *testing.T) {
	cell := simpleCell{Type: "invalid", Data: "test"}
	_, err := buildCellRequest(cell)
	if err == nil {
		t.Fatal("expected error for unknown cell type")
	}
	if !strings.Contains(err.Error(), "unknown cell type") {
		t.Errorf("error = %v, want to contain 'unknown cell type'", err)
	}
}

func TestParseCellTimes_Defaults(t *testing.T) {
	start, end := parseCellTimes("", "")
	if end.Before(start) {
		t.Error("end should not be before start")
	}
	diff := end.Sub(start)
	if diff < 59*time.Minute || diff > 61*time.Minute {
		t.Errorf("default time range = %v, want ~1h", diff)
	}
}

func TestParseCellTimes_Custom(t *testing.T) {
	start, end := parseCellTimes("2025-01-01T00:00:00Z", "2025-01-01T02:00:00Z")
	if start.Year() != 2025 || start.Month() != 1 || start.Day() != 1 {
		t.Errorf("start = %v, want 2025-01-01", start)
	}
	if end.Sub(start) != 2*time.Hour {
		t.Errorf("duration = %v, want 2h", end.Sub(start))
	}
}
