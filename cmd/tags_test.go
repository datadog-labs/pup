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

func TestTagsCmd(t *testing.T) {
	if tagsCmd == nil {
		t.Fatal("tagsCmd is nil")
	}
}

func setupTagsTestClient(t *testing.T) func() {
	t.Helper()
	origClient, origCfg, origFactory := ddClient, cfg, clientFactory
	cfg = &config.Config{Site: "datadoghq.com", APIKey: "test-key-12345678", AppKey: "test-key-12345678"}
	clientFactory = func(c *config.Config) (*client.Client, error) {
		return nil, fmt.Errorf("mock client: no real API connection")
	}
	ddClient = nil
	return func() { ddClient, cfg, clientFactory = origClient, origCfg, origFactory }
}

func TestRunTagsList(t *testing.T) {
	cleanup := setupTagsTestClient(t)
	defer cleanup()
	var buf bytes.Buffer
	outputWriter = &buf
	defer func() { outputWriter = os.Stdout }()
	err := runTagsList(tagsListCmd, []string{})
	if err == nil {
		t.Error("Expected error with mock client")
	}
}

func TestRunTagsGet(t *testing.T) {
	cleanup := setupTagsTestClient(t)
	defer cleanup()
	var buf bytes.Buffer
	outputWriter = &buf
	defer func() { outputWriter = os.Stdout }()
	err := runTagsGet(tagsGetCmd, []string{"host:web-1"})
	if err == nil {
		t.Error("Expected error with mock client")
	}
}

func TestRunTagsAdd(t *testing.T) {
	cleanup := setupTagsTestClient(t)
	defer cleanup()
	var buf bytes.Buffer
	outputWriter = &buf
	defer func() { outputWriter = os.Stdout }()
	err := runTagsAdd(tagsAddCmd, []string{"host:web-1"})
	if err == nil {
		t.Error("Expected error with mock client")
	}
}

func TestRunTagsUpdate(t *testing.T) {
	cleanup := setupTagsTestClient(t)
	defer cleanup()
	var buf bytes.Buffer
	outputWriter = &buf
	defer func() { outputWriter = os.Stdout }()
	err := runTagsUpdate(tagsUpdateCmd, []string{"host:web-1"})
	if err == nil {
		t.Error("Expected error with mock client")
	}
}

func TestRunTagsDelete(t *testing.T) {
	cleanup := setupTagsTestClient(t)
	defer cleanup()
	var buf bytes.Buffer
	outputWriter = &buf
	defer func() { outputWriter = os.Stdout }()
	err := runTagsDelete(tagsDeleteCmd, []string{"host:web-1"})
	if err == nil {
		t.Error("Expected error with mock client")
	}
}
