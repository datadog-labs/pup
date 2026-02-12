// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package useragent

import (
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/DataDog/pup/internal/version"
)

func TestGet_NoAgent(t *testing.T) {
	// NOTE: Not parallel - modifies env vars
	// Clear all agent environment variables
	os.Unsetenv("CLAUDECODE")
	os.Unsetenv("CLAUDE_CODE")
	os.Unsetenv("CURSOR_AGENT")

	result := Get()

	// Check format matches: pup/VERSION (go GOVERSION; os OS; arch ARCH)
	expected := "pup/" + version.Version + " (go " + runtime.Version() + "; os " + runtime.GOOS + "; arch " + runtime.GOARCH + ")"
	if result != expected {
		t.Errorf("Get() = %q, want %q", result, expected)
	}

	// Verify no agent in output
	if strings.Contains(result, "ai-agent") {
		t.Errorf("Get() should not contain ai-agent, got %q", result)
	}
}

func TestGet_WithClaudeCode(t *testing.T) {
	// NOTE: Not parallel - modifies env vars
	tests := []struct {
		name    string
		envVar  string
		envVal  string
		wantSuf string
	}{
		{"CLAUDECODE=1", "CLAUDECODE", "1", "claude-code"},
		{"CLAUDE_CODE=1", "CLAUDE_CODE", "1", "claude-code"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all agent env vars first
			os.Unsetenv("CLAUDECODE")
			os.Unsetenv("CLAUDE_CODE")
			os.Unsetenv("CURSOR_AGENT")

			// Set test env var
			os.Setenv(tt.envVar, tt.envVal)
			defer os.Unsetenv(tt.envVar)

			result := Get()

			// Verify agent is present in structured format
			expectedAgent := "; ai-agent " + tt.wantSuf + ")"
			if !strings.HasSuffix(result, expectedAgent) {
				t.Errorf("Get() = %q, want suffix %q", result, expectedAgent)
			}

			// Verify base format is still correct
			expectedBase := "pup/" + version.Version + " (go " + runtime.Version() + "; os " + runtime.GOOS + "; arch " + runtime.GOARCH
			if !strings.HasPrefix(result, expectedBase) {
				t.Errorf("Get() = %q, want prefix %q", result, expectedBase)
			}
		})
	}
}

func TestGet_WithCursor(t *testing.T) {
	tests := []struct {
		name   string
		envVal string
	}{
		{"CURSOR_AGENT=true", "true"},
		{"CURSOR_AGENT=TRUE", "TRUE"},
		{"CURSOR_AGENT=1", "1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all agent env vars first
			os.Unsetenv("CLAUDECODE")
			os.Unsetenv("CLAUDE_CODE")
			os.Unsetenv("CURSOR_AGENT")

			// Set test env var
			os.Setenv("CURSOR_AGENT", tt.envVal)
			defer os.Unsetenv("CURSOR_AGENT")

			result := Get()

			// Verify agent is present in structured format
			if !strings.HasSuffix(result, "; ai-agent cursor)") {
				t.Errorf("Get() = %q, want suffix '; ai-agent cursor)'", result)
			}

			// Verify base format is still correct
			expectedBase := "pup/" + version.Version + " (go " + runtime.Version() + "; os " + runtime.GOOS + "; arch " + runtime.GOARCH
			if !strings.HasPrefix(result, expectedBase) {
				t.Errorf("Get() = %q, want prefix %q", result, expectedBase)
			}
		})
	}
}

func TestGet_WithMultipleAgents(t *testing.T) {
	// Test precedence: CLAUDECODE should win when both are set
	os.Unsetenv("CLAUDECODE")
	os.Unsetenv("CLAUDE_CODE")
	os.Unsetenv("CURSOR_AGENT")

	os.Setenv("CLAUDECODE", "1")
	os.Setenv("CURSOR_AGENT", "true")
	defer func() {
		os.Unsetenv("CLAUDECODE")
		os.Unsetenv("CURSOR_AGENT")
	}()

	result := Get()

	// Should have ai-agent claude-code, not cursor
	if !strings.HasSuffix(result, "; ai-agent claude-code)") {
		t.Errorf("Get() = %q, want suffix '; ai-agent claude-code)' (CLAUDECODE should take precedence)", result)
	}
	if strings.Contains(result, "cursor") {
		t.Errorf("Get() = %q, should not contain 'cursor' when CLAUDECODE is set", result)
	}
}

