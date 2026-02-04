// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package client

import (
	"strings"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/pup/pkg/config"
)

func TestNew_WithAPIKeys(t *testing.T) {
	cfg := &config.Config{
		APIKey: "test-api-key",
		AppKey: "test-app-key",
		Site:   "datadoghq.com",
	}

	client, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if client == nil {
		t.Fatal("New() returned nil")
	}

	if client.ctx == nil {
		t.Error("ctx is nil")
	}

	if client.api == nil {
		t.Error("api is nil")
	}

	if client.config != cfg {
		t.Error("config not set correctly")
	}

	// Verify context contains API keys
	apiKeys, ok := client.ctx.Value(datadog.ContextAPIKeys).(map[string]datadog.APIKey)
	if !ok {
		t.Fatal("Context does not contain API keys")
	}

	if apiKeys["apiKeyAuth"].Key != "test-api-key" {
		t.Errorf("apiKeyAuth = %s, want test-api-key", apiKeys["apiKeyAuth"].Key)
	}

	if apiKeys["appKeyAuth"].Key != "test-app-key" {
		t.Errorf("appKeyAuth = %s, want test-app-key", apiKeys["appKeyAuth"].Key)
	}
}

func TestNew_NoAuthentication(t *testing.T) {
	cfg := &config.Config{
		APIKey: "",
		AppKey: "",
		Site:   "datadoghq.com",
	}

	_, err := New(cfg)
	if err == nil {
		t.Error("New() expected error but got none")
	}

	if !strings.Contains(err.Error(), "authentication required") {
		t.Errorf("Error = %v, want authentication error", err)
	}
}

func TestNew_MissingAPIKey(t *testing.T) {
	cfg := &config.Config{
		APIKey: "",
		AppKey: "test-app-key",
		Site:   "datadoghq.com",
	}

	_, err := New(cfg)
	if err == nil {
		t.Error("New() expected error but got none")
	}

	if !strings.Contains(err.Error(), "authentication required") {
		t.Errorf("Error = %v, want authentication error", err)
	}
}

func TestNew_MissingAppKey(t *testing.T) {
	cfg := &config.Config{
		APIKey: "test-api-key",
		AppKey: "",
		Site:   "datadoghq.com",
	}

	_, err := New(cfg)
	if err == nil {
		t.Error("New() expected error but got none")
	}

	if !strings.Contains(err.Error(), "authentication required") {
		t.Errorf("Error = %v, want authentication error", err)
	}
}

func TestNew_DifferentSites(t *testing.T) {
	tests := []struct {
		name string
		site string
	}{
		{"US1", "datadoghq.com"},
		{"EU", "datadoghq.eu"},
		{"US3", "us3.datadoghq.com"},
		{"US5", "us5.datadoghq.com"},
		{"AP1", "ap1.datadoghq.com"},
		{"Gov", "ddog-gov.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				APIKey: "test-api-key",
				AppKey: "test-app-key",
				Site:   tt.site,
			}

			client, err := New(cfg)
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}

			if client == nil {
				t.Fatal("New() returned nil")
			}

			if client.config.Site != tt.site {
				t.Errorf("Site = %s, want %s", client.config.Site, tt.site)
			}
		})
	}
}

func TestClient_Context(t *testing.T) {
	cfg := &config.Config{
		APIKey: "test-api-key",
		AppKey: "test-app-key",
		Site:   "datadoghq.com",
	}

	client, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	ctx := client.Context()
	if ctx == nil {
		t.Error("Context() returned nil")
	}

	// Verify context contains API keys
	apiKeys, ok := ctx.Value(datadog.ContextAPIKeys).(map[string]datadog.APIKey)
	if !ok {
		t.Fatal("Context does not contain API keys")
	}

	if apiKeys["apiKeyAuth"].Key != "test-api-key" {
		t.Errorf("apiKeyAuth = %s, want test-api-key", apiKeys["apiKeyAuth"].Key)
	}
}

func TestClient_V1(t *testing.T) {
	cfg := &config.Config{
		APIKey: "test-api-key",
		AppKey: "test-app-key",
		Site:   "datadoghq.com",
	}

	client, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	api := client.V1()
	if api == nil {
		t.Error("V1() returned nil")
	}

	// Verify it's the same instance as the internal api
	if api != client.api {
		t.Error("V1() returned different instance")
	}
}

func TestClient_V2(t *testing.T) {
	cfg := &config.Config{
		APIKey: "test-api-key",
		AppKey: "test-app-key",
		Site:   "datadoghq.com",
	}

	client, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	api := client.V2()
	if api == nil {
		t.Error("V2() returned nil")
	}

	// Verify it's the same instance as the internal api
	if api != client.api {
		t.Error("V2() returned different instance")
	}
}

func TestClient_API(t *testing.T) {
	cfg := &config.Config{
		APIKey: "test-api-key",
		AppKey: "test-app-key",
		Site:   "datadoghq.com",
	}

	client, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	api := client.API()
	if api == nil {
		t.Error("API() returned nil")
	}

	// Verify it's the same instance as the internal api
	if api != client.api {
		t.Error("API() returned different instance")
	}

	// Verify V1(), V2(), and API() all return the same instance
	if client.V1() != client.V2() || client.V1() != client.API() {
		t.Error("V1(), V2(), and API() should return the same instance")
	}
}

func TestClient_Config(t *testing.T) {
	cfg := &config.Config{
		APIKey: "test-api-key",
		AppKey: "test-app-key",
		Site:   "datadoghq.com",
	}

	client, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	returnedCfg := client.Config()
	if returnedCfg == nil {
		t.Error("Config() returned nil")
	}

	if returnedCfg != cfg {
		t.Error("Config() returned different instance")
	}

	if returnedCfg.Site != "datadoghq.com" {
		t.Errorf("Site = %s, want datadoghq.com", returnedCfg.Site)
	}
}

func TestClient_APIConfiguration(t *testing.T) {
	cfg := &config.Config{
		APIKey: "test-api-key",
		AppKey: "test-app-key",
		Site:   "datadoghq.eu",
	}

	client, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Access the configuration through the API client
	// Note: This test verifies that the configuration was set up correctly
	// but we can't directly access the Host field from the client
	// so we verify through successful client creation

	if client.api == nil {
		t.Error("API client not initialized")
	}

	// Verify the configuration was created for the correct site
	// by checking that the client was successfully created with the site config
	if client.config.Site != "datadoghq.eu" {
		t.Errorf("Configuration site = %s, want datadoghq.eu", client.config.Site)
	}
}
