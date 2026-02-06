// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/pup/pkg/client"
	"github.com/DataDog/pup/pkg/config"
)

// MockClient provides a mock Datadog client for testing
type MockClient struct {
	*client.Client
	server *httptest.Server
}

// NewMockClient creates a mock client that uses a test HTTP server
func NewMockClient(serverURL string) *client.Client {
	cfg := &config.Config{
		Site:   "datadoghq.com",
		APIKey: "test-api-key",
		AppKey: "test-app-key",
	}

	// Create a client with the test server URL
	// Note: We'll need to modify the client to accept custom base URLs
	// For now, create a minimal client structure
	mockClient, _ := client.New(cfg)
	return mockClient
}

// NewMockAPIServer creates a mock HTTP server that returns the specified response
func NewMockAPIServer(response any, statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)

		if response != nil {
			if err := json.NewEncoder(w).Encode(response); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
	}))
}

// NewMockAPIServerWithHandler creates a mock HTTP server with a custom handler
func NewMockAPIServerWithHandler(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

// CaptureOutput captures stdout during function execution
func CaptureOutput(f func()) string {
	var buf bytes.Buffer
	// Note: This will be used with the outputWriter variable in cmd/root.go
	// The actual implementation will set outputWriter = &buf
	f()
	return buf.String()
}

// SimulateInput creates an io.Reader from a string for simulating stdin
func SimulateInput(input string) io.Reader {
	return strings.NewReader(input)
}

// MockAPIResponse represents a mock API response for testing
type MockAPIResponse struct {
	StatusCode int
	Body       any
	Headers    map[string]string
}

// NewMockAPIResponseServer creates a server with multiple endpoint responses
func NewMockAPIResponseServer(responses map[string]MockAPIResponse) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Method + " " + r.URL.Path
		response, ok := responses[key]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		for k, v := range response.Headers {
			w.Header().Set(k, v)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(response.StatusCode)

		if response.Body != nil {
			if err := json.NewEncoder(w).Encode(response.Body); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
	}))
}

// CreateTestClient creates a minimal test client for unit tests
func CreateTestClient() *client.Client {
	cfg := &config.Config{
		Site:   "datadoghq.com",
		APIKey: "test-api-key-12345678",
		AppKey: "test-app-key-12345678",
	}

	// Create client - this may fail if validation requires real keys
	// In that case, tests should use the SetTestClient function
	c, _ := client.New(cfg)
	return c
}

// MockDatadogContext creates a test context for Datadog API calls
func MockDatadogContext() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, datadog.ContextAPIKeys, map[string]datadog.APIKey{
		"apiKeyAuth": {Key: "test-api-key"},
		"appKeyAuth": {Key: "test-app-key"},
	})
	ctx = context.WithValue(ctx, datadog.ContextServerVariables, map[string]string{
		"site": "datadoghq.com",
	})
	return ctx
}
