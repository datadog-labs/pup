# Traces Command Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace the placeholder `traces` command with working `search` and `aggregate` subcommands using the typed `datadogV2.SpansApi`.

**Architecture:** Two subcommands (`search`, `aggregate`) backed by the typed Datadog SpansApi client. Search uses `ListSpans` (POST) with manual cursor-based pagination. Aggregate uses `AggregateSpans` (POST) with `--compute` parsing reusing the existing `parseComputeString` helper from logs. Both use `--from`/`--to` with flexible time parsing via `util.ParseTimeToUnixMilli`.

**Tech Stack:** Go, cobra (CLI), datadog-api-client-go v2.54.0 (`datadogV2.SpansApi`), existing `pkg/formatter`, `pkg/util`, `pkg/client`

**Design doc:** `docs/plans/2026-02-14-traces-command-design.md`

---

### Task 1: Delete placeholder files and create traces.go skeleton

**Files:**
- Delete: `cmd/traces_simple.go`
- Delete: `cmd/traces_simple_test.go`
- Create: `cmd/traces.go`
- Create: `cmd/traces_test.go`

**Context:** `cmd/root.go:190` registers `tracesCmd` via `rootCmd.AddCommand(tracesCmd)`. The new file must export the same `tracesCmd` variable name. The existing placeholder (`cmd/traces_simple.go`) defines `tracesCmd` as a cobra.Command that returns an error. We replace it entirely.

**Step 1: Delete the placeholder files**

```bash
rm cmd/traces_simple.go cmd/traces_simple_test.go
```

**Step 2: Create `cmd/traces.go` with command skeleton (no RunE yet)**

```go
// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Command flags
var (
	tracesQuery string
	tracesFrom  string
	tracesTo    string
	tracesLimit int
	tracesSort  string

	// Aggregate flags
	tracesCompute string
	tracesGroupBy string
)

var tracesCmd = &cobra.Command{
	Use:   "traces",
	Short: "Search and aggregate APM traces",
	Long: `Search and aggregate APM span data for distributed tracing analysis.

The traces command provides access to individual span-level data collected by
Datadog APM. Use it to find specific spans matching a query or compute
aggregated statistics over spans.

COMPLEMENTS THE APM COMMAND:
  • apm: Service-level aggregated data (services, operations, dependencies)
  • traces: Individual span-level data (search, aggregate)

  Use 'pup apm' to identify which services have problems.
  Use 'pup traces' to drill into the actual spans.

SPAN QUERY SYNTAX:
  • service:web-server - Match by service
  • resource_name:"GET /api/users" - Match by resource
  • @http.status_code:500 - Match by tag
  • @duration:>1000000000 - Match by duration (nanoseconds)
  • env:production - Match by environment
  • AND, OR, NOT - Boolean operators

TIME RANGES:
  Supported formats: 1h, 30m, 7d, 5min, 2hours, RFC3339, Unix timestamp, now

EXAMPLES:
  # Search for error spans in the last hour
  pup traces search --query="service:web-server @http.status_code:500"

  # Search with custom time range
  pup traces search --query="env:prod @duration:>1000000000" --from="4h" --limit=100

  # Count spans by service
  pup traces aggregate --query="*" --compute="count" --group-by="service"

  # Average duration by resource
  pup traces aggregate --query="service:web-server" --compute="avg(@duration)" --group-by="resource_name"

  # P99 latency
  pup traces aggregate --query="env:prod" --compute="percentile(@duration, 99)"

AUTHENTICATION:
  Requires either OAuth2 authentication (pup auth login) or API keys
  (DD_API_KEY and DD_APP_KEY environment variables).
  OAuth2 requires the apm_read scope.`,
}

var tracesSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search for spans",
	Long: `Search for individual spans matching a query.

Returns span data including service, resource, duration, tags, and trace IDs.
Results are auto-paginated up to --limit.

FLAGS:
  --query     Span search query (default: "*")
  --from      Start time (default: "1h")
  --to        End time (default: "now")
  --limit     Max spans to return (default: 50)
  --sort      Sort order: timestamp or -timestamp (default: "-timestamp")

