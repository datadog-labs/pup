// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/datadog-labs/pup/pkg/config"
)

// fixedTime is used across sparkline tests so outage timestamps are deterministic.
// 2024-01-15 00:00:00 UTC
var fixedTime = time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

const testOutagesJSON = `{
  "data": {
    "attributes": {
      "provider_data": [
        {
          "provider_name": "aws-s3",
          "provider_service": "S3",
          "display_name": "Amazon S3",
          "integration_id": "aws-s3",
          "status_url": "https://health.aws.amazon.com/health/status",
          "monitoring_start_date": 1700000000000,
          "monitored_api_patterns": ["s3.amazonaws.com"],
          "outages": [
            {"start": 1700001000000, "end": 1700002000000, "status": "resolved", "impacted_region": "us-east-1"},
            {"start": 1700003000000, "end": 0, "status": "active", "impacted_region": "us-west-2"}
          ]
        },
        {
          "provider_name": "stripe",
          "display_name": "Stripe",
          "integration_id": "stripe",
          "status_url": "https://status.stripe.com",
          "monitoring_start_date": 1700000000000,
          "monitored_api_patterns": ["api.stripe.com"],
          "outages": [
            {"start": 1700001000000, "end": 1700002000000, "status": "resolved"}
          ]
        },
        {
          "provider_name": "gcp",
          "display_name": "Google Cloud Platform",
          "integration_id": "gcp",
          "status_url": "https://status.cloud.google.com",
          "monitoring_start_date": 1700000000000,
          "monitored_api_patterns": ["googleapis.com"],
          "outages": []
        }
      ]
    },
    "id": "outages",
    "type": "provider_data"
  }
}`

func setupThirdPartyTestServer(t *testing.T, statusCode int, body string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_, _ = w.Write([]byte(body))
	}))
}

// redirectTransport intercepts all HTTP requests and redirects them to a test server.
type redirectTransport struct {
	target string
}

func (t *redirectTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	req.URL.Host = t.target[len("http://"):]
	return http.DefaultTransport.RoundTrip(req)
}

// stripANSI removes ANSI escape codes for test assertions.
func stripANSI(s string) string {
	var b strings.Builder
	inEscape := false
	for _, r := range s {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func TestStatusPagesThirdPartyCmd(t *testing.T) {
	if statusPagesThirdPartyCmd == nil {
		t.Fatal("statusPagesThirdPartyCmd is nil")
	}
	if statusPagesThirdPartyCmd.Use != "third-party" {
		t.Errorf("Use = %s, want third-party", statusPagesThirdPartyCmd.Use)
	}
	if statusPagesThirdPartyCmd.Short == "" {
		t.Error("Short description is empty")
	}
}

func TestStatusPagesThirdPartyCmd_RegisteredAsSubcommand(t *testing.T) {
	found := false
	for _, cmd := range statusPagesCmd.Commands() {
		if cmd.Use == "third-party" {
			found = true
			break
		}
	}
	if !found {
		t.Error("third-party not registered as subcommand of status-pages")
	}
}

func TestFetchThirdPartyOutages(t *testing.T) {
	server := setupThirdPartyTestServer(t, http.StatusOK, testOutagesJSON)
	defer server.Close()

	origClient := httpClient
	defer func() { httpClient = origClient }()
	httpClient = server.Client()

	t.Run("parses valid response", func(t *testing.T) {
		resp, err := httpClient.Get(server.URL)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("unexpected status: %d", resp.StatusCode)
		}
	})
}

func TestFetchThirdPartyOutages_ServerError(t *testing.T) {
	server := setupThirdPartyTestServer(t, http.StatusInternalServerError, "server error")
	defer server.Close()

	origClient := httpClient
	defer func() { httpClient = origClient }()
	httpClient = server.Client()

	resp, err := httpClient.Get(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", resp.StatusCode)
	}
}

func TestFilterProviders(t *testing.T) {
	providers := []thirdPartyProvider{
		{
			ProviderName: "aws-s3",
			DisplayName:  "Amazon S3",
			Outages: []thirdPartyOutage{
				{Status: "resolved"},
				{Status: "active"},
			},
		},
		{
			ProviderName: "stripe",
			DisplayName:  "Stripe",
			Outages: []thirdPartyOutage{
				{Status: "resolved"},
			},
		},
		{
			ProviderName: "gcp",
			DisplayName:  "Google Cloud Platform",
			Outages:      []thirdPartyOutage{},
		},
	}

	tests := []struct {
		name       string
		search     string
		activeOnly bool
		wantCount  int
		wantNames  []string
	}{
		{
			name:      "no filter returns all",
			wantCount: 3,
		},
		{
			name:      "search by provider name",
			search:    "aws",
			wantCount: 1,
			wantNames: []string{"aws-s3"},
		},
		{
			name:      "search by display name",
			search:    "google",
			wantCount: 1,
			wantNames: []string{"gcp"},
		},
		{
			name:      "search by display name Amazon",
			search:    "amazon",
			wantCount: 1,
			wantNames: []string{"aws-s3"},
		},
		{
			name:      "search case insensitive",
			search:    "STRIPE",
			wantCount: 1,
			wantNames: []string{"stripe"},
		},
		{
			name:       "active only",
			activeOnly: true,
			wantCount:  1,
			wantNames:  []string{"aws-s3"},
		},
		{
			name:       "search and active combined",
			search:     "stripe",
			activeOnly: true,
			wantCount:  0,
		},
		{
			name:      "no match",
			search:    "nonexistent",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterProviders(providers, tt.search, tt.activeOnly)
			if len(result) != tt.wantCount {
				t.Errorf("got %d providers, want %d", len(result), tt.wantCount)
			}
			if tt.wantNames != nil {
				for i, name := range tt.wantNames {
					if i < len(result) && result[i].ProviderName != name {
						t.Errorf("provider[%d] = %s, want %s", i, result[i].ProviderName, name)
					}
				}
			}
		})
	}
}

