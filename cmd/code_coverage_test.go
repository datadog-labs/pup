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

func TestCodeCoverageCmd(t *testing.T) {
	if codeCoverageCmd == nil {
		t.Fatal("codeCoverageCmd is nil")
	}
	if codeCoverageCmd.Use != "code-coverage" {
		t.Errorf("Use = %s, want code-coverage", codeCoverageCmd.Use)
	}
	if codeCoverageCmd.Short == "" {
		t.Error("Short description is empty")
	}
}

func TestCodeCoverageCmd_Subcommands(t *testing.T) {
	expectedCommands := []string{"branch-summary", "commit-summary"}
	commands := codeCoverageCmd.Commands()
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

func setupCodeCoverageTestClient(t *testing.T) func() {
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

func TestRunCodeCoverageBranchSummary(t *testing.T) {
	cleanup := setupCodeCoverageTestClient(t)
	defer cleanup()

	codeCoverageRepo = "github.com/org/repo"
	codeCoverageBranch = "main"

	var buf bytes.Buffer
	outputWriter = &buf
	defer func() { outputWriter = os.Stdout }()

	err := runCodeCoverageBranchSummary(codeCoverageBranchSummaryCmd, []string{})
	if err == nil {
		t.Error("expected error due to mock client, got nil")
	}
}

func TestRunCodeCoverageCommitSummary(t *testing.T) {
	cleanup := setupCodeCoverageTestClient(t)
	defer cleanup()

	codeCoverageRepo = "github.com/org/repo"
	codeCoverageCommit = "abc123"

	var buf bytes.Buffer
	outputWriter = &buf
	defer func() { outputWriter = os.Stdout }()

	err := runCodeCoverageCommitSummary(codeCoverageCommitSummaryCmd, []string{})
	if err == nil {
		t.Error("expected error due to mock client, got nil")
	}
}