EXAMPLES:
  # Find error spans
  pup traces search --query="@http.status_code:>=500"

  # Find slow spans (duration > 1s = 1000000000ns)
  pup traces search --query="service:api @duration:>1000000000" --from="1h"

  # Search with ascending sort
  pup traces search --query="env:prod" --sort="timestamp" --limit=20`,
	RunE: runTracesSearch,
}

var tracesAggregateCmd = &cobra.Command{
	Use:   "aggregate",
	Short: "Compute aggregated stats over spans",
	Long: `Compute aggregated statistics over spans matching a query.

Returns computed metrics (count, avg, sum, percentiles, etc.) optionally
grouped by a facet. Unlike search, this returns statistical buckets, not
individual spans.

FLAGS:
  --query      Span search query (default: "*")
  --from       Start time (default: "1h")
  --to         End time (default: "now")
  --compute    Aggregation to compute (required)
  --group-by   Facet to group results by

COMPUTE FORMATS:
  count                        Count of matching spans
  avg(@duration)               Average of a metric
  sum(@duration)               Sum of a metric
  min(@duration)               Minimum of a metric
  max(@duration)               Maximum of a metric
  median(@duration)            Median of a metric
  cardinality(@user.id)        Unique count of a facet
  percentile(@duration, 99)    Percentile of a metric (converts to pc99)
  percentile(@duration, 95)    Percentile of a metric (converts to pc95)

EXAMPLES:
  # Count all error spans
  pup traces aggregate --query="@http.status_code:>=500" --compute="count"

  # Average duration by service
  pup traces aggregate --query="env:prod" --compute="avg(@duration)" --group-by="service"

  # P99 latency by resource
  pup traces aggregate --query="service:api" --compute="percentile(@duration, 99)" --group-by="resource_name"

  # Count unique users
  pup traces aggregate --query="service:web" --compute="cardinality(@usr.id)"`,
	RunE: runTracesAggregate,
}

func init() {
	// Search flags
	tracesSearchCmd.Flags().StringVar(&tracesQuery, "query", "*", "Span search query")
	tracesSearchCmd.Flags().StringVar(&tracesFrom, "from", "1h", "Start time (e.g., 1h, 30m, 7d, RFC3339, Unix timestamp, now)")
	tracesSearchCmd.Flags().StringVar(&tracesTo, "to", "now", "End time")
	tracesSearchCmd.Flags().IntVar(&tracesLimit, "limit", 50, "Max spans to return")
	tracesSearchCmd.Flags().StringVar(&tracesSort, "sort", "-timestamp", "Sort order: timestamp or -timestamp")

	// Aggregate flags
	tracesAggregateCmd.Flags().StringVar(&tracesQuery, "query", "*", "Span search query")
	tracesAggregateCmd.Flags().StringVar(&tracesFrom, "from", "1h", "Start time (e.g., 1h, 30m, 7d, RFC3339, Unix timestamp, now)")
	tracesAggregateCmd.Flags().StringVar(&tracesTo, "to", "now", "End time")
	tracesAggregateCmd.Flags().StringVar(&tracesCompute, "compute", "", "Aggregation: count, avg(@duration), percentile(@duration, 99), etc.")
	tracesAggregateCmd.Flags().StringVar(&tracesGroupBy, "group-by", "", "Facet to group by (e.g., service, resource_name, @http.status_code)")
	if err := tracesAggregateCmd.MarkFlagRequired("compute"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}

	// Register subcommands
	tracesCmd.AddCommand(tracesSearchCmd)
	tracesCmd.AddCommand(tracesAggregateCmd)
}

func runTracesSearch(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("traces search: not yet implemented")
}

func runTracesAggregate(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("traces aggregate: not yet implemented")
}
```

**Step 3: Create `cmd/traces_test.go` with structure tests**

