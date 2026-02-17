// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package agenthelp

import "testing"

func TestGetQuerySyntax(t *testing.T) {
	syntax := GetQuerySyntax()
	if len(syntax) == 0 {
		t.Fatal("GetQuerySyntax() should not be empty")
	}

	expectedDomains := []string{"logs", "metrics", "monitors", "apm", "rum", "security", "events", "traces"}
	for _, domain := range expectedDomains {
		if _, ok := syntax[domain]; !ok {
			t.Errorf("GetQuerySyntax() missing domain %q", domain)
		}
	}
}

func TestGetTimeFormats(t *testing.T) {
	tf := GetTimeFormats()
	if len(tf.Relative) == 0 {
		t.Error("TimeFormats.Relative should not be empty")
	}
	if len(tf.Absolute) == 0 {
		t.Error("TimeFormats.Absolute should not be empty")
	}
	if len(tf.Examples) == 0 {
		t.Error("TimeFormats.Examples should not be empty")
	}
}

func TestGetWorkflows(t *testing.T) {
	workflows := GetWorkflows()
	if len(workflows) == 0 {
		t.Fatal("GetWorkflows() should not be empty")
	}
	for _, w := range workflows {
		if w.Name == "" {
			t.Error("Workflow name should not be empty")
		}
		if len(w.Steps) == 0 {
			t.Errorf("Workflow %q should have steps", w.Name)
		}
	}
}

func TestGetBestPractices(t *testing.T) {
	bp := GetBestPractices()
	if len(bp) == 0 {
		t.Fatal("GetBestPractices() should not be empty")
	}
	for _, p := range bp {
		if p == "" {
			t.Error("Best practice should not be empty string")
		}
	}
}

func TestGetAntiPatterns(t *testing.T) {
	ap := GetAntiPatterns()
	if len(ap) == 0 {
		t.Fatal("GetAntiPatterns() should not be empty")
	}
	for _, p := range ap {
		if p == "" {
			t.Error("Anti-pattern should not be empty string")
		}
	}
}
