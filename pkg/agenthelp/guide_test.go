// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package agenthelp

import (
	"strings"
	"testing"
)

func TestGetGuide(t *testing.T) {
	guide := GetGuide()
	if guide == "" {
		t.Fatal("GetGuide() should not be empty")
	}
	if !strings.Contains(guide, "# Pup Agent Guide") {
		t.Error("Guide should contain the title")
	}
	if !strings.Contains(guide, "## Logs") {
		t.Error("Guide should contain Logs section")
	}
	if !strings.Contains(guide, "## Metrics") {
		t.Error("Guide should contain Metrics section")
	}
	if !strings.Contains(guide, "## Monitors") {
		t.Error("Guide should contain Monitors section")
	}
}

func TestGetGuideSection(t *testing.T) {
	tests := []struct {
		domain      string
		shouldMatch string
	}{
		{"logs", "## Logs"},
		{"metrics", "## Metrics"},
		{"monitors", "## Monitors"},
		{"apm", "## APM"},
		{"Logs", "## Logs"},
	}

	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			section := GetGuideSection(tt.domain)
			if !strings.Contains(section, tt.shouldMatch) {
				t.Errorf("GetGuideSection(%q) should contain %q", tt.domain, tt.shouldMatch)
			}
		})
	}
}

func TestGetGuideSection_NotFound(t *testing.T) {
	section := GetGuideSection("nonexistent_domain_xyz")
	guide := GetGuide()
	if section != guide {
		t.Error("GetGuideSection for unknown domain should return the full guide")
	}
}
