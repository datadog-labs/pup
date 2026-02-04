// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"testing"
)

func TestParseInt64(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int64
	}{
		{
			name:  "valid positive number",
			input: "12345",
			want:  12345,
		},
		{
			name:  "valid negative number",
			input: "-54321",
			want:  -54321,
		},
		{
			name:  "zero",
			input: "0",
			want:  0,
		},
		{
			name:  "large number",
			input: "9223372036854775807",
			want:  9223372036854775807,
		},
		{
			name:  "invalid input - string",
			input: "abc",
			want:  0,
		},
		{
			name:  "invalid input - empty",
			input: "",
			want:  0,
		},
		{
			name:  "invalid input - decimal",
			input: "123.45",
			want:  0,
		},
		{
			name:  "invalid input - mixed",
			input: "123abc",
			want:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseInt64(tt.input)
			if got != tt.want {
				t.Errorf("parseInt64(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseInt64_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int64
	}{
		{
			name:  "overflow - returns 0",
			input: "9223372036854775808", // Max int64 + 1
			want:  0,
		},
		{
			name:  "underflow - returns 0",
			input: "-9223372036854775809", // Min int64 - 1
			want:  0,
		},
		{
			name:  "whitespace - returns 0",
			input: "  123  ",
			want:  0,
		},
		{
			name:  "hex format - returns 0",
			input: "0x123",
			want:  0,
		},
		{
			name:  "octal format - returns 0",
			input: "0123",
			want:  123, // This might parse as decimal 123
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseInt64(tt.input)
			if got != tt.want {
				t.Errorf("parseInt64(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}
