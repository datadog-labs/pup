// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DataDog/pup/pkg/client"
	"github.com/DataDog/pup/pkg/config"
	"github.com/DataDog/pup/pkg/util"
)

func TestMetricsCmd(t *testing.T) {
	if metricsCmd == nil {
		t.Fatal("metricsCmd is nil")
	}

	if metricsCmd.Use != "metrics" {
		t.Errorf("Use = %s, want metrics", metricsCmd.Use)
	}

	if metricsCmd.Short == "" {
		t.Error("Short description is empty")
	}
}

// Helper function to setup metrics test client
func setupMetricsTestClient(t *testing.T) func() {
	t.Helper()

	origClient := ddClient
	origCfg := cfg
	origFactory := clientFactory

	cfg = &config.Config{
		Site:        "datadoghq.com",
		APIKey:      "test-api-key-12345678",
		AppKey:      "test-app-key-12345678",
		AutoApprove: false,
	}

	clientFactory = func(c *config.Config) (*client.Client, error) {
		return nil, fmt.Errorf("mock client: no real API connection in tests")
	}

	ddClient = nil

	return func() {
		ddClient = origClient
		cfg = origCfg
		clientFactory = origFactory
	}
}

func TestRunMetricsSearch(t *testing.T) {
	cleanup := setupMetricsTestClient(t)
	defer cleanup()

	tests := []struct {
		name           string
		query          string
		from           string
		to             string
		wantErr        bool
		wantErrContains string
	}{
		{
			name:           "fails on client creation",
			query:          "avg:system.cpu.user{*}",
			from:           "1h",
			to:             "now",
			wantErr:        true,
			wantErrContains: "mock client",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queryString = tt.query
			fromTime = tt.from
			toTime = tt.to

			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			err := runMetricsSearch(metricsSearchCmd, []string{})

			if (err != nil) != tt.wantErr {
				t.Errorf("runMetricsSearch() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErrContains != "" && err != nil && !strings.Contains(err.Error(), tt.wantErrContains) {
				t.Errorf("runMetricsSearch() error = %v, want error containing %q", err, tt.wantErrContains)
			}
		})
	}
}

func TestRunMetricsQuery(t *testing.T) {
	cleanup := setupMetricsTestClient(t)
	defer cleanup()

	tests := []struct {
		name           string
		query          string
		from           string
		to             string
		wantErr        bool
		wantErrContains string
	}{
		{
			name:           "fails on client creation",
			query:          "avg:system.cpu.user{*}",
			from:           "1h",
			to:             "now",
			wantErr:        true,
			wantErrContains: "mock client",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queryString = tt.query
			fromTime = tt.from
			toTime = tt.to

			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			err := runMetricsQuery(metricsQueryCmd, []string{})

			if (err != nil) != tt.wantErr {
				t.Errorf("runMetricsQuery() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErrContains != "" && err != nil && !strings.Contains(err.Error(), tt.wantErrContains) {
				t.Errorf("runMetricsQuery() error = %v, want error containing %q", err, tt.wantErrContains)
			}
		})
	}
}

func TestMetricsSearchCmd(t *testing.T) {
	if metricsSearchCmd == nil {
		t.Fatal("metricsSearchCmd is nil")
	}

	if metricsSearchCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if metricsSearchCmd.RunE == nil {
		t.Error("RunE is nil")
	}

	// Verify search is registered as a subcommand of metricsCmd
	commands := metricsCmd.Commands()
	commandMap := make(map[string]bool)
	for _, cmd := range commands {
		commandMap[cmd.Name()] = true
	}
	if !commandMap["search"] {
		t.Error("search not registered as subcommand of metricsCmd")
	}

	// Verify required and optional flags
	flags := metricsSearchCmd.Flags()
	if flags.Lookup("query") == nil {
		t.Error("Missing --query flag")
	}
	if flags.Lookup("from") == nil {
		t.Error("Missing --from flag")
	}
	if flags.Lookup("to") == nil {
		t.Error("Missing --to flag")
	}
}

func TestMetricsCmd_Subcommands(t *testing.T) {
	expectedCommands := []string{"query", "search", "list", "metadata", "submit", "tags"}

	commands := metricsCmd.Commands()
	commandMap := make(map[string]bool)
	for _, cmd := range commands {
		commandMap[cmd.Name()] = true
	}

	for _, expected := range expectedCommands {
		if !commandMap[expected] {
			t.Errorf("Missing subcommand: %s", expected)
		}
	}
}

