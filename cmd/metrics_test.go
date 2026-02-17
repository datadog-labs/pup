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
	origAPIKeyFactory := apiKeyClientFactory

	cfg = &config.Config{
		Site:        "datadoghq.com",
		APIKey:      "test-api-key-12345678",
		AppKey:      "test-app-key-12345678",
		AutoApprove: false,
	}

	mockErr := fmt.Errorf("mock client: no real API connection in tests")
	clientFactory = func(c *config.Config) (*client.Client, error) {
		return nil, mockErr
	}
	apiKeyClientFactory = func(c *config.Config) (*client.Client, error) {
		return nil, mockErr
	}

	ddClient = nil

	return func() {
		ddClient = origClient
		cfg = origCfg
		clientFactory = origFactory
		apiKeyClientFactory = origAPIKeyFactory
	}
}

func TestRunMetricsSearch(t *testing.T) {
	cleanup := setupMetricsTestClient(t)
	defer cleanup()

	tests := []struct {
		name            string
		query           string
		from            string
		to              string
		wantErr         bool
		wantErrContains string
	}{
		{
			name:            "fails on client creation",
			query:           "avg:system.cpu.user{*}",
			from:            "1h",
			to:              "now",
			wantErr:         true,
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
		name            string
		query           string
		from            string
		to              string
		wantErr         bool
		wantErrContains string
	}{
		{
			name:            "fails on client creation",
			query:           "avg:system.cpu.user{*}",
			from:            "1h",
			to:              "now",
			wantErr:         true,
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
		name            string
		filter          string
		tagFilterValue  string
		wantErr         bool
		wantErrContains string
	}{
		{
			name:            "no filter",
			filter:          "",
			tagFilterValue:  "",
			wantErr:         true,
			wantErrContains: "mock client",
		},
		{
			name:            "with name filter",
			filter:          "system.*",
			tagFilterValue:  "",
			wantErr:         true,
			wantErrContains: "mock client",
		},
		{
			name:            "with tag filter",
			filter:          "",
			tagFilterValue:  "env:prod",
			wantErr:         true,
			wantErrContains: "mock client",
		},
		{
			name:            "with both filters",
			filter:          "system.*",
			tagFilterValue:  "env:prod",
			wantErr:         true,
			wantErrContains: "mock client",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filterPattern = tt.filter
			tagFilter = tt.tagFilterValue

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

func TestMatchMetricName(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		metric  string
		want    bool
	}{
		// Empty pattern
		{
			name:    "empty pattern matches all",
			pattern: "",
			metric:  "system.cpu.user",
			want:    true,
		},
		// Exact matches
		{
			name:    "exact match",
			pattern: "system.cpu.user",
			metric:  "system.cpu.user",
			want:    true,
		},
		{
			name:    "exact mismatch",
			pattern: "system.cpu.user",
			metric:  "system.cpu.system",
			want:    false,
		},
		// Prefix wildcards
		{
			name:    "prefix wildcard matches",
			pattern: "system.*",
			metric:  "system.cpu.user",
			want:    true,
		},
		{
			name:    "prefix wildcard matches all system metrics",
			pattern: "system.*",
			metric:  "system.mem.used",
			want:    true,
		},
		{
			name:    "prefix wildcard no match",
			pattern: "system.*",
			metric:  "custom.cpu.user",
			want:    false,
		},
		// Suffix wildcards
		{
			name:    "suffix wildcard matches",
			pattern: "*.cpu.user",
			metric:  "system.cpu.user",
			want:    true,
		},
		{
			name:    "suffix wildcard matches different prefix",
			pattern: "*.cpu.user",
			metric:  "custom.cpu.user",
			want:    true,
		},
		{
			name:    "suffix wildcard no match",
			pattern: "*.cpu.user",
			metric:  "system.mem.used",
			want:    false,
		},
		// Middle wildcards
		{
			name:    "middle wildcard matches",
			pattern: "system.*.user",
			metric:  "system.cpu.user",
			want:    true,
		},
		{
			name:    "middle wildcard matches multiple segments",
			pattern: "system.*.user",
			metric:  "system.foo.bar.user",
			want:    true,
		},
		{
			name:    "middle wildcard no match",
			pattern: "system.*.user",
			metric:  "system.cpu.system",
			want:    false,
		},
		// Contains patterns
		{
			name:    "contains pattern matches",
			pattern: "*cpu*",
			metric:  "system.cpu.user",
			want:    true,
		},
		{
			name:    "contains pattern matches middle",
			pattern: "*request*",
			metric:  "app.request.count",
			want:    true,
		},
		{
			name:    "contains pattern no match",
			pattern: "*request*",
			metric:  "system.cpu.user",
			want:    false,
		},
		// Question mark wildcards
		{
			name:    "question mark matches single char",
			pattern: "system.cpu.?ser",
			metric:  "system.cpu.user",
			want:    true,
		},
		{
			name:    "question mark matches any char",
			pattern: "system.cpu.?ser",
			metric:  "system.cpu.aser",
			want:    true,
		},
		{
			name:    "question mark no match multiple chars",
			pattern: "system.cpu.?ser",
			metric:  "system.cpu.abser",
			want:    false,
		},
		{
			name:    "question mark no match too short",
			pattern: "system.cpu.?ser",
			metric:  "system.cpu.ser",
			want:    false,
		},
		// Complex patterns
		{
			name:    "multiple wildcards",
			pattern: "*.cpu.*",
			metric:  "system.cpu.user",
			want:    true,
		},
		{
			name:    "multiple wildcards different segments",
			pattern: "*.cpu.*",
			metric:  "custom.app.cpu.load",
			want:    true,
		},
		{
			name:    "multiple wildcards no match",
			pattern: "*.cpu.*",
			metric:  "system.mem.used",
			want:    false,
		},
		// Edge cases
		{
			name:    "only wildcard matches all",
			pattern: "*",
			metric:  "any.metric.name",
			want:    true,
		},
		{
			name:    "trailing wildcard",
			pattern: "system.cpu.*",
			metric:  "system.cpu.user",
			want:    true,
		},
		{
			name:    "leading wildcard",
			pattern: "*.user",
			metric:  "system.cpu.user",
			want:    true,
		},
		// Real-world examples from the issue
		{
			name:    "kafka_index prefix",
			pattern: "kafka_index*",
			metric:  "kafka_index.fetch.latency",
			want:    true,
		},
		{
			name:    "kafka_index wildcard",
			pattern: "*kafka_index*",
			metric:  "my.kafka_index.metric",
			want:    true,
		},
		{
			name:    "system.cpu exact",
			pattern: "system.cpu",
			metric:  "system.cpu",
			want:    true,
		},
		{
			name:    "system.cpu prefix",
			pattern: "system.cpu*",
			metric:  "system.cpu.user",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchMetricName(tt.pattern, tt.metric)
			if got != tt.want {
				t.Errorf("matchMetricName(%q, %q) = %v, want %v",
					tt.pattern, tt.metric, got, tt.want)
			}
		})
	}
}

func TestRunMetricsMetadataGet(t *testing.T) {
	cleanup := setupMetricsTestClient(t)
	defer cleanup()

	tests := []struct {
		name            string
		metricName      string
		wantErr         bool
		wantErrContains string
	}{
		{
			name:            "fails on client creation",
			metricName:      "system.cpu.user",
			wantErr:         true,
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
		name            string
		metricName      string
		description     string
		unit            string
		metricType      string
		wantErr         bool
		wantErrContains string
	}{
		{
			name:            "update description",
			metricName:      "system.cpu.user",
			description:     "CPU user time",
			unit:            "",
			metricType:      "",
			wantErr:         true,
			wantErrContains: "mock client",
		},
		{
			name:            "update multiple fields",
			metricName:      "system.cpu.user",
			description:     "CPU user time",
			unit:            "percent",
			metricType:      "gauge",
			wantErr:         true,
			wantErrContains: "mock client",
		},
		{
			name:            "no fields specified hits client error first",
			metricName:      "system.cpu.user",
			description:     "",
			unit:            "",
			metricType:      "",
			wantErr:         true,
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
		name            string
		metricName      string
		value           float64
		timestamp       string
		tags            string
		metricType      string
		wantErr         bool
		wantErrContains string
	}{
		{
			name:            "submit gauge",
			metricName:      "custom.metric",
			value:           123.45,
			timestamp:       "now",
			tags:            "env:prod,team:backend",
			metricType:      "gauge",
			wantErr:         true, // Mock client error
			wantErrContains: "mock client",
		},
		{
			name:            "submit count",
			metricName:      "custom.count",
			value:           100,
			timestamp:       "now",
			tags:            "",
			metricType:      "count",
			wantErr:         true,
			wantErrContains: "mock client",
		},
		{
			name:            "invalid metric type",
			metricName:      "custom.metric",
			value:           123,
			timestamp:       "now",
			tags:            "",
			metricType:      "invalid",
			wantErr:         true,
			wantErrContains: "mock client",
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

			if tt.wantErrContains != "" && err != nil && !strings.Contains(err.Error(), tt.wantErrContains) {
				t.Errorf("runMetricsSubmit() error = %v, want error containing %q", err, tt.wantErrContains)
			}
		})
	}
}

func TestRunMetricsSubmit_RequiresAPIKey(t *testing.T) {
	origClient := ddClient
	origCfg := cfg
	origFactory := apiKeyClientFactory
	defer func() {
		ddClient = origClient
		cfg = origCfg
		apiKeyClientFactory = origFactory
	}()

	// Set config with NO API key to trigger the pre-emptive check
	cfg = &config.Config{
		Site:   "datadoghq.com",
		APIKey: "",
		AppKey: "",
	}
	ddClient = nil

	submitName = "custom.metric"
	submitValue = 42.0
	submitTimestamp = "now"
	submitTags = ""
	submitType = "gauge"
	submitInterval = 0

	var buf bytes.Buffer
	outputWriter = &buf
	defer func() { outputWriter = os.Stdout }()

	err := runMetricsSubmit(metricsSubmitCmd, []string{})
	if err == nil {
		t.Fatal("expected error when DD_API_KEY is not set")
	}

	if !strings.Contains(err.Error(), "DD_API_KEY") {
		t.Errorf("error should mention DD_API_KEY, got: %v", err)
	}
	if !strings.Contains(err.Error(), "metrics submit requires") {
		t.Errorf("error should explain the requirement, got: %v", err)
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

// TestV2TimeseriesTimestampConversion verifies that ParseTimeToUnixMilli
// produces second-aligned millisecond timestamps suitable for Datadog APIs.
func TestV2TimeseriesTimestampConversion(t *testing.T) {
	tests := []struct {
		name    string
		timeStr string
	}{
		{"now keyword", "now"},
		{"relative 1 hour", "1h"},
		{"relative 30 minutes", "30m"},
		{"relative 7 days", "7d"},
		{"unix timestamp", "1700000000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msTimestamp, err := util.ParseTimeToUnixMilli(tt.timeStr)
			if err != nil {
				t.Fatalf("ParseTimeToUnixMilli(%q) unexpected error: %v", tt.timeStr, err)
			}

			if msTimestamp <= 0 {
				t.Errorf("expected positive millisecond timestamp, got %d", msTimestamp)
			}

			// Must be on a second boundary (divisible by 1000)
			if msTimestamp%1000 != 0 {
				t.Errorf("timestamp %d is not on a second boundary (remainder: %d)", msTimestamp, msTimestamp%1000)
			}

			// Verify round-trip: converting back should produce a valid second
			roundTripped := time.Unix(msTimestamp/1000, 0)
			if roundTripped.Unix() != msTimestamp/1000 {
				t.Errorf("round-trip failed: expected Unix=%d, got Unix=%d",
					msTimestamp/1000, roundTripped.Unix())
			}
		})
	}
}

// TestV2TimestampUnixMilliVsUnixTimes1000 demonstrates the difference between
// UnixMilli() and Unix()*1000 when a time.Time has sub-second precision.
// parseTimeParam("now") calls time.Now() which includes nanosecond precision,
// and relative times are computed via time.Duration arithmetic that also
// preserves nanosecond precision.
func TestV2TimestampUnixMilliVsUnixTimes1000(t *testing.T) {
	// Construct a time with known sub-second precision
	// 2024-01-15 12:00:00.123456789 UTC
	ts := time.Date(2024, 1, 15, 12, 0, 0, 123456789, time.UTC)

	unixMilli := ts.UnixMilli()       // Includes millisecond component: ...123
	unixTimes1000 := ts.Unix() * 1000 // Truncated to second boundary: ...000

	// UnixMilli includes the sub-second milliseconds
	if unixMilli%1000 == 0 {
		t.Error("expected UnixMilli() to have sub-second component for a time with nanoseconds")
	}

	// Unix()*1000 is always on a second boundary
	if unixTimes1000%1000 != 0 {
		t.Errorf("expected Unix()*1000 to be on second boundary, got remainder %d", unixTimes1000%1000)
	}

	// The difference should be exactly the millisecond component (123ms)
	diff := unixMilli - unixTimes1000
	if diff != 123 {
		t.Errorf("expected difference of 123ms, got %d", diff)
	}

	// Both should represent approximately the same point in time
	// (within 1 second)
	if diff < 0 || diff >= 1000 {
		t.Errorf("timestamps diverged by more than 1 second: diff=%dms", diff)
	}
}
