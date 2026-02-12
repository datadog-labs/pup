// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package callback

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	t.Parallel()
	server, err := NewServer()
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	if server == nil {
		t.Fatal("NewServer() returned nil")
	}

	// Verify port is one of the DCR ports
	port := server.Port()
	found := false
	for _, p := range DCRRedirectPorts {
		if port == p {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Port %d not in DCRRedirectPorts %v", port, DCRRedirectPorts)
	}

	// Verify result channel is created
	if server.resultCh == nil {
		t.Error("resultCh is nil")
	}
}

func TestServer_Port(t *testing.T) {
	t.Parallel()
	server, err := NewServer()
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	port := server.Port()
	if port == 0 {
		t.Error("Port() returned 0")
	}

	// Verify port is in the expected range
	found := false
	for _, p := range DCRRedirectPorts {
		if port == p {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Port() = %d, want one of %v", port, DCRRedirectPorts)
	}
}

func TestServer_RedirectURI(t *testing.T) {
	t.Parallel()
	server, err := NewServer()
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	uri := server.RedirectURI()

	// Verify URI format matches TypeScript PR #84
	expectedPrefix := "http://127.0.0.1:"
	if !strings.HasPrefix(uri, expectedPrefix) {
		t.Errorf("RedirectURI() = %s, want prefix %s", uri, expectedPrefix)
	}

	expectedSuffix := "/oauth/callback"
	if !strings.HasSuffix(uri, expectedSuffix) {
		t.Errorf("RedirectURI() = %s, want suffix %s", uri, expectedSuffix)
	}

	// Verify URI includes the correct port
	expectedURI := fmt.Sprintf("http://127.0.0.1:%d/oauth/callback", server.Port())
	if uri != expectedURI {
		t.Errorf("RedirectURI() = %s, want %s", uri, expectedURI)
	}
}

func TestServer_StartStop(t *testing.T) {
	// NOTE: Not parallel - binds to network port
	server, err := NewServer()
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	// Start server
	if err := server.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Verify server is running by making a request
	resp, err := http.Get(server.RedirectURI())
	if err != nil {
		t.Errorf("Server not responding: %v", err)
	} else {
		resp.Body.Close()
	}

	// Stop server
	if err := server.Stop(); err != nil {
		t.Errorf("Stop() error = %v", err)
	}

	// Verify server is stopped
	time.Sleep(100 * time.Millisecond)
	_, err = http.Get(server.RedirectURI())
	if err == nil {
		t.Error("Server still responding after Stop()")
	}
}

func TestServer_Stop_WhenNotStarted(t *testing.T) {
	t.Parallel()
	server, err := NewServer()
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	// Stop without starting should not error
	if err := server.Stop(); err != nil {
		t.Errorf("Stop() error = %v, want nil", err)
	}
}

func TestServer_WaitForCallback_Success(t *testing.T) {
	// NOTE: Not parallel - binds to network port
	server, err := NewServer()
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	if err := server.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer server.Stop()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Make callback request in background
	go func() {
		url := fmt.Sprintf("%s?code=test-code&state=test-state", server.RedirectURI())
		resp, err := http.Get(url)
		if err != nil {
			t.Errorf("Callback request failed: %v", err)
			return
		}
		defer resp.Body.Close()
	}()

	// Wait for callback
	result, err := server.WaitForCallback(5 * time.Second)
	if err != nil {
		t.Fatalf("WaitForCallback() error = %v", err)
	}

	if result.Code != "test-code" {
		t.Errorf("Code = %s, want test-code", result.Code)
	}
	if result.State != "test-state" {
		t.Errorf("State = %s, want test-state", result.State)
	}
	if result.Error != "" {
		t.Errorf("Error = %s, want empty", result.Error)
	}
}

func TestServer_WaitForCallback_Error(t *testing.T) {
	// NOTE: Not parallel - binds to network port
	server, err := NewServer()
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	if err := server.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer server.Stop()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Make callback request with error in background
	go func() {
		url := fmt.Sprintf("%s?error=access_denied&error_description=User+denied+access", server.RedirectURI())
		resp, err := http.Get(url)
		if err != nil {
			t.Errorf("Callback request failed: %v", err)
			return
		}
		defer resp.Body.Close()
	}()

	// Wait for callback
	result, err := server.WaitForCallback(5 * time.Second)
	if err != nil {
		t.Fatalf("WaitForCallback() error = %v", err)
	}

	if result.Error != "access_denied" {
		t.Errorf("Error = %s, want access_denied", result.Error)
	}
	if result.ErrorDescription != "User denied access" {
		t.Errorf("ErrorDescription = %s, want User denied access", result.ErrorDescription)
	}
	if result.Code != "" {
		t.Errorf("Code = %s, want empty", result.Code)
	}
}

func TestServer_WaitForCallback_Timeout(t *testing.T) {
	// NOTE: Not parallel - binds to network port
	server, err := NewServer()
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	if err := server.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer server.Stop()

	// Wait for callback without making request
	_, err = server.WaitForCallback(500 * time.Millisecond)
	if err == nil {
		t.Error("WaitForCallback() expected timeout error but got none")
	}
	if !strings.Contains(err.Error(), "timeout") {
		t.Errorf("Error = %v, want timeout error", err)
	}
}

func TestServer_HandleCallback_SuccessResponse(t *testing.T) {
	t.Parallel()
	server := &Server{
		port:     8000,
		resultCh: make(chan CallbackResult, 1),
	}

	// Create test request
	req := httptest.NewRequest("GET", "/oauth/callback?code=test-code&state=test-state", nil)
	w := httptest.NewRecorder()

	// Handle callback
	server.handleCallback(w, req)

	// Verify response
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if !strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
		t.Errorf("Content-Type = %s, want text/html", resp.Header.Get("Content-Type"))
	}

	// Verify result was sent to channel
	select {
	case result := <-server.resultCh:
		if result.Code != "test-code" {
			t.Errorf("Code = %s, want test-code", result.Code)
		}
		if result.State != "test-state" {
			t.Errorf("State = %s, want test-state", result.State)
		}
	case <-time.After(1 * time.Second):
		t.Error("Result not received on channel")
	}
}

