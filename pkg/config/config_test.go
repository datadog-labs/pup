// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// NOTE: Not parallel - modifies env vars
	// Save original env vars
	origAPIKey := os.Getenv("DD_API_KEY")
	origAppKey := os.Getenv("DD_APP_KEY")
	origSite := os.Getenv("DD_SITE")
	origAutoApprove := os.Getenv("DD_AUTO_APPROVE")
	origCLIAutoApprove := os.Getenv("DD_CLI_AUTO_APPROVE")

	// Clean up after test
	defer func() {
		os.Setenv("DD_API_KEY", origAPIKey)
		os.Setenv("DD_APP_KEY", origAppKey)
		os.Setenv("DD_SITE", origSite)
		os.Setenv("DD_AUTO_APPROVE", origAutoApprove)
		os.Setenv("DD_CLI_AUTO_APPROVE", origCLIAutoApprove)
	}()

	tests := []struct {
		name        string
		envVars     map[string]string
		wantAPIKey  string
		wantAppKey  string
		wantSite    string
		wantApprove bool
	}{
		{
			name: "all env vars set",
			envVars: map[string]string{
				"DD_API_KEY":      "test-api-key",
				"DD_APP_KEY":      "test-app-key",
				"DD_SITE":         "datadoghq.eu",
				"DD_AUTO_APPROVE": "true",
			},
			wantAPIKey:  "test-api-key",
			wantAppKey:  "test-app-key",
			wantSite:    "datadoghq.eu",
			wantApprove: true,
		},
		{
			name: "default site",
			envVars: map[string]string{
				"DD_API_KEY": "api",
				"DD_APP_KEY": "app",
			},
			wantAPIKey:  "api",
			wantAppKey:  "app",
			wantSite:    "datadoghq.com",
			wantApprove: false,
		},
		{
			name: "auto approve with DD_CLI_AUTO_APPROVE",
			envVars: map[string]string{
				"DD_CLI_AUTO_APPROVE": "true",
			},
			wantSite:    "datadoghq.com",
			wantApprove: true,
		},
		{
			name: "no auto approve when false",
			envVars: map[string]string{
				"DD_AUTO_APPROVE": "false",
			},
			wantSite:    "datadoghq.com",
			wantApprove: false,
		},
		{
			name:        "empty config uses defaults",
			envVars:     map[string]string{},
			wantSite:    "datadoghq.com",
			wantApprove: false,
		},
		{
			name: "on-call domain via DD_SITE",
			envVars: map[string]string{
				"DD_API_KEY": "key",
				"DD_APP_KEY": "app",
				"DD_SITE":    "navy.oncall.datadoghq.com",
			},
			wantAPIKey: "key",
			wantAppKey: "app",
			wantSite:   "navy.oncall.datadoghq.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all env vars
			os.Unsetenv("DD_API_KEY")
			os.Unsetenv("DD_APP_KEY")
			os.Unsetenv("DD_SITE")
			os.Unsetenv("DD_AUTO_APPROVE")
			os.Unsetenv("DD_CLI_AUTO_APPROVE")

			// Set test env vars
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			cfg, err := Load()
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			if cfg.APIKey != tt.wantAPIKey {
				t.Errorf("APIKey = %q, want %q", cfg.APIKey, tt.wantAPIKey)
			}
			if cfg.AppKey != tt.wantAppKey {
				t.Errorf("AppKey = %q, want %q", cfg.AppKey, tt.wantAppKey)
			}
			if cfg.Site != tt.wantSite {
				t.Errorf("Site = %q, want %q", cfg.Site, tt.wantSite)
			}
			if cfg.AutoApprove != tt.wantApprove {
				t.Errorf("AutoApprove = %v, want %v", cfg.AutoApprove, tt.wantApprove)
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Site: "datadoghq.com",
			},
			wantErr: false,
		},
		{
			name: "valid config with keys",
			config: &Config{
				APIKey: "test-api-key",
				AppKey: "test-app-key",
				Site:   "datadoghq.eu",
			},
			wantErr: false,
		},
		{
			name: "empty site",
			config: &Config{
				Site: "",
			},
			wantErr: true,
		},
		{
			name: "valid US3 site",
			config: &Config{
				Site: "us3.datadoghq.com",
			},
			wantErr: false,
		},
		{
			name: "valid gov site",
			config: &Config{
				Site: "ddog-gov.com",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsOnCallSite(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		site string
		want bool
	}{
		// Standard sites — not on-call
		{name: "US1", site: "datadoghq.com", want: false},
		{name: "EU", site: "datadoghq.eu", want: false},
		{name: "US3", site: "us3.datadoghq.com", want: false},
		{name: "Gov", site: "ddog-gov.com", want: false},

		// On-call domains
		{name: "on-call navy", site: "navy.oncall.datadoghq.com", want: true},
		{name: "on-call army", site: "army.oncall.datadoghq.com", want: true},
		{name: "on-call staging", site: "test.oncall.datad0g.com", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := IsOnCallSite(tt.site)
			if got != tt.want {
				t.Errorf("IsOnCallSite(%q) = %v, want %v", tt.site, got, tt.want)
			}
		})
	}
}

func TestConfig_GetAPIURL(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		site string
		want string
	}{
		{
			name: "US1 site",
			site: "datadoghq.com",
			want: "https://api.datadoghq.com",
		},
		{
			name: "EU site",
			site: "datadoghq.eu",
			want: "https://api.datadoghq.eu",
		},
		{
			name: "US3 site",
			site: "us3.datadoghq.com",
			want: "https://api.us3.datadoghq.com",
		},
		{
			name: "US5 site",
			site: "us5.datadoghq.com",
			want: "https://api.us5.datadoghq.com",
		},
		{
			name: "AP1 site",
			site: "ap1.datadoghq.com",
			want: "https://api.ap1.datadoghq.com",
		},
		{
			name: "Gov site",
			site: "ddog-gov.com",
			want: "https://api.ddog-gov.com",
		},
		{
			name: "Staging site",
			site: "datad0g.com",
			want: "https://api.datad0g.com",
		},
		{
			name: "on-call domain used as-is",
			site: "navy.oncall.datadoghq.com",
			want: "https://navy.oncall.datadoghq.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := &Config{Site: tt.site}
			got := cfg.GetAPIURL()
			if got != tt.want {
				t.Errorf("GetAPIURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestConfig_GetAPIHost(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		site string
		want string
	}{
		{name: "US1", site: "datadoghq.com", want: "api.datadoghq.com"},
		{name: "EU", site: "datadoghq.eu", want: "api.datadoghq.eu"},
		{name: "US3", site: "us3.datadoghq.com", want: "api.us3.datadoghq.com"},
		{name: "Gov", site: "ddog-gov.com", want: "api.ddog-gov.com"},
		{name: "on-call", site: "navy.oncall.datadoghq.com", want: "navy.oncall.datadoghq.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := &Config{Site: tt.site}
			got := cfg.GetAPIHost()
			if got != tt.want {
				t.Errorf("GetAPIHost() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetEnvWithDefault(t *testing.T) {
	t.Parallel()
	// Save original
	origValue := os.Getenv("TEST_ENV_VAR")
	defer os.Setenv("TEST_ENV_VAR", origValue)

	tests := []struct {
		name         string
		key          string
		defaultValue string
		setValue     string
		setEnv       bool
		want         string
	}{
		{
			name:         "env var set",
			key:          "TEST_ENV_VAR",
			defaultValue: "default",
			setValue:     "custom",
			setEnv:       true,
			want:         "custom",
		},
		{
			name:         "env var not set",
			key:          "TEST_ENV_VAR",
			defaultValue: "default",
			setEnv:       false,
			want:         "default",
		},
		{
			name:         "env var set to empty",
			key:          "TEST_ENV_VAR",
			defaultValue: "default",
			setValue:     "",
			setEnv:       true,
			want:         "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				os.Setenv(tt.key, tt.setValue)
			} else {
				os.Unsetenv(tt.key)
			}

			got := getEnvWithDefault(tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("getEnvWithDefault() = %q, want %q", got, tt.want)
			}
		})
	}
}
