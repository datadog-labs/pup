// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"testing"
)

func TestServiceCatalogCmd(t *testing.T) {
	if serviceCatalogCmd == nil {
		t.Fatal("serviceCatalogCmd is nil")
	}

	if serviceCatalogCmd.Use != "service-catalog" {
		t.Errorf("Use = %s, want service-catalog", serviceCatalogCmd.Use)
	}

	if serviceCatalogCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if serviceCatalogCmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestServiceCatalogCmd_Subcommands(t *testing.T) {
	expectedCommands := []string{"list", "get"}

	commands := serviceCatalogCmd.Commands()

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

func TestServiceCatalogListCmd(t *testing.T) {
	if serviceCatalogListCmd == nil {
		t.Fatal("serviceCatalogListCmd is nil")
	}

	if serviceCatalogListCmd.Use != "list" {
		t.Errorf("Use = %s, want list", serviceCatalogListCmd.Use)
	}

	if serviceCatalogListCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if serviceCatalogListCmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestServiceCatalogGetCmd(t *testing.T) {
	if serviceCatalogGetCmd == nil {
		t.Fatal("serviceCatalogGetCmd is nil")
	}

	if serviceCatalogGetCmd.Use != "get [service-name]" {
		t.Errorf("Use = %s, want 'get [service-name]'", serviceCatalogGetCmd.Use)
	}

	if serviceCatalogGetCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if serviceCatalogGetCmd.RunE == nil {
		t.Error("RunE is nil")
	}

	if serviceCatalogGetCmd.Args == nil {
		t.Error("Args validator is nil")
	}
}

func TestServiceCatalogCmd_ParentChild(t *testing.T) {
	commands := serviceCatalogCmd.Commands()

	for _, cmd := range commands {
		if cmd.Parent() != serviceCatalogCmd {
			t.Errorf("Command %s parent is not serviceCatalogCmd", cmd.Use)
		}
	}
}
