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

func TestSlosCmd(t *testing.T) {
	if slosCmd == nil {
		t.Fatal("slosCmd is nil")
	}

	if slosCmd.Use != "slos" {
		t.Errorf("Use = %s, want slos", slosCmd.Use)
	}
}

// Helper function to setup SLO test client
func setupSlosTestClient(t *testing.T) func() {
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

func TestRunSlosList(t *testing.T) {
	cleanup := setupSlosTestClient(t)
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

			err := runSlosList(slosListCmd, []string{})

			if (err != nil) != tt.wantErr {
				t.Errorf("runSlosList() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunSlosGet(t *testing.T) {
	cleanup := setupSlosTestClient(t)
	defer cleanup()

	tests := []struct {
		name    string
		sloID   string
		wantErr bool
	}{
		{
			name:    "with valid SLO ID",
			sloID:   "abc123",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			err := runSlosGet(slosGetCmd, []string{tt.sloID})

			if (err != nil) != tt.wantErr {
				t.Errorf("runSlosGet() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunSlosDelete_AutoApprove(t *testing.T) {
	cleanup := setupSlosTestClient(t)
	defer cleanup()

	cfg.AutoApprove = true

	tests := []struct {
		name    string
		sloID   string
		wantErr bool
	}{
		{
			name:    "with auto-approve",
			sloID:   "abc123",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			err := runSlosDelete(slosDeleteCmd, []string{tt.sloID})

			if (err != nil) != tt.wantErr {
				t.Errorf("runSlosDelete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSlosStatusCmd(t *testing.T) {
	if slosStatusCmd == nil {
		t.Fatal("slosStatusCmd is nil")
	}
	if slosStatusCmd.Use != "status [slo-id]" {
		t.Errorf("Use = %s, want 'status [slo-id]'", slosStatusCmd.Use)
	}
}

func TestRunSlosStatus(t *testing.T) {
	cleanup := setupSlosTestClient(t)
	defer cleanup()

	sloStatusFrom = "1700000000"
	sloStatusTo = "1700003600"

	var buf bytes.Buffer
	outputWriter = &buf
	defer func() { outputWriter = os.Stdout }()

	err := runSlosStatus(slosStatusCmd, []string{"slo-123"})
	if err == nil {
		t.Error("expected error due to mock client, got nil")
	}
}

func TestRunSlosStatus_InvalidTimestamp(t *testing.T) {
	cleanup := setupSlosTestClient(t)
	defer cleanup()

	sloStatusFrom = "not-a-number"
	sloStatusTo = "1700003600"

	var buf bytes.Buffer
	outputWriter = &buf
	defer func() { outputWriter = os.Stdout }()

	err := runSlosStatus(slosStatusCmd, []string{"slo-123"})
	if err == nil {
		t.Error("expected error for invalid timestamp, got nil")
	}
}

func TestRunSlosDelete_WithConfirmation(t *testing.T) {
	cleanup := setupSlosTestClient(t)
	defer cleanup()

	cfg.AutoApprove = false

	tests := []struct {
		name    string
		sloID   string
		input   string
		wantErr bool
	}{
		{
			name:    "fails on client creation (mock)",
			sloID:   "abc123",
			input:   "n\n",
			wantErr: true,
		},
		{
			name:    "fails on client creation with yes (mock)",
			sloID:   "abc123",
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

			err := runSlosDelete(slosDeleteCmd, []string{tt.sloID})

			if (err != nil) != tt.wantErr {
				t.Errorf("runSlosDelete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