```go
// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"testing"
)

func TestTracesCmd(t *testing.T) {
	if tracesCmd == nil {
		t.Fatal("tracesCmd is nil")
	}
	if tracesCmd.Use != "traces" {
		t.Errorf("Use = %s, want traces", tracesCmd.Use)
	}
}

func TestTracesSubcommands(t *testing.T) {
	subcommands := map[string]bool{
		"search":    false,
		"aggregate": false,
	}
	for _, cmd := range tracesCmd.Commands() {
		if _, ok := subcommands[cmd.Name()]; ok {
			subcommands[cmd.Name()] = true
		}
	}
	for name, found := range subcommands {
		if !found {
			t.Errorf("missing subcommand: %s", name)
		}
	}
}

func TestTracesSearchFlags(t *testing.T) {
	flags := tracesSearchCmd.Flags()

	tests := []struct {
		name         string
		defaultValue string
	}{
		{"query", "*"},
		{"from", "1h"},
		{"to", "now"},
		{"sort", "-timestamp"},
	}

	for _, tt := range tests {
		f := flags.Lookup(tt.name)
		if f == nil {
			t.Errorf("search command missing --%s flag", tt.name)
			continue
		}
		if f.DefValue != tt.defaultValue {
			t.Errorf("--%s default = %q, want %q", tt.name, f.DefValue, tt.defaultValue)
		}
	}

	// Check limit flag separately (int type)
	limitFlag := flags.Lookup("limit")
	if limitFlag == nil {
		t.Error("search command missing --limit flag")
	} else if limitFlag.DefValue != "50" {
		t.Errorf("--limit default = %q, want %q", limitFlag.DefValue, "50")
	}
}

func TestTracesAggregateFlags(t *testing.T) {
	flags := tracesAggregateCmd.Flags()

	tests := []struct {
		name         string
		defaultValue string
	}{
		{"query", "*"},
		{"from", "1h"},
		{"to", "now"},
		{"compute", ""},
		{"group-by", ""},
	}

	for _, tt := range tests {
		f := flags.Lookup(tt.name)
		if f == nil {
			t.Errorf("aggregate command missing --%s flag", tt.name)
			continue
		}
		if f.DefValue != tt.defaultValue {
			t.Errorf("--%s default = %q, want %q", tt.name, f.DefValue, tt.defaultValue)
		}
	}
}
```

**Step 4: Verify build and tests pass**

```bash
go build ./...
go test ./cmd/ -run TestTraces -v
```

Expected: All 4 tests pass (TestTracesCmd, TestTracesSubcommands, TestTracesSearchFlags, TestTracesAggregateFlags). The `runTracesSearch` and `runTracesAggregate` stubs return errors but aren't called by structure tests.

**Step 5: Commit**

```bash
git add cmd/traces.go cmd/traces_test.go
git rm cmd/traces_simple.go cmd/traces_simple_test.go
git commit -m "feat(traces): replace placeholder with command skeleton

Add tracesCmd with search and aggregate subcommands, flags, and
structure tests. RunE functions are stubs that will be implemented
in subsequent commits.

Closes #49"
```

---

### Task 2: Implement traces search

**Files:**
- Modify: `cmd/traces.go` — replace `runTracesSearch` stub

**Context:** The search implementation follows the same pattern as `runLogsSearch` in `cmd/logs_simple.go:729-879`. Key differences: (1) uses `SpansApi` instead of `LogsApi`, (2) uses `SpansSort` instead of `LogsSort`, (3) Spans API uses `SpansQueryFilter` with `From`/`To` as ISO8601/unix-ms strings (same as logs). The `parseComputeString` helper in `logs_simple.go:659` is already exported at package level and can be reused.

**Step 1: Write the failing test for search execution**

Add to `cmd/traces_test.go`:

```go
import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/DataDog/pup/pkg/client"
	"github.com/DataDog/pup/pkg/config"
)

func setupTracesTestClient(t *testing.T) func() {
	t.Helper()

	origClient := ddClient
	origCfg := cfg
	origFactory := clientFactory

	cfg = &config.Config{
		Site:   "datadoghq.com",
		APIKey: "test-api-key-12345678",
		AppKey: "test-app-key-12345678",
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

func TestRunTracesSearch(t *testing.T) {
	cleanup := setupTracesTestClient(t)
	defer cleanup()

	tests := []struct {
		name    string
		query   string
		from    string
		to      string
		limit   int
		sort    string
		wantErr bool
	}{
		{
			name:    "valid query returns error from mock client",
			query:   "service:web-server",
			from:    "1h",
			to:      "now",
			limit:   50,
			sort:    "-timestamp",
			wantErr: true,
		},
		{
			name:    "invalid from time",
			query:   "*",
			from:    "invalid-time",
			to:      "now",
			limit:   50,
			sort:    "-timestamp",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracesQuery = tt.query
			tracesFrom = tt.from
			tracesTo = tt.to
			tracesLimit = tt.limit
			tracesSort = tt.sort

			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			err := runTracesSearch(tracesSearchCmd, []string{})

			if (err != nil) != tt.wantErr {
				t.Errorf("runTracesSearch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
```

**Step 2: Run the test to verify it fails**

```bash
go test ./cmd/ -run TestRunTracesSearch -v
```

Expected: FAIL — the stub returns a generic "not yet implemented" error, but the "invalid from time" test case should get a different error message once implemented. The test structure works.

**Step 3: Implement `runTracesSearch` in `cmd/traces.go`**

Replace the stub with:

```go
import (
	"fmt"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/DataDog/pup/pkg/formatter"
	"github.com/DataDog/pup/pkg/util"
	"github.com/spf13/cobra"
)

func runTracesSearch(cmd *cobra.Command, args []string) error {
	fromTime, err := util.ParseTimeToUnixMilli(tracesFrom)
	if err != nil {
		return fmt.Errorf("invalid --from time: %w", err)
	}

	toTime, err := util.ParseTimeToUnixMilli(tracesTo)
	if err != nil {
		return fmt.Errorf("invalid --to time: %w", err)
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewSpansApi(client.V2())

	query := tracesQuery
	from := fmt.Sprintf("%d", fromTime)
	to := fmt.Sprintf("%d", toTime)
	pageLimit := int32(tracesLimit)
	if pageLimit > 1000 {
		pageLimit = 1000
	}
	sort := datadogV2.SpansSort(tracesSort)

	body := datadogV2.SpansListRequest{
		Data: &datadogV2.SpansListRequestData{
			Attributes: &datadogV2.SpansListRequestAttributes{
				Filter: &datadogV2.SpansQueryFilter{
					Query: &query,
					From:  &from,
					To:    &to,
				},
				Page: &datadogV2.SpansListRequestPage{
					Limit: &pageLimit,
				},
				Sort: &sort,
			},
			Type: datadogV2.SPANSLISTREQUESTTYPE_SEARCH_REQUEST.Ptr(),
		},
	}

	// Fetch first page
	resp, r, err := api.ListSpans(client.Context(), body)
	if err != nil {
		if r != nil {
			apiBody := extractAPIErrorBody(err)
			if apiBody != "" {
				fromTimeObj := time.UnixMilli(fromTime).UTC()
				toTimeObj := time.UnixMilli(toTime).UTC()
				return fmt.Errorf("failed to search spans: %w\nStatus: %d\nAPI Response: %s\n\nRequest Details:\n- Query: %s\n- From: %s UTC (parsed from: %s)\n- To: %s UTC (parsed from: %s)\n- Limit: %d\n\nTroubleshooting:\n- Verify your query follows span search syntax\n- Check that your time range is valid\n- Ensure you have the apm_read OAuth scope or valid API keys",
					err, r.StatusCode, apiBody,
					tracesQuery,
					fromTimeObj.Format(time.RFC3339), tracesFrom,
					toTimeObj.Format(time.RFC3339), tracesTo,
					tracesLimit)
			}
			return fmt.Errorf("failed to search spans: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to search spans: %w", err)
	}

	// Collect spans with pagination
	allSpans := resp.GetData()
	pageCount := 1

	for tracesLimit > 0 && len(allSpans) < tracesLimit {
		meta, ok := resp.GetMetaOk()
		if !ok || meta == nil {
			break
		}
		cursor, ok := meta.GetAfterOk()
		if !ok || cursor == nil || *cursor == "" {
			break
		}

		remaining := tracesLimit - len(allSpans)
		if remaining <= 0 {
			break
		}
		remainingLimit := int32(remaining)
		if remainingLimit > 1000 {
			remainingLimit = 1000
		}
		body.Data.Attributes.Page.Limit = &remainingLimit
		body.Data.Attributes.Page.Cursor = cursor

		resp, r, err = api.ListSpans(client.Context(), body)
		if err != nil {
			printOutput("Warning: Failed to fetch page %d: %v\n", pageCount+1, err)
			break
		}

		allSpans = append(allSpans, resp.GetData()...)
		pageCount++
	}

	if tracesLimit > 0 && len(allSpans) > tracesLimit {
		allSpans = allSpans[:tracesLimit]
	}

	if len(allSpans) == 0 {
		printOutput("No spans found matching your query.\n\n")
		printOutput("Tips:\n")
		printOutput("- Try a broader time range (e.g., --from=\"24h\")\n")
		printOutput("- Verify the service name exists in your traces\n")
		printOutput("- Check your query syntax: https://docs.datadoghq.com/tracing/trace_explorer/search/\n")
		printOutput("- Try a simpler query like --query=\"*\" to see any spans\n")
		return nil
	}

	finalResp := resp
	if pageCount > 1 {
		finalResp.SetData(allSpans)
		printOutput("Fetched %d spans across %d pages\n\n", len(allSpans), pageCount)
	}

	output, err := formatter.FormatOutput(finalResp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}

	printOutput("%s\n", output)
	return nil
}
```