func TestDetectAgent(t *testing.T) {
	tests := []struct {
		name     string
		setup    func()
		teardown func()
		want     string
	}{
		{
			name: "no agent",
			setup: func() {
				os.Unsetenv("CLAUDECODE")
				os.Unsetenv("CLAUDE_CODE")
				os.Unsetenv("CURSOR_AGENT")
			},
			teardown: func() {},
			want:     "",
		},
		{
			name: "CLAUDECODE=1",
			setup: func() {
				os.Unsetenv("CLAUDECODE")
				os.Unsetenv("CLAUDE_CODE")
				os.Unsetenv("CURSOR_AGENT")
				os.Setenv("CLAUDECODE", "1")
			},
			teardown: func() {
				os.Unsetenv("CLAUDECODE")
			},
			want: "claude-code",
		},
		{
			name: "CLAUDE_CODE=1",
			setup: func() {
				os.Unsetenv("CLAUDECODE")
				os.Unsetenv("CLAUDE_CODE")
				os.Unsetenv("CURSOR_AGENT")
				os.Setenv("CLAUDE_CODE", "1")
			},
			teardown: func() {
				os.Unsetenv("CLAUDE_CODE")
			},
			want: "claude-code",
		},
		{
			name: "CURSOR_AGENT=true",
			setup: func() {
				os.Unsetenv("CLAUDECODE")
				os.Unsetenv("CLAUDE_CODE")
				os.Unsetenv("CURSOR_AGENT")
				os.Setenv("CURSOR_AGENT", "true")
			},
			teardown: func() {
				os.Unsetenv("CURSOR_AGENT")
			},
			want: "cursor",
		},
		{
			name: "CURSOR_AGENT=1",
			setup: func() {
				os.Unsetenv("CLAUDECODE")
				os.Unsetenv("CLAUDE_CODE")
				os.Unsetenv("CURSOR_AGENT")
				os.Setenv("CURSOR_AGENT", "1")
			},
			teardown: func() {
				os.Unsetenv("CURSOR_AGENT")
			},
			want: "cursor",
		},
		{
			name: "CURSOR_AGENT=false (invalid, should not detect)",
			setup: func() {
				os.Unsetenv("CLAUDECODE")
				os.Unsetenv("CLAUDE_CODE")
				os.Unsetenv("CURSOR_AGENT")
				os.Setenv("CURSOR_AGENT", "false")
			},
			teardown: func() {
				os.Unsetenv("CURSOR_AGENT")
			},
			want: "",
		},
		{
			name: "multiple agents (CLAUDECODE precedence)",
			setup: func() {
				os.Unsetenv("CLAUDECODE")
				os.Unsetenv("CLAUDE_CODE")
				os.Unsetenv("CURSOR_AGENT")
				os.Setenv("CLAUDECODE", "1")
				os.Setenv("CURSOR_AGENT", "true")
			},
			teardown: func() {
				os.Unsetenv("CLAUDECODE")
				os.Unsetenv("CURSOR_AGENT")
			},
			want: "claude-code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer tt.teardown()

			got := detectAgent()
			if got != tt.want {
				t.Errorf("detectAgent() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGet_Format(t *testing.T) {
	// Clear all agent env vars
	os.Unsetenv("CLAUDECODE")
	os.Unsetenv("CLAUDE_CODE")
	os.Unsetenv("CURSOR_AGENT")

	result := Get()

	// Verify format: pup/VERSION (go GOVERSION; os OS; arch ARCH)
	if !strings.HasPrefix(result, "pup/") {
		t.Errorf("Get() should start with 'pup/', got %q", result)
	}

	if !strings.Contains(result, "(go ") {
		t.Errorf("Get() should contain '(go ', got %q", result)
	}

	if !strings.Contains(result, "; os ") {
		t.Errorf("Get() should contain '; os ', got %q", result)
	}

	if !strings.Contains(result, "; arch ") {
		t.Errorf("Get() should contain '; arch ', got %q", result)
	}

	// Verify no extra spaces or malformed output
	if strings.Contains(result, "  ") {
		t.Errorf("Get() should not contain double spaces, got %q", result)
	}

	// Test with agent
	os.Setenv("CLAUDECODE", "1")
	defer os.Unsetenv("CLAUDECODE")

	resultWithAgent := Get()

	// Verify agent is in structured format
	expectedSuffix := "; ai-agent claude-code)"
	if !strings.HasSuffix(resultWithAgent, expectedSuffix) {
		t.Errorf("Get() with agent should end with %q, got %q", expectedSuffix, resultWithAgent)
	}

	// Verify base part is present
	expectedBase := "pup/" + version.Version + " (go " + runtime.Version() + "; os " + runtime.GOOS + "; arch " + runtime.GOARCH
	if !strings.HasPrefix(resultWithAgent, expectedBase) {
		t.Errorf("Get() with agent should start with %q, got %q", expectedBase, resultWithAgent)
	}
}
