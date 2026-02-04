// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package util

import (
	"testing"
	"time"
)

func TestParseTimeParam(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		input     string
		wantError bool
		checkFunc func(time.Time) bool
	}{
		{
			name:      "now keyword",
			input:     "now",
			wantError: false,
			checkFunc: func(t time.Time) bool {
				// Should be within 1 second of now
				return time.Since(t) < time.Second
			},
		},
		{
			name:      "1 hour ago",
			input:     "1h",
			wantError: false,
			checkFunc: func(t time.Time) bool {
				expected := now.Add(-1 * time.Hour)
				diff := expected.Sub(t).Abs()
				return diff < time.Second
			},
		},
		{
			name:      "30 minutes ago",
			input:     "30m",
			wantError: false,
			checkFunc: func(t time.Time) bool {
				expected := now.Add(-30 * time.Minute)
				diff := expected.Sub(t).Abs()
				return diff < time.Second
			},
		},
		{
			name:      "7 days ago",
			input:     "7d",
			wantError: false,
			checkFunc: func(t time.Time) bool {
				expected := now.Add(-7 * 24 * time.Hour)
				diff := expected.Sub(t).Abs()
				return diff < time.Second
			},
		},
		{
			name:      "45 seconds ago",
			input:     "45s",
			wantError: false,
			checkFunc: func(t time.Time) bool {
				expected := now.Add(-45 * time.Second)
				diff := expected.Sub(t).Abs()
				return diff < time.Second
			},
		},
		{
			name:      "Unix timestamp",
			input:     "1704067200",
			wantError: false,
			checkFunc: func(t time.Time) bool {
				return t.Unix() == 1704067200
			},
		},
		{
			name:      "ISO 8601 format",
			input:     "2024-01-01T00:00:00Z",
			wantError: false,
			checkFunc: func(t time.Time) bool {
				expected, _ := time.Parse(time.RFC3339, "2024-01-01T00:00:00Z")
				return t.Equal(expected)
			},
		},
		{
			name:      "invalid format",
			input:     "invalid",
			wantError: true,
		},
		{
			name:      "invalid relative time",
			input:     "10x",
			wantError: true,
		},
		{
			name:      "negative value",
			input:     "-5h",
			wantError: true,
		},
		{
			name:      "empty string",
			input:     "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTimeParam(tt.input)

			if tt.wantError {
				if err == nil {
					t.Errorf("ParseTimeParam(%q) expected error but got none", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseTimeParam(%q) unexpected error: %v", tt.input, err)
				return
			}

			if tt.checkFunc != nil && !tt.checkFunc(result) {
				t.Errorf("ParseTimeParam(%q) = %v, validation failed", tt.input, result)
			}
		})
	}
}

func TestParseTimeToUnix(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
		wantValue int64
	}{
		{
			name:      "Unix timestamp",
			input:     "1704067200",
			wantError: false,
			wantValue: 1704067200,
		},
		{
			name:      "ISO date",
			input:     "2024-01-01T00:00:00Z",
			wantError: false,
			wantValue: 1704067200,
		},
		{
			name:      "invalid format",
			input:     "invalid",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTimeToUnix(tt.input)

			if tt.wantError {
				if err == nil {
					t.Errorf("ParseTimeToUnix(%q) expected error but got none", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseTimeToUnix(%q) unexpected error: %v", tt.input, err)
				return
			}

			if result != tt.wantValue {
				t.Errorf("ParseTimeToUnix(%q) = %d, want %d", tt.input, result, tt.wantValue)
			}
		})
	}
}

func TestParseTimeToUnixMilli(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
		wantValue int64
	}{
		{
			name:      "Unix timestamp",
			input:     "1704067200",
			wantError: false,
			wantValue: 1704067200000,
		},
		{
			name:      "ISO date",
			input:     "2024-01-01T00:00:00Z",
			wantError: false,
			wantValue: 1704067200000,
		},
		{
			name:      "invalid format",
			input:     "invalid",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTimeToUnixMilli(tt.input)

			if tt.wantError {
				if err == nil {
					t.Errorf("ParseTimeToUnixMilli(%q) expected error but got none", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseTimeToUnixMilli(%q) unexpected error: %v", tt.input, err)
				return
			}

			if result != tt.wantValue {
				t.Errorf("ParseTimeToUnixMilli(%q) = %d, want %d", tt.input, result, tt.wantValue)
			}
		})
	}
}

func TestParseTimeParam_RelativeTimeEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"large hour value", "100h"},
		{"large day value", "365d"},
		{"single minute", "1m"},
		{"single second", "1s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTimeParam(tt.input)
			if err != nil {
				t.Errorf("ParseTimeParam(%q) unexpected error: %v", tt.input, err)
			}
			if result.IsZero() {
				t.Errorf("ParseTimeParam(%q) returned zero time", tt.input)
			}
			if result.After(time.Now()) {
				t.Errorf("ParseTimeParam(%q) returned future time: %v", tt.input, result)
			}
		})
	}
}