func TestProviderCurrentStatus(t *testing.T) {
	tests := []struct {
		name       string
		provider   thirdPartyProvider
		wantStatus string
	}{
		{
			name:       "operational with no outages",
			provider:   thirdPartyProvider{Outages: []thirdPartyOutage{}},
			wantStatus: "operational",
		},
		{
			name: "operational with only resolved outages",
			provider: thirdPartyProvider{
				Outages: []thirdPartyOutage{{Status: "resolved"}, {Status: "resolved"}},
			},
			wantStatus: "operational",
		},
		{
			name: "active outage",
			provider: thirdPartyProvider{
				Outages: []thirdPartyOutage{{Status: "resolved"}, {Status: "active"}},
			},
			wantStatus: "active",
		},
		{
			name: "investigating outage",
			provider: thirdPartyProvider{
				Outages: []thirdPartyOutage{{Status: "investigating"}},
			},
			wantStatus: "investigating",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := providerCurrentStatus(tt.provider)
			if status != tt.wantStatus {
				t.Errorf("status = %q, want %q", status, tt.wantStatus)
			}
		})
	}
}

func TestBuildSparkline(t *testing.T) {
	origTimeNow := timeNow
	defer func() { timeNow = origTimeNow }()
	timeNow = func() time.Time { return fixedTime }

	nowMs := fixedTime.UnixMilli()

	t.Run("all green when no outages", func(t *testing.T) {
		p := thirdPartyProvider{
			MonitoringStartDate: nowMs - 60*dayMs, // started 60 days ago
			Outages:             []thirdPartyOutage{},
		}
		raw := buildSparkline(p)
		plain := stripANSI(raw)
		if len([]rune(plain)) != sparklineWidth {
			t.Errorf("sparkline length = %d, want %d", len([]rune(plain)), sparklineWidth)
		}
		if strings.Contains(raw, ansiRed) {
			t.Error("expected no red blocks for provider with no outages")
		}
		if !strings.Contains(raw, ansiGreen) {
			t.Error("expected green blocks for operational provider")
		}
	})

	t.Run("red block for active outage today", func(t *testing.T) {
		p := thirdPartyProvider{
			MonitoringStartDate: nowMs - 60*dayMs,
			Outages: []thirdPartyOutage{
				{Start: nowMs - dayMs/2, End: 0, Status: "active"}, // started 12h ago, ongoing
			},
		}
		raw := buildSparkline(p)
		if !strings.Contains(raw, ansiRed) {
			t.Error("expected red block for active outage")
		}
		if !strings.Contains(raw, ansiGreen) {
			t.Error("expected green blocks for non-outage days")
		}
	})

	t.Run("red block for resolved outage in range", func(t *testing.T) {
		// outage from 5 days ago lasting 1 day
		p := thirdPartyProvider{
			MonitoringStartDate: nowMs - 60*dayMs,
			Outages: []thirdPartyOutage{
				{Start: nowMs - 5*dayMs, End: nowMs - 4*dayMs, Status: "resolved"},
			},
		}
		raw := buildSparkline(p)
		if !strings.Contains(raw, ansiRed) {
			t.Error("expected red block for resolved outage within 30-day window")
		}
	})

	t.Run("dim dots for pre-monitoring period", func(t *testing.T) {
		// monitoring started 10 days ago, so first 20 buckets should be dim dots
		p := thirdPartyProvider{
			MonitoringStartDate: nowMs - 10*dayMs,
			Outages:             []thirdPartyOutage{},
		}
		raw := buildSparkline(p)
		plain := stripANSI(raw)
		dotCount := strings.Count(plain, "·")
		if dotCount < 19 {
			t.Errorf("expected at least 19 dim dots for 10-day monitoring, got %d", dotCount)
		}
		if !strings.Contains(raw, ansiDim) {
			t.Error("expected dim ANSI codes for pre-monitoring period")
		}
	})

	t.Run("outage outside window not shown", func(t *testing.T) {
		// outage was 45 days ago (outside 30-day window)
		p := thirdPartyProvider{
			MonitoringStartDate: nowMs - 60*dayMs,
			Outages: []thirdPartyOutage{
				{Start: nowMs - 45*dayMs, End: nowMs - 44*dayMs, Status: "resolved"},
			},
		}
		raw := buildSparkline(p)
		if strings.Contains(raw, ansiRed) {
			t.Error("expected no red blocks for outage outside 30-day window")
		}
	})
}

