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

// allAgentEnvVars returns all env vars used by agent detectors plus FORCE_AGENT_MODE.
func allAgentEnvVars() []string {
	vars := []string{"FORCE_AGENT_MODE"}
	for _, d := range agentDetectors {
		vars = append(vars, d.EnvVars...)
	}
	return vars
}

// clearAllAgentEnvVars unsets every agent-related env var.
func clearAllAgentEnvVars() {
	for _, v := range allAgentEnvVars() {
		os.Unsetenv(v)
	}
}

func TestGet_NoAgent(t *testing.T) {
	clearAllAgentEnvVars()

	result := Get()

	expected := "pup/" + version.Version + " (go " + runtime.Version() + "; os " + runtime.GOOS + "; arch " + runtime.GOARCH + ")"
	if result != expected {
		t.Errorf("Get() = %q, want %q", result, expected)
	}

	if strings.Contains(result, "ai-agent") {
		t.Errorf("Get() should not contain ai-agent, got %q", result)
	}
}

func TestGet_AllAgents(t *testing.T) {
	tests := []struct {
		name     string
		envVar   string
		envVal   string
		wantName string
	}{
		{"CLAUDECODE=1", "CLAUDECODE", "1", "claude-code"},
		{"CLAUDE_CODE=1", "CLAUDE_CODE", "1", "claude-code"},
		{"CURSOR_AGENT=true", "CURSOR_AGENT", "true", "cursor"},
		{"CURSOR_AGENT=1", "CURSOR_AGENT", "1", "cursor"},
		{"CODEX=1", "CODEX", "1", "codex"},
		{"OPENAI_CODEX=1", "OPENAI_CODEX", "1", "codex"},
		{"OPENCODE=1", "OPENCODE", "1", "opencode"},
		{"AIDER=1", "AIDER", "1", "aider"},
		{"CLINE=1", "CLINE", "1", "cline"},
		{"WINDSURF_AGENT=1", "WINDSURF_AGENT", "1", "windsurf"},
		{"GITHUB_COPILOT=1", "GITHUB_COPILOT", "1", "github-copilot"},
		{"AMAZON_Q=1", "AMAZON_Q", "1", "amazon-q"},
		{"AWS_Q_DEVELOPER=true", "AWS_Q_DEVELOPER", "true", "amazon-q"},
		{"GEMINI_CODE_ASSIST=1", "GEMINI_CODE_ASSIST", "1", "gemini-code"},
		{"SRC_CODY=1", "SRC_CODY", "1", "sourcegraph-cody"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearAllAgentEnvVars()
			os.Setenv(tt.envVar, tt.envVal)
			defer os.Unsetenv(tt.envVar)

			result := Get()

			expectedSuffix := "; ai-agent " + tt.wantName + ")"
			if !strings.HasSuffix(result, expectedSuffix) {
				t.Errorf("Get() = %q, want suffix %q", result, expectedSuffix)
			}

			expectedBase := "pup/" + version.Version + " (go " + runtime.Version() + "; os " + runtime.GOOS + "; arch " + runtime.GOARCH
			if !strings.HasPrefix(result, expectedBase) {
				t.Errorf("Get() = %q, want prefix %q", result, expectedBase)
			}
		})
	}
}

