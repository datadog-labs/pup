// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/pup/pkg/config"
)

func TestRootCmd_SilenceUsage(t *testing.T) {
	// Verify SilenceUsage is set on root command
	// This applies to all subcommands automatically
	if !rootCmd.SilenceUsage {
		t.Error("rootCmd.SilenceUsage should be true to prevent help on errors globally")
	}
}

// mockHTTPResponse implements the httpResponse interface for testing
type mockHTTPResponse struct {
	statusCode int
}

func (m *mockHTTPResponse) StatusCode() int {
	return m.statusCode
}

func TestFormatAPIError(t *testing.T) {
	tests := []struct {
		name           string
		operation      string
		err            error
		response       any
		wantContains   []string
		wantNotContain []string
	}{
		{
			name:      "500 server error",
			operation: "list monitors",
			err:       errors.New("internal server error"),
			response:  &mockHTTPResponse{statusCode: 500},
			wantContains: []string{
				"failed to list monitors",
				"status: 500",
				"Datadog API is experiencing issues",
				"https://status.datadoghq.com/",
			},
		},
		{
			name:      "502 bad gateway",
			operation: "get dashboard",
			err:       errors.New("bad gateway"),
			response:  &mockHTTPResponse{statusCode: 502},
			wantContains: []string{
				"failed to get dashboard",
				"status: 502",
				"Datadog API is experiencing issues",
			},
		},
		{
			name:      "504 gateway timeout",
			operation: "list monitors",
			err:       errors.New("gateway timeout"),
			response:  &mockHTTPResponse{statusCode: 504},
			wantContains: []string{
				"failed to list monitors",
				"status: 504",
				"Datadog API is experiencing issues",
				"try again later",
			},
		},
		{
			name:      "429 rate limit",
			operation: "create monitor",
			err:       errors.New("rate limited"),
			response:  &mockHTTPResponse{statusCode: 429},
			wantContains: []string{
				"failed to create monitor",
				"status: 429",
				"rate limited",
				"wait a moment",
			},
		},
		{
			name:      "403 forbidden",
			operation: "delete monitor",
			err:       errors.New("forbidden"),
			response:  &mockHTTPResponse{statusCode: 403},
			wantContains: []string{
				"failed to delete monitor",
				"status: 403",
				"Access denied",
				"API/App keys",
				"permissions",
			},
		},
		{
			name:      "401 unauthorized",
			operation: "get monitor",
			err:       errors.New("unauthorized"),
			response:  &mockHTTPResponse{statusCode: 401},
			wantContains: []string{
				"failed to get monitor",
				"status: 401",
				"Authentication failed",
				"pup auth login",
				"DD_API_KEY",
			},
		},
		{
			name:      "404 not found",
			operation: "get monitor",
			err:       errors.New("not found"),
			response:  &mockHTTPResponse{statusCode: 404},
			wantContains: []string{
				"failed to get monitor",
				"status: 404",
				"Resource not found",
				"Verify the ID",
			},
		},
		{
			name:      "400 bad request",
			operation: "create monitor",
			err:       errors.New("bad request"),
			response:  &mockHTTPResponse{statusCode: 400},
			wantContains: []string{
				"failed to create monitor",
				"status: 400",
				"Invalid request",
				"Check your parameters",
			},
		},
		{
			name:         "no response object",
			operation:    "list monitors",
			err:          errors.New("network error"),
			response:     nil,
			wantContains: []string{"failed to list monitors", "network error"},
			wantNotContain: []string{
				"status:",
				"Datadog API",
				"rate limited",
				"Authentication",
			},
		},
		{
			name:         "invalid response type",
			operation:    "get monitor",
			err:          errors.New("some error"),
			response:     "not a valid response",
			wantContains: []string{"failed to get monitor", "some error"},
			wantNotContain: []string{
				"status:",
				"Datadog API",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := formatAPIError(tt.operation, tt.err, tt.response)

			if err == nil {
				t.Fatal("formatAPIError() returned nil error")
			}

			errMsg := err.Error()

			for _, want := range tt.wantContains {
				if !strings.Contains(errMsg, want) {
					t.Errorf("formatAPIError() error message missing expected string:\n  got:  %q\n  want: %q", errMsg, want)
				}
			}

			for _, notWant := range tt.wantNotContain {
				if strings.Contains(errMsg, notWant) {
					t.Errorf("formatAPIError() error message contains unexpected string:\n  got:  %q\n  should not contain: %q", errMsg, notWant)
				}
			}
		})
	}
}

func TestFormatAPIError_AllStatusCodes(t *testing.T) {
	// Test that all documented status codes get special handling
	statusTests := []struct {
		code         int
		wantSpecial  bool
		wantContains string
	}{
		{400, true, "Invalid request"},
		{401, true, "Authentication failed"},
		{403, true, "Access denied"},
		{404, true, "Resource not found"},
		{429, true, "rate limited"},
		{500, true, "Datadog API is experiencing issues"},
		{502, true, "Datadog API is experiencing issues"},
		{503, true, "Datadog API is experiencing issues"},
		{504, true, "Datadog API is experiencing issues"},
		{200, false, ""},               // Should just show basic error
		{201, false, ""},               // Should just show basic error
		{418, true, "Invalid request"}, // Other 4xx
	}

	for _, tt := range statusTests {
		t.Run(fmt.Sprintf("%d", tt.code), func(t *testing.T) {
			err := formatAPIError("test operation", errors.New("test error"), &mockHTTPResponse{statusCode: tt.code})

			if err == nil {
				t.Fatal("formatAPIError() returned nil error")
			}

			errMsg := err.Error()

			if tt.wantSpecial && tt.wantContains != "" {
				if !strings.Contains(errMsg, tt.wantContains) {
					t.Errorf("formatAPIError() for status %d should contain %q, got: %q", tt.code, tt.wantContains, errMsg)
				}
			}

			// All errors should contain the basic info
			if !strings.Contains(errMsg, "failed to test operation") {
				t.Errorf("formatAPIError() should contain operation name, got: %q", errMsg)
			}
		})
	}
}

