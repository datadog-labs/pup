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