func TestGet_WithMultipleAgents(t *testing.T) {
	clearAllAgentEnvVars()
	os.Setenv("CLAUDECODE", "1")
	os.Setenv("CURSOR_AGENT", "true")
	defer func() {
		os.Unsetenv("CLAUDECODE")
		os.Unsetenv("CURSOR_AGENT")
	}()

	result := Get()

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
		envVar   string
		envVal   string
		want     string
	}{
		{"no agent", "", "", ""},
		{"CLAUDECODE=1", "CLAUDECODE", "1", "claude-code"},
		{"CLAUDE_CODE=1", "CLAUDE_CODE", "1", "claude-code"},
		{"CURSOR_AGENT=true", "CURSOR_AGENT", "true", "cursor"},
		{"CURSOR_AGENT=1", "CURSOR_AGENT", "1", "cursor"},
		{"CURSOR_AGENT=false (not truthy)", "CURSOR_AGENT", "false", ""},
		{"CODEX=1", "CODEX", "1", "codex"},
		{"AIDER=true", "AIDER", "true", "aider"},
		{"WINDSURF_AGENT=1", "WINDSURF_AGENT", "1", "windsurf"},
		{"GITHUB_COPILOT=1", "GITHUB_COPILOT", "1", "github-copilot"},
		{"AMAZON_Q=1", "AMAZON_Q", "1", "amazon-q"},
		{"GEMINI_CODE_ASSIST=1", "GEMINI_CODE_ASSIST", "1", "gemini-code"},
		{"SRC_CODY=1", "SRC_CODY", "1", "sourcegraph-cody"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearAllAgentEnvVars()
			if tt.envVar != "" {
				os.Setenv(tt.envVar, tt.envVal)
				defer os.Unsetenv(tt.envVar)
			}

			got := detectAgent()
			if got != tt.want {
				t.Errorf("detectAgent() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsAgentMode(t *testing.T) {
	tests := []struct {
		name   string
		envVar string
		envVal string
		want   bool
	}{
		{"no env", "", "", false},
		{"FORCE_AGENT_MODE=1", "FORCE_AGENT_MODE", "1", true},
		{"FORCE_AGENT_MODE=true", "FORCE_AGENT_MODE", "true", true},
		{"FORCE_AGENT_MODE=false", "FORCE_AGENT_MODE", "false", false},
		{"CLAUDECODE=1", "CLAUDECODE", "1", true},
		{"CURSOR_AGENT=1", "CURSOR_AGENT", "1", true},
		{"AIDER=1", "AIDER", "1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearAllAgentEnvVars()
			if tt.envVar != "" {
				os.Setenv(tt.envVar, tt.envVal)
				defer os.Unsetenv(tt.envVar)
			}

			got := IsAgentMode()
			if got != tt.want {
				t.Errorf("IsAgentMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectAgentInfo(t *testing.T) {
	tests := []struct {
		name   string
		envVar string
		envVal string
		want   AgentInfo
	}{
		{"no agent", "", "", AgentInfo{}},
		{"claude-code", "CLAUDECODE", "1", AgentInfo{Name: "claude-code", Detected: true}},
		{"cursor", "CURSOR_AGENT", "1", AgentInfo{Name: "cursor", Detected: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearAllAgentEnvVars()
			if tt.envVar != "" {
				os.Setenv(tt.envVar, tt.envVal)
				defer os.Unsetenv(tt.envVar)
			}

			got := DetectAgentInfo()
			if got != tt.want {
				t.Errorf("DetectAgentInfo() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestGet_Format(t *testing.T) {
	clearAllAgentEnvVars()

	result := Get()

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
	if strings.Contains(result, "  ") {
		t.Errorf("Get() should not contain double spaces, got %q", result)
	}

	os.Setenv("CLAUDECODE", "1")
	defer os.Unsetenv("CLAUDECODE")

	resultWithAgent := Get()
	expectedSuffix := "; ai-agent claude-code)"
	if !strings.HasSuffix(resultWithAgent, expectedSuffix) {
		t.Errorf("Get() with agent should end with %q, got %q", expectedSuffix, resultWithAgent)
	}

	expectedBase := "pup/" + version.Version + " (go " + runtime.Version() + "; os " + runtime.GOOS + "; arch " + runtime.GOARCH
	if !strings.HasPrefix(resultWithAgent, expectedBase) {
		t.Errorf("Get() with agent should start with %q, got %q", expectedBase, resultWithAgent)
	}
}

func TestIsEnvTruthy(t *testing.T) {
	tests := []struct {
		val  string
		want bool
	}{
		{"1", true},
		{"true", true},
		{"TRUE", true},
		{"True", true},
		{"0", false},
		{"false", false},
		{"", false},
		{"yes", false},
	}

	for _, tt := range tests {
		t.Run(tt.val, func(t *testing.T) {
			os.Setenv("TEST_TRUTHY", tt.val)
			defer os.Unsetenv("TEST_TRUTHY")

			got := isEnvTruthy("TEST_TRUTHY")
			if got != tt.want {
				t.Errorf("isEnvTruthy(%q) = %v, want %v", tt.val, got, tt.want)
			}
		})
	}
}
