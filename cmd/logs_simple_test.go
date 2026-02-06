// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"bytes"
	"fmt"
	"os"
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