func TestRunMetricsList(t *testing.T) {
	cleanup := setupMetricsTestClient(t)
	defer cleanup()

	tests := []struct {
		name           string
		filter         string
		wantErr        bool
		wantErrContains string
	}{
		{
			name:           "no filter",
			filter:         "",
			wantErr:        true,
			wantErrContains: "mock client",
		},
		{
			name:           "with filter",
			filter:         "system.*",
			wantErr:        true,
			wantErrContains: "mock client",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filterPattern = tt.filter

			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			err := runMetricsList(metricsListCmd, []string{})

			if (err != nil) != tt.wantErr {
				t.Errorf("runMetricsList() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErrContains != "" && err != nil && !strings.Contains(err.Error(), tt.wantErrContains) {
				t.Errorf("runMetricsList() error = %v, want error containing %q", err, tt.wantErrContains)
			}
		})
	}
}

func TestRunMetricsMetadataGet(t *testing.T) {
	cleanup := setupMetricsTestClient(t)
	defer cleanup()

	tests := []struct {
		name           string
		metricName     string
		wantErr        bool
		wantErrContains string
	}{
		{
			name:           "fails on client creation",
			metricName:     "system.cpu.user",
			wantErr:        true,
			wantErrContains: "mock client",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			err := runMetricsMetadataGet(metricsMetadataGetCmd, []string{tt.metricName})

			if (err != nil) != tt.wantErr {
				t.Errorf("runMetricsMetadataGet() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErrContains != "" && err != nil && !strings.Contains(err.Error(), tt.wantErrContains) {
				t.Errorf("runMetricsMetadataGet() error = %v, want error containing %q", err, tt.wantErrContains)
			}
		})
	}
}

func TestRunMetricsMetadataUpdate(t *testing.T) {
	cleanup := setupMetricsTestClient(t)
	defer cleanup()

	tests := []struct {
		name           string
		metricName     string
		description    string
		unit           string
		metricType     string
		wantErr        bool
		wantErrContains string
	}{
		{
			name:           "update description",
			metricName:     "system.cpu.user",
			description:    "CPU user time",
			unit:           "",
			metricType:     "",
			wantErr:        true,
			wantErrContains: "mock client",
		},
		{
			name:           "update multiple fields",
			metricName:     "system.cpu.user",
			description:    "CPU user time",
			unit:           "percent",
			metricType:     "gauge",
			wantErr:        true,
			wantErrContains: "mock client",
		},
		{
			name:           "no fields specified hits client error first",
			metricName:     "system.cpu.user",
			description:    "",
			unit:           "",
			metricType:     "",
			wantErr:        true,
			wantErrContains: "mock client",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadataDescription = tt.description
			metadataUnit = tt.unit
			metadataType = tt.metricType
			metadataPerUnit = ""
			metadataShortName = ""

			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			err := runMetricsMetadataUpdate(metricsMetadataUpdateCmd, []string{tt.metricName})

			if (err != nil) != tt.wantErr {
				t.Errorf("runMetricsMetadataUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErrContains != "" && err != nil && !strings.Contains(err.Error(), tt.wantErrContains) {
				t.Errorf("runMetricsMetadataUpdate() error = %v, want error containing %q", err, tt.wantErrContains)
			}
		})
	}
}