func TestFormatThirdPartyTable(t *testing.T) {
	origTimeNow := timeNow
	defer func() { timeNow = origTimeNow }()
	timeNow = func() time.Time { return fixedTime }

	nowMs := fixedTime.UnixMilli()

	providers := []thirdPartyProvider{
		{
			ProviderName:         "aws-s3",
			DisplayName:          "Amazon S3",
			ProviderService:      "S3",
			MonitoringStartDate:  nowMs - 60*dayMs,
			Outages: []thirdPartyOutage{
				{Start: nowMs - dayMs/2, End: 0, Status: "active"},
			},
		},
		{
			ProviderName:         "stripe",
			DisplayName:          "Stripe",
			MonitoringStartDate:  nowMs - 60*dayMs,
			Outages: []thirdPartyOutage{
				{Start: nowMs - 45*dayMs, End: nowMs - 44*dayMs, Status: "resolved"},
			},
		},
	}

	output := formatThirdPartyTable(providers)

	// Verify headers
	for _, header := range []string{"PROVIDER", "DISPLAY NAME", "SERVICE", "UPTIME", "STATUS"} {
		if !strings.Contains(output, header) {
			t.Errorf("expected table to contain header %q", header)
		}
	}

	// Verify data
	if !strings.Contains(output, "aws-s3") {
		t.Error("expected table to contain 'aws-s3'")
	}
	if !strings.Contains(output, "Amazon S3") {
		t.Error("expected table to contain 'Amazon S3'")
	}

	// Verify sparkline contains colored blocks
	if !strings.Contains(output, "█") {
		t.Error("expected table to contain sparkline block characters")
	}

	// aws-s3 has active outage → should show red
	if !strings.Contains(output, ansiRed) {
		t.Error("expected red in sparkline for provider with active outage")
	}

	// Status column
	if !strings.Contains(output, "active") {
		t.Error("expected 'active' status for aws-s3")
	}
	if !strings.Contains(output, "operational") {
		t.Error("expected 'operational' status for stripe (outage outside window)")
	}
}

func TestFormatThirdPartyTable_Empty(t *testing.T) {
	output := formatThirdPartyTable(nil)
	if !strings.Contains(output, "PROVIDER") {
		t.Error("expected empty table to still contain headers")
	}
}

func setupThirdPartyRunTest(t *testing.T, serverURL string) func() {
	t.Helper()
	origClient := httpClient
	origWriter := outputWriter
	origSearch := thirdPartySearch
	origActive := thirdPartyActiveOnly
	origCfg := cfg
	origFormat := outputFormat
	origTimeNow := timeNow

	httpClient = &http.Client{
		Transport: &redirectTransport{target: serverURL},
	}
	cfg = &config.Config{Site: "datadoghq.com"}
	thirdPartySearch = ""
	thirdPartyActiveOnly = false
	outputFormat = "json"
	timeNow = func() time.Time { return fixedTime }

	return func() {
		httpClient = origClient
		outputWriter = origWriter
		thirdPartySearch = origSearch
		thirdPartyActiveOnly = origActive
		cfg = origCfg
		outputFormat = origFormat
		timeNow = origTimeNow
	}
}

