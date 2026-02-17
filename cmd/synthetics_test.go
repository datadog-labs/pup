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

func TestSyntheticsCmd(t *testing.T) {
	if syntheticsCmd == nil {
		t.Fatal("syntheticsCmd is nil")
	}
}

func setupSyntheticsTestClient(t *testing.T) func() {
	t.Helper()
	origClient, origCfg, origFactory := ddClient, cfg, clientFactory
	cfg = &config.Config{Site: "datadoghq.com", APIKey: "test-key-12345678", AppKey: "test-key-12345678"}
	clientFactory = func(c *config.Config) (*client.Client, error) {
		return nil, fmt.Errorf("mock client: no real API connection")
	}
	ddClient = nil
	return func() { ddClient, cfg, clientFactory = origClient, origCfg, origFactory }
}

func TestRunSyntheticsTestsList(t *testing.T) {
	cleanup := setupSyntheticsTestClient(t)
	defer cleanup()
	var buf bytes.Buffer
	outputWriter = &buf
	defer func() { outputWriter = os.Stdout }()
	err := runSyntheticsTestsList(syntheticsTestsListCmd, []string{})
	if err == nil {
		t.Error("Expected error with mock client")
	}
}

func TestRunSyntheticsTestsGet(t *testing.T) {
	cleanup := setupSyntheticsTestClient(t)
	defer cleanup()
	var buf bytes.Buffer
	outputWriter = &buf
	defer func() { outputWriter = os.Stdout }()
	err := runSyntheticsTestsGet(syntheticsTestsGetCmd, []string{"test-123"})
	if err == nil {
		t.Error("Expected error with mock client")
	}
}

func TestRunSyntheticsTestsSearch(t *testing.T) {
	cleanup := setupSyntheticsTestClient(t)
	defer cleanup()
	var buf bytes.Buffer
	outputWriter = &buf
	defer func() { outputWriter = os.Stdout }()
	err := runSyntheticsTestsSearch(syntheticsTestsSearchCmd, []string{})
	if err == nil {
		t.Error("Expected error with mock client")
	}
}

func TestRunSyntheticsLocationsList(t *testing.T) {
	cleanup := setupSyntheticsTestClient(t)
	defer cleanup()
	var buf bytes.Buffer
	outputWriter = &buf
	defer func() { outputWriter = os.Stdout }()
	err := runSyntheticsLocationsList(syntheticsLocationsListCmd, []string{})
	if err == nil {
		t.Error("Expected error with mock client")
	}
}
