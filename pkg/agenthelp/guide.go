// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package agenthelp

import (
	_ "embed"
	"strings"
)

//go:embed guide.md
var guideContent string

// GetGuide returns the full steering guide.
func GetGuide() string {
	return guideContent
}

// GetGuideSection returns a specific domain section from the guide.
// Returns the full guide if the domain is not found.
func GetGuideSection(domain string) string {
	// Try multiple casing strategies to find the section heading
	candidates := []string{
		"## " + strings.ToUpper(domain[:1]) + domain[1:], // "## Logs"
		"## " + strings.ToUpper(domain),                   // "## APM"
		"## " + domain,                                    // "## logs" (exact)
	}

	var idx int = -1
	var heading string
	for _, candidate := range candidates {
		idx = strings.Index(guideContent, candidate)
		if idx != -1 {
			heading = candidate
			break
		}
	}

	// Fall back to case-insensitive line scan
	if idx == -1 {
		lowerDomain := strings.ToLower(domain)
		for i, line := range strings.Split(guideContent, "\n") {
			if strings.HasPrefix(line, "## ") && strings.Contains(strings.ToLower(line), lowerDomain) {
				// Reconstruct idx from line number
				idx = 0
				for _, l := range strings.Split(guideContent, "\n")[:i] {
					idx += len(l) + 1
				}
				heading = line
				break
			}
		}
	}

	if idx == -1 {
		return guideContent
	}

	// Find the next ## heading after this one
	rest := guideContent[idx+len(heading):]
	nextSection := strings.Index(rest, "\n## ")
	if nextSection == -1 {
		return guideContent[idx:]
	}
	return guideContent[idx : idx+len(heading)+nextSection]
}
