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

	"github.com/DataDog/pup/pkg/client"
	"github.com/DataDog/pup/pkg/config"
)

func TestLogsCmd(t *testing.T) {
	if logsCmd == nil {
		t.Fatal("logsCmd is nil")
	}

	if logsCmd.Use != "logs" {
		t.Errorf("Use = %s, want logs", logsCmd.Use)
	}
}

func TestParseTimeString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(int64) bool
	}{
		{
			name:  "relative time - 1 hour",
			input: "1h",
			check: func(ts int64) bool {
				// Should be roughly 1 hour ago in milliseconds
				// Timestamps should be 13 digits for milliseconds since epoch
				return ts > 1000000000000 && ts < 9999999999999
			},
		},
		{
			name:  "relative time - 7 days",
			input: "7d",
			check: func(ts int64) bool {
				// Should be roughly 7 days ago in milliseconds
				return ts > 1000000000000 && ts < 9999999999999
			},
		},
		{
			name:  "now",
			input: "now",
			check: func(ts int64) bool {
				// Should be current time in milliseconds (13 digits)
				return ts > 1000000000000 && ts < 9999999999999
			},
		},
		{
			name:    "invalid format",
			input:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTimeString(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseTimeString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil && !tt.check(got) {
				t.Errorf("parseTimeString() = %d, validation failed", got)
			}
		})
	}
}

func TestParseComputeString(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		wantAggregation string
		wantMetric      string
		wantErr         bool
		errContains     string
	}{
		{
			name:            "count - no metric",
			input:           "count",
			wantAggregation: "count",
			wantMetric:      "",
			wantErr:         false,
		},
		{
			name:            "count - uppercase",
			input:           "COUNT",
			wantAggregation: "count",
			wantMetric:      "",
			wantErr:         false,
		},
		{
			name:            "avg with metric",
			input:           "avg(@duration)",
			wantAggregation: "avg",
			wantMetric:      "@duration",
			wantErr:         false,
		},
		{
			name:            "sum with metric",
			input:           "sum(@bytes)",
			wantAggregation: "sum",
			wantMetric:      "@bytes",
			wantErr:         false,
		},
		{
			name:            "min with metric",
			input:           "min(@response_time)",
			wantAggregation: "min",
			wantMetric:      "@response_time",
			wantErr:         false,
		},
		{
			name:            "max with metric",
			input:           "max(@duration)",
			wantAggregation: "max",
			wantMetric:      "@duration",
			wantErr:         false,
		},
		{
			name:            "cardinality with metric",
			input:           "cardinality(@user.id)",
			wantAggregation: "cardinality",
			wantMetric:      "@user.id",
			wantErr:         false,
		},
		{
			name:            "percentile with metric and parameter - converts to pc99",
			input:           "percentile(@duration, 99)",
			wantAggregation: "pc99",
			wantMetric:      "@duration",
			wantErr:         false,
		},
		{
			name:            "percentile pc95",
			input:           "percentile(@latency, 95)",
			wantAggregation: "pc95",
			wantMetric:      "@latency",
			wantErr:         false,
		},
		{
			name:            "percentile pc50 (median)",
			input:           "percentile(@response_time, 50)",
			wantAggregation: "pc50",
			wantMetric:      "@response_time",
			wantErr:         false,
		},
		{
			name:        "percentile without value",
			input:       "percentile(@duration)",
			wantErr:     true,
			errContains: "percentile requires a percentile value",
		},
		{
			name:            "median with metric",
			input:           "median(@latency)",
			wantAggregation: "median",
			wantMetric:      "@latency",
			wantErr:         false,
		},
		{
			name:            "metric with dots and underscores",
			input:           "avg(@http.response_time)",
			wantAggregation: "avg",
			wantMetric:      "@http.response_time",
			wantErr:         false,
		},
		{
			name:            "whitespace handling",
			input:           "  avg(@duration)  ",
			wantAggregation: "avg",
			wantMetric:      "@duration",
			wantErr:         false,
		},
		{
			name:        "invalid - unknown function",
			input:       "invalid(@duration)",
			wantErr:     true,
			errContains: "unknown aggregation function",
		},
		{
			name:        "invalid - malformed",
			input:       "avg(@duration",
			wantErr:     true,
			errContains: "invalid compute format",
		},
		{
			name:        "invalid - no function name",
			input:       "(@duration)",
			wantErr:     true,
			errContains: "invalid compute format",
		},
		{
			name:        "invalid - empty string",
			input:       "",
			wantErr:     true,
			errContains: "invalid compute format",
		},
		{
			name:            "case insensitive function",
			input:           "AVG(@duration)",
			wantAggregation: "avg",
			wantMetric:      "@duration",
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAgg, gotMetric, err := parseComputeString(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseComputeString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("parseComputeString() error = %v, should contain %q", err, tt.errContains)
				}
				return
			}

			if gotAgg != tt.wantAggregation {
				t.Errorf("parseComputeString() aggregation = %q, want %q", gotAgg, tt.wantAggregation)
			}

			if gotMetric != tt.wantMetric {
				t.Errorf("parseComputeString() metric = %q, want %q", gotMetric, tt.wantMetric)
			}
		})
	}
}