func TestServer_HandleCallback_ErrorResponse(t *testing.T) {
	t.Parallel()
	server := &Server{
		port:     8000,
		resultCh: make(chan CallbackResult, 1),
	}

	// Create test request with error
	req := httptest.NewRequest("GET", "/oauth/callback?error=invalid_request&error_description=Missing+parameter", nil)
	w := httptest.NewRecorder()

	// Handle callback
	server.handleCallback(w, req)

	// Verify response
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}

	if !strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
		t.Errorf("Content-Type = %s, want text/html", resp.Header.Get("Content-Type"))
	}

	// Verify result was sent to channel
	select {
	case result := <-server.resultCh:
		if result.Error != "invalid_request" {
			t.Errorf("Error = %s, want invalid_request", result.Error)
		}
		if result.ErrorDescription != "Missing parameter" {
			t.Errorf("ErrorDescription = %s, want Missing parameter", result.ErrorDescription)
		}
	case <-time.After(1 * time.Second):
		t.Error("Result not received on channel")
	}
}

func TestServer_RenderSuccess(t *testing.T) {
	t.Parallel()
	server := &Server{port: 8000}

	w := httptest.NewRecorder()
	server.renderSuccess(w)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if !strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
		t.Errorf("Content-Type = %s, want text/html", resp.Header.Get("Content-Type"))
	}

	// Verify response contains success message
	body := w.Body.String()
	if !strings.Contains(body, "Authentication Successful") {
		t.Error("Response missing 'Authentication Successful' message")
	}
	if !strings.Contains(body, "✓") {
		t.Error("Response missing success icon")
	}
}

func TestServer_RenderError(t *testing.T) {
	t.Parallel()
	server := &Server{port: 8000}

	w := httptest.NewRecorder()
	server.renderError(w, "access_denied", "User denied access")

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}

	if !strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
		t.Errorf("Content-Type = %s, want text/html", resp.Header.Get("Content-Type"))
	}

	// Verify response contains error details
	body := w.Body.String()
	if !strings.Contains(body, "Authentication Failed") {
		t.Error("Response missing 'Authentication Failed' message")
	}
	if !strings.Contains(body, "✗") {
		t.Error("Response missing error icon")
	}
	if !strings.Contains(body, "access_denied") {
		t.Error("Response missing error code")
	}
	if !strings.Contains(body, "User denied access") {
		t.Error("Response missing error description")
	}
}

func TestServer_RenderError_NoDescription(t *testing.T) {
	t.Parallel()
	server := &Server{port: 8000}

	w := httptest.NewRecorder()
	server.renderError(w, "server_error", "")

	resp := w.Result()
	defer resp.Body.Close()

	// Verify response contains error code even without description
	body := w.Body.String()
	if !strings.Contains(body, "server_error") {
		t.Error("Response missing error code")
	}
}

func TestDCRRedirectPorts(t *testing.T) {
	t.Parallel()
	// Verify DCRRedirectPorts matches TypeScript PR #84
	expected := []int{8000, 8080, 8888, 9000}

	if len(DCRRedirectPorts) != len(expected) {
		t.Errorf("DCRRedirectPorts length = %d, want %d", len(DCRRedirectPorts), len(expected))
	}

	for i, port := range expected {
		if DCRRedirectPorts[i] != port {
			t.Errorf("DCRRedirectPorts[%d] = %d, want %d", i, DCRRedirectPorts[i], port)
		}
	}
}
