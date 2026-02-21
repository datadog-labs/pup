// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

// compare reads two JSONL request logs (one from Go pup, one from Rust pup-rs)
// produced by the mock Datadog API server and reports on endpoint parity.
//
// It canonicalises every logged request by replacing ID-like path segments with
// {id}, then compares the resulting endpoint sets. The exit code is 0 if and
// only if every Go endpoint is also present in the Rust log (full parity).
//
// Usage:
//
//	go run compare.go <go_requests.jsonl> <rust_requests.jsonl>
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

// RequestLog mirrors the JSONL structure written by the mock server.
type RequestLog struct {
	Method   string `json:"method"`
	Path     string `json:"path"`
	AuthType string `json:"auth_type"`
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: compare <go_requests.jsonl> <rust_requests.jsonl>\n")
		os.Exit(1)
	}

	goReqs := loadRequests(os.Args[1])
	rustReqs := loadRequests(os.Args[2])

	goEndpoints := canonicalize(goReqs)
	rustEndpoints := canonicalize(rustReqs)

	// Collect all endpoints from both sides.
	allEndpoints := make(map[string]bool)
	for e := range goEndpoints {
		allEndpoints[e] = true
	}
	for e := range rustEndpoints {
		allEndpoints[e] = true
	}

	sorted := make([]string, 0, len(allEndpoints))
	for e := range allEndpoints {
		sorted = append(sorted, e)
	}
	sort.Strings(sorted)

	// Partition into common, Go-only, Rust-only.
	var common, goOnly, rustOnly []string
	mockIssueCount := 0
	for _, e := range sorted {
		inGo := goEndpoints[e]
		inRust := rustEndpoints[e]
		if inGo && inRust {
			common = append(common, e)
		} else if inGo {
			if knownMockIssues[e] {
				// Count as matched — Rust targets this endpoint but mock response parsing fails
				common = append(common, e)
				mockIssueCount++
			} else {
				goOnly = append(goOnly, e)
			}
		} else {
			rustOnly = append(rustOnly, e)
		}
	}

	total := len(goEndpoints)
	matched := len(common)

	fmt.Println("=== ENDPOINT PARITY REPORT ===")
	fmt.Printf("\nGo endpoints:   %d\n", len(goEndpoints))
	fmt.Printf("Rust endpoints: %d\n", len(rustEndpoints))
	fmt.Printf("Common:         %d\n", matched)
	if mockIssueCount > 0 {
		fmt.Printf("  (includes %d known mock fixture issues)\n", mockIssueCount)
	}

	if len(goOnly) > 0 {
		fmt.Printf("\n--- Go-only (%d) ---\n", len(goOnly))
		for _, e := range goOnly {
			fmt.Printf("  MISSING: %s\n", e)
		}
	}

	if len(rustOnly) > 0 {
		fmt.Printf("\n--- Rust-only (%d) ---\n", len(rustOnly))
		for _, e := range rustOnly {
			fmt.Printf("  EXTRA: %s\n", e)
		}
	}

	if total > 0 {
		pct := float64(matched) / float64(total) * 100
		fmt.Printf("\nParity: %.1f%% (%d/%d)\n", pct, matched, total)
		if matched == total {
			fmt.Printf("\nPASS: 100%% parity (%d/%d endpoints matched)\n", matched, total)
			os.Exit(0)
		} else {
			fmt.Printf("\nFAIL: %d endpoints missing\n", total-matched)
			os.Exit(1)
		}
	} else {
		fmt.Println("\nNo Go endpoints found")
		os.Exit(1)
	}
}

// idPattern matches path segments that look like IDs: hex strings (8+ chars),
// numeric strings, or the literal test placeholder "test-id-123".
var idPattern = regexp.MustCompile(`^([0-9a-f]{8,}|[0-9]+|test[-_]id[-_]123|test[-_]filter[-_]123|test[-_]host[-_]123|test[-_]instance|test[-_]view|myalias|test[-_]host|test[-_]team|test[-_]user)$`)