**Step 4: Verify build and tests pass**

```bash
go build ./...
go test ./cmd/ -run TestRunTracesSearch -v
```

Expected: Both test cases pass. "valid query" fails at the mock client (wantErr: true). "invalid from time" fails at time parsing (wantErr: true).

**Step 5: Commit**

```bash
git add cmd/traces.go cmd/traces_test.go
git commit -m "feat(traces): implement search subcommand

Uses typed SpansApi.ListSpans with manual cursor-based pagination.
Supports --query, --from, --to, --limit, --sort flags with flexible
time parsing. Includes detailed error messages with troubleshooting
hints."
```

---

### Task 3: Implement traces aggregate

**Files:**
- Modify: `cmd/traces.go` — replace `runTracesAggregate` stub

**Context:** Follows the pattern from `runLogsAggregate` in `cmd/logs_simple.go:1025-1122`. Reuses the existing `parseComputeString` helper (defined in `cmd/logs_simple.go:659`) which is package-level and accessible from `cmd/traces.go`. The SpansApi aggregate types mirror LogsApi aggregate types: `SpansCompute` ↔ `LogsCompute`, `SpansGroupBy` ↔ `LogsGroupBy`, `SpansAggregationFunction` ↔ `LogsAggregationFunction`.

**Step 1: Write the failing test for aggregate execution**

Add to `cmd/traces_test.go`:

```go
func TestRunTracesAggregate(t *testing.T) {
	cleanup := setupTracesTestClient(t)
	defer cleanup()

	tests := []struct {
		name    string
		query   string
		from    string
		to      string
		compute string
		groupBy string
		wantErr bool
	}{
		{
			name:    "count aggregation returns error from mock client",
			query:   "service:web-server",
			from:    "1h",
			to:      "now",
			compute: "count",
			wantErr: true,
		},
		{
			name:    "avg with metric returns error from mock client",
			query:   "*",
			from:    "1h",
			to:      "now",
			compute: "avg(@duration)",
			groupBy: "service",
			wantErr: true,
		},
		{
			name:    "invalid compute format",
			query:   "*",
			from:    "1h",
			to:      "now",
			compute: "",
			wantErr: true,
		},
		{
			name:    "invalid from time",
			query:   "*",
			from:    "invalid-time",
			to:      "now",
			compute: "count",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracesQuery = tt.query
			tracesFrom = tt.from
			tracesTo = tt.to
			tracesCompute = tt.compute
			tracesGroupBy = tt.groupBy

			var buf bytes.Buffer
			outputWriter = &buf
			defer func() { outputWriter = os.Stdout }()

			err := runTracesAggregate(tracesAggregateCmd, []string{})

			if (err != nil) != tt.wantErr {
				t.Errorf("runTracesAggregate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
```