// Helper function to setup logs test client
func setupLogsTestClient(t *testing.T) func() {
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

func TestRunLogsSearch(t *testing.T) {
	cleanup := setupLogsTestClient(t)
	defer cleanup()

	tests := []struct {
		name    string
		query   string
		from    string
		to      string
		wantErr bool
	}{
		{
			name:    "valid query",
			query:   "status:error",
			from:    "1h",
			to:      "now",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logsQuery = tt.query
			logsFrom = tt.from
			logsTo = tt.to

			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			err := runLogsSearch(logsSearchCmd, []string{})

			if (err != nil) != tt.wantErr {
				t.Errorf("runLogsSearch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunLogsAggregate(t *testing.T) {
	cleanup := setupLogsTestClient(t)
	defer cleanup()

	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "valid aggregate query",
			query:   "status:error",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logsQuery = tt.query

			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			err := runLogsAggregate(logsAggregateCmd, []string{})

			if (err != nil) != tt.wantErr {
				t.Errorf("runLogsAggregate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunLogsList(t *testing.T) {
	cleanup := setupLogsTestClient(t)
	defer cleanup()

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "list logs",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			err := runLogsList(logsListCmd, []string{})

			if (err != nil) != tt.wantErr {
				t.Errorf("runLogsList() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunLogsQuery(t *testing.T) {
	cleanup := setupLogsTestClient(t)
	defer cleanup()

	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "query logs",
			query:   "status:error",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logsQuery = tt.query

			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			err := runLogsQuery(logsQueryCmd, []string{})

			if (err != nil) != tt.wantErr {
				t.Errorf("runLogsQuery() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunLogsCustomDestinationsList(t *testing.T) {
	cleanup := setupLogsTestClient(t)
	defer cleanup()

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "list custom destinations",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			err := runLogsCustomDestinationsList(logsCustomDestinationsListCmd, []string{})

			if (err != nil) != tt.wantErr {
				t.Errorf("runLogsCustomDestinationsList() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunLogsCustomDestinationsGet(t *testing.T) {
	cleanup := setupLogsTestClient(t)
	defer cleanup()

	tests := []struct {
		name          string
		destinationID string
		wantErr       bool
	}{
		{
			name:          "get custom destination",
			destinationID: "dest-123",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			err := runLogsCustomDestinationsGet(logsCustomDestinationsGetCmd, []string{tt.destinationID})

			if (err != nil) != tt.wantErr {
				t.Errorf("runLogsCustomDestinationsGet() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunLogsArchivesList(t *testing.T) {
	cleanup := setupLogsTestClient(t)
	defer cleanup()

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "list archives",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			err := runLogsArchivesList(logsArchivesListCmd, []string{})

			if (err != nil) != tt.wantErr {
				t.Errorf("runLogsArchivesList() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunLogsArchivesGet(t *testing.T) {
	cleanup := setupLogsTestClient(t)
	defer cleanup()

	tests := []struct {
		name       string
		archiveID  string
		wantErr    bool
	}{
		{
			name:      "get archive",
			archiveID: "archive-123",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			err := runLogsArchivesGet(logsArchivesGetCmd, []string{tt.archiveID})

			if (err != nil) != tt.wantErr {
				t.Errorf("runLogsArchivesGet() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunLogsMetricsList(t *testing.T) {
	cleanup := setupLogsTestClient(t)
	defer cleanup()

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "list metrics",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			err := runLogsMetricsList(logsMetricsListCmd, []string{})

			if (err != nil) != tt.wantErr {
				t.Errorf("runLogsMetricsList() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunLogsMetricsGet(t *testing.T) {
	cleanup := setupLogsTestClient(t)
	defer cleanup()

	tests := []struct {
		name      string
		metricID  string
		wantErr   bool
	}{
		{
			name:     "get metric",
			metricID: "metric-123",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			err := runLogsMetricsGet(logsMetricsGetCmd, []string{tt.metricID})

			if (err != nil) != tt.wantErr {
				t.Errorf("runLogsMetricsGet() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunLogsArchivesDelete(t *testing.T) {
	cleanup := setupLogsTestClient(t)
	defer cleanup()

	tests := []struct {
		name       string
		archiveID  string
		wantErr    bool
	}{
		{
			name:      "delete archive",
			archiveID: "archive-123",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			err := runLogsArchivesDelete(logsArchivesDeleteCmd, []string{tt.archiveID})

			if (err != nil) != tt.wantErr {
				t.Errorf("runLogsArchivesDelete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunLogsMetricsDelete(t *testing.T) {
	cleanup := setupLogsTestClient(t)
	defer cleanup()

	tests := []struct {
		name      string
		metricID  string
		wantErr   bool
	}{
		{
			name:     "delete metric",
			metricID: "metric-123",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			err := runLogsMetricsDelete(logsMetricsDeleteCmd, []string{tt.metricID})

			if (err != nil) != tt.wantErr {
				t.Errorf("runLogsMetricsDelete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunLogsRestrictionQueriesList(t *testing.T) {
	cleanup := setupLogsTestClient(t)
	defer cleanup()

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "list restriction queries",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			err := runLogsRestrictionQueriesList(logsRestrictionQueriesListCmd, []string{})

			if (err != nil) != tt.wantErr {
				t.Errorf("runLogsRestrictionQueriesList() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunLogsRestrictionQueriesGet(t *testing.T) {
	cleanup := setupLogsTestClient(t)
	defer cleanup()

	tests := []struct {
		name    string
		queryID string
		wantErr bool
	}{
		{
			name:    "get restriction query",
			queryID: "query-123",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			err := runLogsRestrictionQueriesGet(logsRestrictionQueriesGetCmd, []string{tt.queryID})

			if (err != nil) != tt.wantErr {
				t.Errorf("runLogsRestrictionQueriesGet() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
