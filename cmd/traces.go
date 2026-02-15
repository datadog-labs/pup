// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"fmt"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/DataDog/pup/pkg/formatter"
	"github.com/DataDog/pup/pkg/util"
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
		page, ok := meta.GetPageOk()
		if !ok || page == nil {
			break
		}
		cursor, ok := page.GetAfterOk()
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

func runTracesAggregate(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("traces aggregate: not yet implemented")
}