**Step 2: Run the test to verify it fails**

```bash
go test ./cmd/ -run TestRunTracesAggregate -v
```

Expected: FAIL — the stub returns a generic error for all cases, but "invalid compute format" and "invalid from time" should produce specific errors once implemented.

**Step 3: Implement `runTracesAggregate` in `cmd/traces.go`**

Replace the stub with:

```go
func runTracesAggregate(cmd *cobra.Command, args []string) error {
	// Parse compute string (reuse parseComputeString from logs_simple.go)
	aggregation, metric, err := parseComputeString(tracesCompute)
	if err != nil {
		return fmt.Errorf("invalid --compute value: %w", err)
	}

	fromTime, err := util.ParseTimeToUnixMilli(tracesFrom)
	if err != nil {
		return fmt.Errorf("invalid --from time: %w", err)
	}

	toTime, err := util.ParseTimeToUnixMilli(tracesTo)
	if err != nil {
		return fmt.Errorf("invalid --to time: %w", err)
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewSpansApi(client.V2())

	// Build compute
	compute := datadogV2.SpansCompute{
		Aggregation: datadogV2.SpansAggregationFunction(aggregation),
	}
	if metric != "" {
		compute.Metric = &metric
	}

	query := tracesQuery
	from := fmt.Sprintf("%d", fromTime)
	to := fmt.Sprintf("%d", toTime)

	body := datadogV2.SpansAggregateRequest{
		Data: &datadogV2.SpansAggregateData{
			Attributes: &datadogV2.SpansAggregateRequestAttributes{
				Compute: []datadogV2.SpansCompute{compute},
				Filter: &datadogV2.SpansQueryFilter{
					Query: &query,
					From:  &from,
					To:    &to,
				},
			},
			Type: datadogV2.SPANSAGGREGATEREQUESTTYPE_AGGREGATE_REQUEST.Ptr(),
		},
	}

	// Add group by if specified
	if tracesGroupBy != "" {
		body.Data.Attributes.GroupBy = []datadogV2.SpansGroupBy{
			{
				Facet: tracesGroupBy,
			},
		}
	}

	resp, r, err := api.AggregateSpans(client.Context(), body)
	if err != nil {
		if r != nil {
			apiBody := extractAPIErrorBody(err)
			if apiBody != "" {
				fromTimeObj := time.UnixMilli(fromTime).UTC()
				toTimeObj := time.UnixMilli(toTime).UTC()
				return fmt.Errorf("failed to aggregate spans: %w\nStatus: %d\nAPI Response: %s\n\nRequest Details:\n- Query: %s\n- Compute: %s (parsed as: aggregation=%q, metric=%q)\n- Group By: %s\n- From: %s UTC (parsed from: %s)\n- To: %s UTC (parsed from: %s)\n\nTroubleshooting:\n- Verify the aggregation function is supported\n- Ensure the metric field exists in your spans (e.g., @duration)\n- Check your query syntax\n- Ensure you have the apm_read OAuth scope or valid API keys",
					err, r.StatusCode, apiBody,
					tracesQuery,
					tracesCompute, aggregation, metric,
					tracesGroupBy,
					fromTimeObj.Format(time.RFC3339), tracesFrom,
					toTimeObj.Format(time.RFC3339), tracesTo)
			}
			return fmt.Errorf("failed to aggregate spans: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to aggregate spans: %w", err)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}

	printOutput("%s\n", output)
	return nil
}
```

**Step 4: Verify build and tests pass**

```bash
go build ./...
go test ./cmd/ -run TestRunTracesAggregate -v
```

Expected: All 4 test cases pass. "count aggregation" and "avg with metric" fail at mock client (wantErr: true). "invalid compute format" fails at parseComputeString. "invalid from time" fails at time parsing.

**Step 5: Commit**