func TestRunStatusPagesThirdParty(t *testing.T) {
	server := setupThirdPartyTestServer(t, http.StatusOK, testOutagesJSON)
	defer server.Close()
	cleanup := setupThirdPartyRunTest(t, server.URL)
	defer cleanup()

	var buf bytes.Buffer
	outputWriter = &buf

	err := runStatusPagesThirdParty(statusPagesThirdPartyCmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("expected output, got empty string")
	}
	if !strings.Contains(output, "aws-s3") {
		t.Error("expected output to contain 'aws-s3'")
	}
	if !strings.Contains(output, "stripe") {
		t.Error("expected output to contain 'stripe'")
	}
}

func TestRunStatusPagesThirdParty_WithSearch(t *testing.T) {
	server := setupThirdPartyTestServer(t, http.StatusOK, testOutagesJSON)
	defer server.Close()
	cleanup := setupThirdPartyRunTest(t, server.URL)
	defer cleanup()

	var buf bytes.Buffer
	outputWriter = &buf
	thirdPartySearch = "stripe"

	err := runStatusPagesThirdParty(statusPagesThirdPartyCmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "stripe") {
		t.Error("expected output to contain 'stripe'")
	}
	if strings.Contains(output, "aws-s3") {
		t.Error("expected output NOT to contain 'aws-s3' when searching for stripe")
	}
}

func TestRunStatusPagesThirdParty_SearchByDisplayName(t *testing.T) {
	server := setupThirdPartyTestServer(t, http.StatusOK, testOutagesJSON)
	defer server.Close()
	cleanup := setupThirdPartyRunTest(t, server.URL)
	defer cleanup()

	var buf bytes.Buffer
	outputWriter = &buf
	thirdPartySearch = "Amazon"

	err := runStatusPagesThirdParty(statusPagesThirdPartyCmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "aws-s3") {
		t.Error("expected searching 'Amazon' to match display name 'Amazon S3'")
	}
	if strings.Contains(output, "stripe") {
		t.Error("expected output NOT to contain 'stripe' when searching for Amazon")
	}
}

func TestRunStatusPagesThirdParty_TableFormat(t *testing.T) {
	server := setupThirdPartyTestServer(t, http.StatusOK, testOutagesJSON)
	defer server.Close()
	cleanup := setupThirdPartyRunTest(t, server.URL)
	defer cleanup()

	var buf bytes.Buffer
	outputWriter = &buf
	outputFormat = "table"

	err := runStatusPagesThirdParty(statusPagesThirdPartyCmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "UPTIME") {
		t.Error("expected table header 'UPTIME' in output")
	}
	if !strings.Contains(output, "█") {
		t.Error("expected sparkline block characters in table output")
	}
	if !strings.Contains(output, "active") {
		t.Error("expected 'active' status for aws-s3")
	}
	if !strings.Contains(output, "operational") {
		t.Error("expected 'operational' status for providers without active outages")
	}
}

func TestRunStatusPagesThirdParty_TableFormatEmpty(t *testing.T) {
	server := setupThirdPartyTestServer(t, http.StatusOK, testOutagesJSON)
	defer server.Close()
	cleanup := setupThirdPartyRunTest(t, server.URL)
	defer cleanup()

	var buf bytes.Buffer
	outputWriter = &buf
	outputFormat = "table"
	thirdPartySearch = "nonexistent"

	err := runStatusPagesThirdParty(statusPagesThirdPartyCmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No results found") {
		t.Errorf("expected 'No results found', got: %s", output)
	}
}

func TestRunStatusPagesThirdParty_ServerError(t *testing.T) {
	server := setupThirdPartyTestServer(t, http.StatusInternalServerError, "error")
	defer server.Close()
	cleanup := setupThirdPartyRunTest(t, server.URL)
	defer cleanup()

	var buf bytes.Buffer
	outputWriter = &buf

	err := runStatusPagesThirdParty(statusPagesThirdPartyCmd, []string{})
	if err == nil {
		t.Error("expected error for server error response, got nil")
	}
}

func TestRunStatusPagesThirdParty_InvalidJSON(t *testing.T) {
	server := setupThirdPartyTestServer(t, http.StatusOK, "not json")
	defer server.Close()
	cleanup := setupThirdPartyRunTest(t, server.URL)
	defer cleanup()

	var buf bytes.Buffer
	outputWriter = &buf

	err := runStatusPagesThirdParty(statusPagesThirdPartyCmd, []string{})
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func init() {
	_ = os.Stdout
}