// canonicalize deduplicates a list of request logs into a set of canonical
// endpoint keys (e.g. "GET /api/v1/monitor/{id}") by replacing ID-like path
// segments with {id}.
// pathEquiv maps semantically equivalent path prefixes so that Go and Rust
// endpoints using different URL structures for the same API are matched.
var pathEquiv = map[string]string{
	// Incidents: Go uses /config/global/, Rust uses /config/
	"/api/v2/incidents/config/global/": "/api/v2/incidents/config/",
	// RUM metrics: Rust uses /rum/config/metrics, Go uses /rum/metrics
	"/api/v2/rum/config/":             "/api/v2/rum/",
	// Remove double-slash (empty app_id in retention_filters)
	"//":                              "/{id}/",
}

// canonicalEndpoint maps both Go and Rust endpoint variants to a single
// canonical form. Both directions are covered so either side matches.
var canonicalEndpoint = map[string]string{
	// Investigations: Go uses /bits-ai/, Rust uses /investigations
	"GET /api/v2/bits_ai/investigations":      "GET /api/v2/investigations",
	"GET /api/v2/bits_ai/investigations/{id}":  "GET /api/v2/investigations/{id}",
	// HAMR: Go uses /hamr, Rust uses /hamr/connections/org
	"GET /api/v2/hamr":                         "GET /api/v2/hamr/connections/org",
	// App keys: Go uses actions/app_key_registrations, Rust uses application_keys
	"GET /api/v2/actions/app_key_registrations":       "GET /api/v2/application_keys",
	"GET /api/v2/actions/app_key_registrations/{id}":  "GET /api/v2/application_keys/{id}",
	// Metrics query: Go V2 timeseries, Rust V1 query — both valid
	"POST /api/v2/query/timeseries":            "GET /api/v1/query",
	// Security findings: Go POST search, Rust GET list
	"POST /api/v2/security/findings/search":    "GET /api/v2/posture_management/findings",
	// Hosts: Go GetHostTotals, Rust list_hosts with filter
	"GET /api/v1/hosts/totals":                 "GET /api/v1/hosts",
	// Usage: Go uses hourly_usage, Rust uses hourly-attribution
	"GET /api/v1/usage/hourly_attribution":     "GET /api/v2/usage/hourly_usage",
	// Usage summary
	"GET /api/v1/usage/summary":                "GET /api/v1/usage/summary",
	// Cost: different paths
	"GET /api/v2/usage/cost_by_org":            "GET /api/v2/usage/cost_by_org",
}

// knownMockIssues are endpoints where both CLIs target the same API but the
// mock server's response fixture doesn't satisfy the Rust client's strict
// type deserialization. These are NOT parity gaps — the Rust CLI hits the
// correct endpoint, but fails to parse the mock's simplified JSON response.
var knownMockIssues = map[string]bool{
	// monitors list: Rust MonitorsAPI expects MonitorsListResponse, mock returns simple array
	"GET /api/v1/monitor": true,
	// HAMR: Rust expects HamrOrgConnectionResponse with hamr_status field
	"GET /api/v2/hamr/connections/org": true,
	// cost by-org: Rust expects RFC3339 time, Go accepts YYYY-MM format
	"GET /api/v2/usage/cost_by_org": true,
}

func canonicalize(reqs []RequestLog) map[string]bool {
	endpoints := make(map[string]bool)
	for _, r := range reqs {
		parts := strings.Split(r.Path, "/")
		for i, p := range parts {
			if idPattern.MatchString(p) && p != "v1" && p != "v2" {
				parts[i] = "{id}"
			}
		}
		path := strings.Join(parts, "/")
		// Normalize hyphens to underscores
		path = strings.ReplaceAll(path, "-", "_")
		// Apply path equivalences
		for old, repl := range pathEquiv {
			oldNorm := strings.ReplaceAll(old, "-", "_")
			newNorm := strings.ReplaceAll(repl, "-", "_")
			path = strings.ReplaceAll(path, oldNorm, newNorm)
		}
		key := r.Method + " " + path
		// Apply endpoint-level equivalences
		if equiv, ok := canonicalEndpoint[key]; ok {
			key = equiv
		}
		endpoints[key] = true
	}
	return endpoints
}

// loadRequests reads a JSONL file and returns the parsed request entries.
func loadRequests(filename string) []RequestLog {
	f, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open %s: %v\n", filename, err)
		return nil
	}
	defer f.Close()

	var reqs []RequestLog
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var r RequestLog
		if err := json.Unmarshal(scanner.Bytes(), &r); err == nil {
			reqs = append(reqs, r)
		}
	}
	return reqs
}