```bash
git add cmd/traces.go cmd/traces_test.go
git commit -m "feat(traces): implement aggregate subcommand

Uses typed SpansApi.AggregateSpans with --compute parsing reusing
the existing parseComputeString helper. Supports grouping by any
span facet via --group-by."
```

---

### Task 4: Run full test suite and verify coverage

**Files:**
- No changes — verification only

**Context:** CI requires >80% coverage and all tests passing with race detection. Run the full test suite to catch any regressions from the traces command changes.

**Step 1: Run full test suite with race detection**

```bash
go test -race ./...
```

Expected: All tests pass including the new traces tests and all existing tests.

**Step 2: Check coverage for cmd package**

```bash
go test ./cmd/ -coverprofile=coverage.out -covermode=atomic
go tool cover -func=coverage.out | grep traces
```

Expected: `cmd/traces.go` coverage >80%.

**Step 3: Run linter**

```bash
golangci-lint run ./...
```

Expected: No lint errors in `cmd/traces.go` or `cmd/traces_test.go`.

**Step 4: Commit if any fixes were needed**

If linting or tests required fixes, commit them:

```bash
git add cmd/traces.go cmd/traces_test.go
git commit -m "fix(traces): address lint and test issues"
```

---

### Task 5: Update documentation

**Files:**
- Modify: `docs/COMMANDS.md:25` — update traces row
- Modify: `docs/COMMANDS.md:111` — update traces description in domain categories

**Context:** `docs/COMMANDS.md` currently shows traces as `| traces | - | cmd/traces_simple.go | ❌ |` and the domain categories section says "(not yet implemented - use apm commands instead)".

**Step 1: Update COMMANDS.md**

In the command index table (line 25), change:
```
| traces | - | cmd/traces_simple.go | ❌ |
```
to:
```
| traces | search, aggregate | cmd/traces.go | ✅ |
```

In the domain categories section (line 111), change:
```
- **traces** - APM traces (not yet implemented - use `apm` commands instead)
```
to:
```
- **traces** - APM span search and aggregation (search, aggregate)
```

Update the summary line (line 61) to reflect one more working command.

**Step 2: Verify no other stale references**

```bash
grep -r "traces_simple" docs/ cmd/root.go
grep -r "traces.*under development" .
```

Expected: No matches (all references to the old placeholder are gone).

**Step 3: Commit**

```bash
git add docs/COMMANDS.md
git commit -m "docs(traces): update command reference for traces implementation

Update traces status from placeholder to working. Add search and
aggregate subcommands to command index. Update domain categories
description."
```

---

### Task 6: Create PR

**Files:** None — git operations only

**Step 1: Push branch and create PR**

```bash
gh pr create \
  --title "feat(traces): implement search and aggregate subcommands" \
  --body "$(cat <<'EOF'
## Summary
Replace the placeholder traces command with working `search` and `aggregate` subcommands using the typed `datadogV2.SpansApi`.

Closes #49

## Changes
- Delete `cmd/traces_simple.go` placeholder (returned "under development" error)
- Create `cmd/traces.go` with two subcommands:
  - `traces search` — find individual spans with auto-pagination (SpansApi.ListSpans)
  - `traces aggregate` — compute stats over spans (SpansApi.AggregateSpans)
- Create `cmd/traces_test.go` with structure, flag, and execution tests
- Update `docs/COMMANDS.md` — traces status ❌ → ✅

## Design
See `docs/plans/2026-02-14-traces-command-design.md` for full design rationale.

Key decisions:
- **Two subcommands** (search + aggregate) — one for individual spans, one for computed stats
- **Typed API client** — uses `datadogV2.SpansApi` (not RawRequest) for type safety
- **`--from`/`--to` with flexible parsing** — consistent with logs, metrics, rum (17+ subcommands use this pattern)
- **Auto-pagination** for search — agents get complete results up to `--limit`
- **Reuses `parseComputeString`** from logs for `--compute` flag parsing

## Testing
- Command structure tests (Use, Short, subcommand registration)
- Flag existence, types, and defaults
- Search execution with mock client
- Aggregate execution with mock client (including compute parsing, group-by)
- All existing tests pass with `go test -race ./...`

---
🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```
