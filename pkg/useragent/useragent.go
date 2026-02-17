// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package useragent

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/DataDog/pup/internal/version"
)

// AgentInfo contains information about a detected AI coding agent.
type AgentInfo struct {
	Name     string
	Detected bool
}

// agentDetector defines how to detect a specific AI coding agent.
type agentDetector struct {
	Name    string
	EnvVars []string
}

// agentDetectors is a table-driven registry of known AI coding agents.
// Order matters: first match wins when multiple agents are detected.
var agentDetectors = []agentDetector{
	{Name: "claude-code", EnvVars: []string{"CLAUDECODE", "CLAUDE_CODE"}},
	{Name: "cursor", EnvVars: []string{"CURSOR_AGENT"}},
	{Name: "codex", EnvVars: []string{"CODEX", "OPENAI_CODEX"}},
	{Name: "opencode", EnvVars: []string{"OPENCODE"}},
	{Name: "aider", EnvVars: []string{"AIDER"}},
	{Name: "cline", EnvVars: []string{"CLINE"}},
	{Name: "windsurf", EnvVars: []string{"WINDSURF_AGENT"}},
	{Name: "github-copilot", EnvVars: []string{"GITHUB_COPILOT"}},
	{Name: "amazon-q", EnvVars: []string{"AMAZON_Q", "AWS_Q_DEVELOPER"}},
	{Name: "gemini-code", EnvVars: []string{"GEMINI_CODE_ASSIST"}},
	{Name: "sourcegraph-cody", EnvVars: []string{"SRC_CODY"}},
}

// Get returns the user agent string for pup CLI with optional AI agent detection.
//
// Format without agent:
//
//	pup/v0.1.0 (go go1.25.0; os darwin; arch arm64)
//
// Format with agent:
//
//	pup/v0.1.0 (go go1.25.0; os darwin; arch arm64; ai-agent claude-code)
func Get() string {
	base := fmt.Sprintf(
		"pup/%s (go %s; os %s; arch %s",
		version.Version,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
	)

	if agent := detectAgent(); agent != "" {
		return base + fmt.Sprintf("; ai-agent %s)", agent)
	}
	return base + ")"
}

// IsAgentMode returns true if any AI agent is detected or FORCE_AGENT_MODE=1 is set.
func IsAgentMode() bool {
	if isEnvTruthy("FORCE_AGENT_MODE") {
		return true
	}
	return detectAgent() != ""
}

// DetectAgentInfo returns information about the detected AI coding agent.
func DetectAgentInfo() AgentInfo {
	agent := detectAgent()
	if agent == "" {
		return AgentInfo{}
	}
	return AgentInfo{Name: agent, Detected: true}
}

// detectAgent detects AI coding assistant from environment variables.
// Returns empty string if no agent is detected.
func detectAgent() string {
	for _, d := range agentDetectors {
		for _, envVar := range d.EnvVars {
			if isEnvTruthy(envVar) {
				return d.Name
			}
		}
	}
	return ""
}

// isEnvTruthy checks if an environment variable is set to a truthy value.
func isEnvTruthy(key string) bool {
	val := strings.ToLower(os.Getenv(key))
	return val == "1" || val == "true"
}