func TestRunMetricsSubmit(t *testing.T) {
	cleanup := setupMetricsTestClient(t)
	defer cleanup()

	tests := []struct {
		name       string
		metricName string
		value      float64
		timestamp  string
		tags       string
		metricType string
		wantErr    bool
	}{
		{
			name:       "submit gauge",
			metricName: "custom.metric",
			value:      123.45,
			timestamp:  "now",
			tags:       "env:prod,team:backend",
			metricType: "gauge",
			wantErr:    true, // Mock client error
		},
		{
			name:       "submit count",
			metricName: "custom.count",
			value:      100,
			timestamp:  "now",
			tags:       "",
			metricType: "count",
			wantErr:    true,
		},
		{
			name:       "invalid metric type",
			metricName: "custom.metric",
			value:      123,
			timestamp:  "now",
			tags:       "",
			metricType: "invalid",
			wantErr:    true, // Will error on invalid type validation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			submitName = tt.metricName
			submitValue = tt.value
			submitTimestamp = tt.timestamp
			submitTags = tt.tags
			submitType = tt.metricType
			submitInterval = 0

			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			err := runMetricsSubmit(metricsSubmitCmd, []string{})

			if (err != nil) != tt.wantErr {
				t.Errorf("runMetricsSubmit() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check for specific error on invalid type
			// Note: Invalid type validation happens after client creation,
			// but with mock client, client creation fails first
			if tt.metricType == "invalid" && err != nil {
				// Accept either "invalid metric type" or "mock client" error
				if !strings.Contains(err.Error(), "invalid metric type") && !strings.Contains(err.Error(), "mock client") {
					t.Errorf("runMetricsSubmit() error = %v, want 'invalid metric type' or 'mock client' error", err)
				}
			}
		})
	}
}

func TestRunMetricsTagsList(t *testing.T) {
	cleanup := setupMetricsTestClient(t)
	defer cleanup()

	tests := []struct {
		name       string
		metricName string
		wantErr    bool
	}{
		{
			name:       "not supported",
			metricName: "system.cpu.user",
			wantErr:    true, // Should return "not supported" error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			err := runMetricsTagsList(metricsTagsListCmd, []string{tt.metricName})

			if (err != nil) != tt.wantErr {
				t.Errorf("runMetricsTagsList() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && !strings.Contains(err.Error(), "not supported") {
				t.Errorf("runMetricsTagsList() error = %v, want 'not supported' error", err)
			}
		})
	}
}

func TestParseTimeParam(t *testing.T) {
	tests := []struct {
		name    string
		timeStr string
		wantErr bool
	}{
		{
			name:    "now keyword",
			timeStr: "now",
			wantErr: false,
		},
		{
			name:    "NOW uppercase",
			timeStr: "NOW",
			wantErr: false,
		},
		{
			name:    "relative hours",
			timeStr: "1h",
			wantErr: false,
		},
		{
			name:    "relative hours multiple digits",
			timeStr: "24h",
			wantErr: false,
		},
		{
			name:    "relative minutes",
			timeStr: "30m",
			wantErr: false,
		},
		{
			name:    "relative days",
			timeStr: "7d",
			wantErr: false,
		},
		{
			name:    "relative weeks",
			timeStr: "2w",
			wantErr: false,
		},
		{
			name:    "unix timestamp",
			timeStr: "1640000000",
			wantErr: false,
		},
		{
			name:    "invalid format",
			timeStr: "invalid",
			wantErr: true,
		},
		{
			name:    "empty string",
			timeStr: "",
			wantErr: true,
		},
		{
			name:    "single character",
			timeStr: "h",
			wantErr: true,
		},
		{
			name:    "invalid unit",
			timeStr: "5x",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := util.ParseTimeParam(tt.timeStr)

			if (err != nil) != tt.wantErr {
				t.Errorf("util.ParseTimeParam(%q) error = %v, wantErr %v", tt.timeStr, err, tt.wantErr)
			}

			// Validate result for successful cases
			if err == nil {
				if result.IsZero() {
					t.Errorf("util.ParseTimeParam(%q) returned zero time", tt.timeStr)
				}
			}
		})
	}
}

func TestParseTimeParam_RelativeTime(t *testing.T) {
	tests := []struct {
		name        string
		timeStr     string
		expectPast  bool
		description string
	}{
		{
			name:        "1 hour ago",
			timeStr:     "1h",
			expectPast:  true,
			description: "should be ~1 hour before now",
		},
		{
			name:        "30 minutes ago",
			timeStr:     "30m",
			expectPast:  true,
			description: "should be ~30 minutes before now",
		},
		{
			name:        "7 days ago",
			timeStr:     "7d",
			expectPast:  true,
			description: "should be ~7 days before now",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := util.ParseTimeParam(tt.timeStr)
			if err != nil {
				t.Fatalf("util.ParseTimeParam(%q) unexpected error: %v", tt.timeStr, err)
			}

			now := time.Now()
			if tt.expectPast && result.After(now) {
				t.Errorf("util.ParseTimeParam(%q) = %v, expected time in the past", tt.timeStr, result)
			}
		})
	}
}

func TestParseTimeParam_NowKeyword(t *testing.T) {
	result, err := util.ParseTimeParam("now")
	if err != nil {
		t.Fatalf("util.ParseTimeParam(\"now\") unexpected error: %v", err)
	}

	now := time.Now()
	diff := now.Sub(result)

	// Should be very close to current time (within 1 second)
	if diff > time.Second || diff < -time.Second {
		t.Errorf("util.ParseTimeParam(\"now\") = %v, too far from current time %v (diff: %v)", result, now, diff)
	}
}