func TestTestCmd_EmptyKeys(t *testing.T) {
	// Save original config
	origCfg := cfg

	// Set up test config with empty keys
	cfg = &config.Config{
		Site:   "datadoghq.com",
		APIKey: "",
		AppKey: "",
	}
	defer func() { cfg = origCfg }()

	// Execute test command
	err := testCmd.RunE(testCmd, []string{})

	// Should not panic and should succeed
	if err != nil {
		t.Errorf("testCmd.RunE() with empty keys failed: %v", err)
	}
}

func TestTestCmd_ShortKeys(t *testing.T) {
	// Save original config
	origCfg := cfg

	// Set up test config with short keys
	cfg = &config.Config{
		Site:   "datadoghq.com",
		APIKey: "short",
		AppKey: "key",
	}
	defer func() { cfg = origCfg }()

	// Execute test command
	err := testCmd.RunE(testCmd, []string{})

	// Should not panic and should succeed
	if err != nil {
		t.Errorf("testCmd.RunE() with short keys failed: %v", err)
	}
}

func TestTestCmd_ValidKeys(t *testing.T) {
	// Save original config
	origCfg := cfg

	// Set up test config with valid length keys
	cfg = &config.Config{
		Site:   "datadoghq.com",
		APIKey: "1234567890abcdef1234567890abcdef",
		AppKey: "abcdefghijklmnopqrstuvwxyz123456",
	}
	defer func() { cfg = origCfg }()

	// Execute test command
	err := testCmd.RunE(testCmd, []string{})

	// Should not panic and should succeed
	if err != nil {
		t.Errorf("testCmd.RunE() with valid keys failed: %v", err)
	}
}

func TestTestCmd_InvalidSite(t *testing.T) {
	// Save original config
	origCfg := cfg

	// Set up test config with empty site
	cfg = &config.Config{
		Site:   "",
		APIKey: "1234567890abcdef",
		AppKey: "abcdefghijklmnop",
	}
	defer func() { cfg = origCfg }()

	// Execute test command
	err := testCmd.RunE(testCmd, []string{})

	// Should fail validation
	if err == nil {
		t.Error("testCmd.RunE() with empty site should fail")
	}
	if !strings.Contains(err.Error(), "DD_SITE") {
		t.Errorf("testCmd.RunE() error should mention DD_SITE, got: %v", err)
	}
}

func TestExtractAPIErrorBody(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "GenericOpenAPIError with body",
			err: datadog.GenericOpenAPIError{
				ErrorBody:    []byte(`{"errors":["Invalid query: avg:nonexistent.metric{*}"]}`),
				ErrorMessage: "400 Bad Request",
			},
			want: `{"errors":["Invalid query: avg:nonexistent.metric{*}"]}`,
		},
		{
			name: "GenericOpenAPIError with empty body",
			err: datadog.GenericOpenAPIError{
				ErrorBody:    []byte{},
				ErrorMessage: "400 Bad Request",
			},
			want: "",
		},
		{
			name: "GenericOpenAPIError with nil body",
			err: datadog.GenericOpenAPIError{
				ErrorBody:    nil,
				ErrorMessage: "400 Bad Request",
			},
			want: "",
		},
		{
			name: "wrapped GenericOpenAPIError",
			err: fmt.Errorf("api call failed: %w", datadog.GenericOpenAPIError{
				ErrorBody:    []byte(`{"errors":["bad query"]}`),
				ErrorMessage: "400 Bad Request",
			}),
			want: `{"errors":["bad query"]}`,
		},
		{
			name: "non-GenericOpenAPIError",
			err:  errors.New("some other error"),
			want: "",
		},
		{
			name: "nil error",
			err:  nil,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractAPIErrorBody(tt.err)
			if got != tt.want {
				t.Errorf("extractAPIErrorBody() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatAPIError_IncludesResponseBody(t *testing.T) {
	// This test verifies that formatAPIError surfaces the API response body
	// from GenericOpenAPIError, which was previously lost because the code
	// tried to re-read the already-consumed http.Response.Body.
	apiErr := datadog.GenericOpenAPIError{
		ErrorBody:    []byte(`{"errors":["Query parse error: unknown metric"]}`),
		ErrorMessage: "400 Bad Request",
	}

	err := formatAPIError("query metrics", apiErr, &mockHTTPResponse{statusCode: 400})
	errMsg := err.Error()

	if !strings.Contains(errMsg, "unknown metric") {
		t.Errorf("formatAPIError() should include API response body, got: %q", errMsg)
	}
	if !strings.Contains(errMsg, "status: 400") {
		t.Errorf("formatAPIError() should include status code, got: %q", errMsg)
	}
}
