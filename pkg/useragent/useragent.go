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

// Get returns the user agent string for pup CLI with optional AI agent detection.
//
// Format without agent:
//
//	pup/v0.1.0 (go go1.25.0; os darwin; arch arm64)
//
// Format with agent:
//
//	pup/v0.1.0 (go go1.25.0; os darwin; arch arm64; ai-agent claude-code)
//
// AI agents are detected via environment variables:
//   - CLAUDECODE=1 or CLAUDE_CODE=1 → adds "ai-agent claude-code"
//   - CURSOR_AGENT=true or CURSOR_AGENT=1 → adds "ai-agent cursor"
//
// If multiple agents are detected, CLAUDECODE takes precedence.
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

// detectAgent detects AI coding assistant from environment variables.
// Returns empty string if no agent is detected.
func detectAgent() string {
	// Check Claude Code (CLAUDECODE or CLAUDE_CODE)
	if os.Getenv("CLAUDECODE") == "1" || os.Getenv("CLAUDE_CODE") == "1" {
		return "claude-code"
	}

	// Check Cursor (CURSOR_AGENT=true or CURSOR_AGENT=1)
	cursorAgent := strings.ToLower(os.Getenv("CURSOR_AGENT"))
	if cursorAgent == "true" || cursorAgent == "1" {
		return "cursor"
	}

	return ""
}
