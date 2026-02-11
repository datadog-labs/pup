// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/DataDog/pup/pkg/formatter"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Search and analyze logs",
	Long: `Search and analyze log data with flexible queries and time ranges.

The logs command provides comprehensive access to Datadog's log management capabilities
including search, querying, aggregation, archives management, custom destinations,
log-based metrics, and restriction queries.

CAPABILITIES:
  • Search logs with flexible queries (v1 API)
  • Query and aggregate logs (v2 API)
  • List logs with filtering (v2 API)
  • Manage log archives (CRUD operations)
  • Manage custom destinations for logs
  • Create and manage log-based metrics
  • Configure restriction queries for access control

LOG QUERY SYNTAX:
  Logs use a query language similar to web search:
  • status:error - Match by status
  • service:web-app - Match by service
  • @user.id:12345 - Match by attribute
  • host:i-* - Wildcard matching
  • "exact phrase" - Exact phrase matching
  • AND, OR, NOT - Boolean operators

TIME RANGES:
  Supported time formats:
  • Relative: 1h, 30m, 7d, 1w (hour, minute, day, week)
  • Absolute: Unix timestamp in milliseconds
  • now: Current time

EXAMPLES:
  # Search for error logs in the last hour
  pup logs search --query="status:error" --from="1h"

  # Query logs from a specific service
  pup logs query --query="service:web-app" --from="4h" --to="now"

  # Aggregate logs by status
  pup logs aggregate --query="*" --compute="count" --group-by="status"

  # List log archives
  pup logs archives list

  # Get specific archive details
  pup logs archives get "my-archive-id"

  # List log-based metrics
  pup logs metrics list

  # Create a log-based metric
  pup logs metrics create --name="error.count" --query="status:error"

  # List custom destinations
  pup logs custom-destinations list

  # List restriction queries
  pup logs restriction-queries list

AUTHENTICATION:
  Requires either OAuth2 authentication (pup auth login) or API keys
  (DD_API_KEY and DD_APP_KEY environment variables).`,
}

// V1 Logs API Commands (logs.yaml)

var logsSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search logs (v1 API)",
	Long: `Search logs using the v1 Logs API with flexible query syntax.

This command provides access to historical log data using Datadog's search query
language. Results are returned in reverse chronological order (newest first).

QUERY SYNTAX:
  • Basic: status:error
  • Service: service:web-app
  • Attributes: @user.id:12345
  • Tags: env:production
  • Wildcards: host:i-*
  • Boolean: status:error AND service:web-app
  • Negation: -status:info

TIME PARAMETERS:
  --from    Start time (required)
            • Relative: 1h, 30m, 7d (ago from now)
            • Absolute: Unix timestamp in milliseconds
  --to      End time (default: now)
            • Same format as --from
            • Must be after --from

OPTIONS:
  --limit   Maximum number of logs to return (default: 50, max: 1000)
  --sort    Sort order: asc or desc (default: desc)
  --index   Comma-separated list of log indexes to search

EXAMPLES:
  # Search for errors in the last hour
  pup logs search --query="status:error" --from="1h"

  # Search specific service with time range
  pup logs search --query="service:api" --from="2h" --to="1h"

  # Search with attributes and limit
  pup logs search --query="@http.status_code:500" --from="30m" --limit=100

  # Search multiple conditions
  pup logs search --query="status:error AND service:web" --from="4h"

  # Search in specific indexes
  pup logs search --query="*" --from="1h" --index="main,retention"

OUTPUT:
  Returns an array of log events with:
  • id: Log event ID
  • content: Log message/content
  • timestamp: Event timestamp
  • attributes: Log attributes (tags, metadata)
  • service: Service name
  • host: Host identifier`,
	RunE: runLogsSearch,
}

var logsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List logs (v2 API)",
	Long: `List logs using the v2 Logs API with advanced filtering.

This command provides access to log data with more advanced filtering and
pagination capabilities compared to the v1 search API.

FILTERS:
  --query   Log query using search syntax
  --from    Start time (required)
  --to      End time (default: now)
  --limit   Number of logs to return (default: 10)
  --sort    Sort order: timestamp, -timestamp (default: -timestamp)

EXAMPLES:
  # List recent logs
  pup logs list --from="1h"

  # List logs with query filter
  pup logs list --query="service:api" --from="2h" --limit=50

  # List logs sorted by timestamp ascending
  pup logs list --query="*" --from="30m" --sort="timestamp"`,
	RunE: runLogsList,
}

// V2 Logs API Commands

var logsQueryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query logs (v2 API)",
	Long: `Query logs using the v2 Logs API with advanced capabilities.

This is the recommended modern API for querying logs with better performance
and more features than the v1 search API.

OPTIONS:
  --query   Log query (required)
  --from    Start time (required)
  --to      End time (default: now)
  --limit   Maximum results (default: 50)
  --sort    Sort order: timestamp, -timestamp
  --timezone Timezone for timestamps (e.g., "America/New_York")

EXAMPLES:
  # Query recent errors
  pup logs query --query="status:error" --from="1h"

  # Query with specific timezone
  pup logs query --query="service:web" --from="4h" --timezone="America/New_York"

  # Query with custom sort
  pup logs query --query="@user.action:login" --from="1d" --sort="timestamp"`,
	RunE: runLogsQuery,
}

var logsAggregateCmd = &cobra.Command{
	Use:   "aggregate",
	Short: "Aggregate logs (v2 API)",
	Long: `Aggregate logs with grouping and metrics computation.

Perform statistical analysis on log data by grouping and computing metrics.
This is useful for understanding log patterns, volumes, and distributions.

AGGREGATION OPTIONS:
  --query     Log query to filter data (required)
  --from      Start time (required)
  --to        End time (default: now)
  --compute   Metric to compute (count, cardinality, percentile, etc.)
  --group-by  Field to group by (e.g., "status", "service", "@http.status_code")
  --limit     Maximum number of groups (default: 10)

COMPUTE METRICS:
  • count: Count of logs
  • cardinality(@field): Unique values of a field
  • avg(@field): Average value
  • sum(@field): Sum of values
  • min(@field): Minimum value
  • max(@field): Maximum value
  • percentile(@field, 99): Percentile calculation

EXAMPLES:
  # Count logs by status
  pup logs aggregate --query="*" --from="1h" --compute="count" --group-by="status"

  # Count unique users
  pup logs aggregate --query="service:web" --from="4h" --compute="cardinality(@user.id)"

  # Average response time by service
  pup logs aggregate --query="*" --from="1h" --compute="avg(@duration)" --group-by="service"

  # 99th percentile latency
  pup logs aggregate --query="service:api" --from="2h" --compute="percentile(@duration, 99)"

  # Error rate by HTTP status code
  pup logs aggregate --query="status:error" --from="1d" --compute="count" --group-by="@http.status_code"`,
	RunE: runLogsAggregate,
}

// Logs Archives Commands (logs_archives.yaml)

var logsArchivesCmd = &cobra.Command{
	Use:   "archives",
	Short: "Manage log archives",
	Long: `Manage log archives for long-term storage.

Log archives allow you to store logs in external storage (S3, GCS, Azure)
for compliance, auditing, and cost optimization. Archives can be rehydrated
back into Datadog for analysis.

CAPABILITIES:
  • List all log archives
  • Get archive details
  • Create new archives
  • Update archive configuration
  • Delete archives
  • Manage archive ordering

STORAGE DESTINATIONS:
  • AWS S3 buckets
  • Google Cloud Storage
  • Azure Blob Storage

EXAMPLES:
  # List all archives
  pup logs archives list

  # Get specific archive
  pup logs archives get "my-archive-id"

  # Delete archive
  pup logs archives delete "my-archive-id"`,
}

var logsArchivesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all log archives",
	Long: `List all configured log archives.

Returns details about all log archives including their storage destinations,
query filters, and rehydration settings.

OUTPUT:
  • archive_id: Unique archive identifier
  • name: Archive name
  • query: Log query filter for archive
  • destination: Storage destination details
  • state: Archive state (active, paused)
  • rehydration_max_scan_size_in_gb: Max rehydration size

EXAMPLES:
  # List all archives
  pup logs archives list

  # List and filter with jq
  pup logs archives list | jq '.data[] | select(.attributes.state == "active")'`,
	RunE: runLogsArchivesList,
}

var logsArchivesGetCmd = &cobra.Command{
	Use:   "get [archive-id]",
	Short: "Get log archive details",
	Long: `Get detailed information about a specific log archive.

ARGUMENTS:
  archive-id    The unique identifier of the archive

EXAMPLES:
  # Get archive details
  pup logs archives get "my-archive-id"

  # Save archive config to file
  pup logs archives get "my-archive-id" > archive-config.json`,
	Args: cobra.ExactArgs(1),
	RunE: runLogsArchivesGet,
}

var logsArchivesDeleteCmd = &cobra.Command{
	Use:   "delete [archive-id]",
	Short: "Delete a log archive",
	Long: `Delete a log archive configuration.

WARNING: This removes the archive configuration from Datadog. It does not
delete the archived data from the storage destination.

ARGUMENTS:
  archive-id    The unique identifier of the archive to delete

FLAGS:
  --yes, -y    Skip confirmation prompt

EXAMPLES:
  # Delete with confirmation
  pup logs archives delete "my-archive-id"

  # Delete without confirmation
  pup logs archives delete "my-archive-id" --yes`,
	Args: cobra.ExactArgs(1),
	RunE: runLogsArchivesDelete,
}

// Custom Destinations Commands (logs_custom_destinations.yaml)

var logsCustomDestinationsCmd = &cobra.Command{
	Use:   "custom-destinations",
	Short: "Manage custom log destinations",
	Long: `Manage custom destinations for forwarding logs.

Custom destinations allow you to forward logs to external systems in real-time
for processing, storage, or integration with third-party tools.

DESTINATION TYPES:
  • HTTP endpoints
  • Splunk
  • Elasticsearch
  • Custom integrations

EXAMPLES:
  # List all custom destinations
  pup logs custom-destinations list

  # Get destination details
  pup logs custom-destinations get "destination-id"`,
}

var logsCustomDestinationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List custom log destinations",
	Long: `List all configured custom log destinations.

OUTPUT:
  • id: Destination identifier
  • name: Destination name
  • type: Destination type (http, splunk, etc.)
  • enabled: Whether destination is active
  • query: Log query filter

EXAMPLES:
  # List all destinations
  pup logs custom-destinations list`,
	RunE: runLogsCustomDestinationsList,
}

var logsCustomDestinationsGetCmd = &cobra.Command{
	Use:   "get [destination-id]",
	Short: "Get custom destination details",
	Args:  cobra.ExactArgs(1),
	RunE:  runLogsCustomDestinationsGet,
}

// Logs Metrics Commands (logs_metrics.yaml)

var logsMetricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Manage log-based metrics",
	Long: `Manage log-based metrics for long-term trending and alerting.

Log-based metrics convert log data into metrics for:
  • Long-term storage and trending (15 months)
  • Efficient alerting and monitoring
  • Dashboard visualization
  • Cost optimization (metrics are cheaper than logs)

METRIC TYPES:
  • Count: Number of logs matching a query
  • Distribution: Statistical distribution of a numeric field

EXAMPLES:
  # List all log-based metrics
  pup logs metrics list

  # Get metric details
  pup logs metrics get "error.count"

  # Delete a metric
  pup logs metrics delete "error.count"`,
}

var logsMetricsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List log-based metrics",
	Long: `List all configured log-based metrics.

OUTPUT:
  • id: Metric identifier
  • name: Metric name
  • type: count or distribution
  • query: Log query filter
  • group_by: Grouping dimensions
  • compute: Aggregation field (for distribution metrics)

EXAMPLES:
  # List all metrics
  pup logs metrics list

  # Filter active metrics
  pup logs metrics list | jq '.data[] | select(.attributes.is_active == true)'`,
	RunE: runLogsMetricsList,
}

var logsMetricsGetCmd = &cobra.Command{
	Use:   "get [metric-id]",
	Short: "Get log-based metric details",
	Args:  cobra.ExactArgs(1),
	RunE:  runLogsMetricsGet,
}

var logsMetricsDeleteCmd = &cobra.Command{
	Use:   "delete [metric-id]",
	Short: "Delete a log-based metric",
	Args:  cobra.ExactArgs(1),
	RunE:  runLogsMetricsDelete,
}

// Restriction Queries Commands (logs_restriction_queries.yaml)

var logsRestrictionQueriesCmd = &cobra.Command{
	Use:   "restriction-queries",
	Short: "Manage log restriction queries",
	Long: `Manage restriction queries for log access control.

Restriction queries control which logs users and roles can access based on
query filters. This enables fine-grained access control for sensitive data.

USE CASES:
  • Limit access to production logs
  • Restrict PII/sensitive data access
  • Enforce compliance requirements
  • Multi-tenant log isolation

EXAMPLES:
  # List all restriction queries
  pup logs restriction-queries list

  # Get restriction query details
  pup logs restriction-queries get "query-id"`,
}

var logsRestrictionQueriesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List restriction queries",
	RunE:  runLogsRestrictionQueriesList,
}

var logsRestrictionQueriesGetCmd = &cobra.Command{
	Use:   "get [query-id]",
	Short: "Get restriction query details",
	Args:  cobra.ExactArgs(1),
	RunE:  runLogsRestrictionQueriesGet,
}

// Command flags

var (
	// Common flags
	logsQuery    string
	logsFrom     string
	logsTo       string
	logsLimit    int
	logsSort     string
	logsIndex    string
	logsTimezone string

	// Aggregate flags
	logsCompute string
	logsGroupBy string
)

func init() {
	// Search command flags (v1)
	logsSearchCmd.Flags().StringVar(&logsQuery, "query", "", "Search query (required)")
	logsSearchCmd.Flags().StringVar(&logsFrom, "from", "", "Start time: 1h, 30m, 7d, or timestamp (required)")
	logsSearchCmd.Flags().StringVar(&logsTo, "to", "now", "End time: 1h, 30m, now, or timestamp")
	logsSearchCmd.Flags().IntVar(&logsLimit, "limit", 50, "Maximum number of logs (1-1000)")
	logsSearchCmd.Flags().StringVar(&logsSort, "sort", "desc", "Sort order: asc or desc")
	logsSearchCmd.Flags().StringVar(&logsIndex, "index", "", "Comma-separated log indexes")
	if err := logsSearchCmd.MarkFlagRequired("query"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}
	if err := logsSearchCmd.MarkFlagRequired("from"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}

	// List command flags (v2)
	logsListCmd.Flags().StringVar(&logsQuery, "query", "*", "Search query")
	logsListCmd.Flags().StringVar(&logsFrom, "from", "", "Start time (required)")
	logsListCmd.Flags().StringVar(&logsTo, "to", "now", "End time")
	logsListCmd.Flags().IntVar(&logsLimit, "limit", 10, "Number of logs")
	logsListCmd.Flags().StringVar(&logsSort, "sort", "-timestamp", "Sort order")
	if err := logsListCmd.MarkFlagRequired("from"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}

	// Query command flags (v2)
	logsQueryCmd.Flags().StringVar(&logsQuery, "query", "", "Log query (required)")
	logsQueryCmd.Flags().StringVar(&logsFrom, "from", "", "Start time (required)")
	logsQueryCmd.Flags().StringVar(&logsTo, "to", "now", "End time")
	logsQueryCmd.Flags().IntVar(&logsLimit, "limit", 50, "Maximum results")
	logsQueryCmd.Flags().StringVar(&logsSort, "sort", "-timestamp", "Sort order")
	logsQueryCmd.Flags().StringVar(&logsTimezone, "timezone", "", "Timezone for timestamps")
	if err := logsQueryCmd.MarkFlagRequired("query"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}
	if err := logsQueryCmd.MarkFlagRequired("from"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}

	// Aggregate command flags (v2)
	logsAggregateCmd.Flags().StringVar(&logsQuery, "query", "", "Log query (required)")
	logsAggregateCmd.Flags().StringVar(&logsFrom, "from", "", "Start time (required)")
	logsAggregateCmd.Flags().StringVar(&logsTo, "to", "now", "End time")
	logsAggregateCmd.Flags().StringVar(&logsCompute, "compute", "count", "Metric to compute")
	logsAggregateCmd.Flags().StringVar(&logsGroupBy, "group-by", "", "Field to group by")
	logsAggregateCmd.Flags().IntVar(&logsLimit, "limit", 10, "Maximum groups")
	if err := logsAggregateCmd.MarkFlagRequired("query"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}
	if err := logsAggregateCmd.MarkFlagRequired("from"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}

	// Add subcommands
	logsCmd.AddCommand(logsSearchCmd)
	logsCmd.AddCommand(logsListCmd)
	logsCmd.AddCommand(logsQueryCmd)
	logsCmd.AddCommand(logsAggregateCmd)

	// Archives subcommands
	logsArchivesCmd.AddCommand(logsArchivesListCmd)
	logsArchivesCmd.AddCommand(logsArchivesGetCmd)
	logsArchivesCmd.AddCommand(logsArchivesDeleteCmd)
	logsCmd.AddCommand(logsArchivesCmd)

	// Custom destinations subcommands
	logsCustomDestinationsCmd.AddCommand(logsCustomDestinationsListCmd)
	logsCustomDestinationsCmd.AddCommand(logsCustomDestinationsGetCmd)
	logsCmd.AddCommand(logsCustomDestinationsCmd)

	// Metrics subcommands
	logsMetricsCmd.AddCommand(logsMetricsListCmd)
	logsMetricsCmd.AddCommand(logsMetricsGetCmd)
	logsMetricsCmd.AddCommand(logsMetricsDeleteCmd)
	logsCmd.AddCommand(logsMetricsCmd)

	// Restriction queries subcommands
	logsRestrictionQueriesCmd.AddCommand(logsRestrictionQueriesListCmd)
	logsRestrictionQueriesCmd.AddCommand(logsRestrictionQueriesGetCmd)
	logsCmd.AddCommand(logsRestrictionQueriesCmd)
}

// Helper functions

// parseTimeString converts relative or absolute time to Unix timestamp in milliseconds
func parseTimeString(timeStr string) (int64, error) {
	if timeStr == "now" {
		return time.Now().UnixMilli(), nil
	}

	// Try parsing as relative time (1h, 30m, 7d)
	if len(timeStr) >= 2 {
		unit := timeStr[len(timeStr)-1:]
		valueStr := timeStr[:len(timeStr)-1]

		var value int64
		if _, err := fmt.Sscanf(valueStr, "%d", &value); err == nil {
			var duration time.Duration
			switch unit {
			case "s":
				duration = time.Duration(value) * time.Second
			case "m":
				duration = time.Duration(value) * time.Minute
			case "h":
				duration = time.Duration(value) * time.Hour
			case "d":
				duration = time.Duration(value) * 24 * time.Hour
			case "w":
				duration = time.Duration(value) * 7 * 24 * time.Hour
			default:
				return 0, fmt.Errorf("invalid time unit: %s (use s, m, h, d, or w)", unit)
			}
			return time.Now().Add(-duration).UnixMilli(), nil
		}
	}

	// Try parsing as Unix timestamp (milliseconds)
	var timestamp int64
	if _, err := fmt.Sscanf(timeStr, "%d", &timestamp); err == nil {
		return timestamp, nil
	}

	return 0, fmt.Errorf("invalid time format: %s (use relative like '1h' or Unix timestamp)", timeStr)
}

// parseComputeString parses compute strings like "count", "avg(@duration)", "percentile(@duration, 99)"
// and returns the aggregation function and metric field
func parseComputeString(compute string) (aggregation string, metric string, err error) {
	compute = strings.TrimSpace(compute)

	// List of valid aggregation functions (from API error message)
	validFunctions := []string{
		"count", "max", "min", "avg", "sum", "median",
		"cardinality", "delta", "most_frequent", "earliest",
		"any", "latest", "dd_sketch", "top_n",
	}

	// Check for simple count
	if strings.ToLower(compute) == "count" {
		return "count", "", nil
	}

	// Parse format: function(metric) or function(metric, param)
	// Examples: avg(@duration), percentile(@duration, 99), cardinality(@user.id)
	re := regexp.MustCompile(`^(\w+)\(([^,)]+)(?:,\s*\d+)?\)$`)
	matches := re.FindStringSubmatch(compute)

	if matches == nil {
		// No parentheses - treat as a simple aggregation function
		funcLower := strings.ToLower(compute)
		for _, valid := range validFunctions {
			if funcLower == valid {
				return funcLower, "", nil
			}
		}
		return "", "", fmt.Errorf("invalid compute format: %q\n\nExpected format:\n  - count\n  - function(metric) e.g. avg(@duration), sum(@bytes), cardinality(@user.id)\n  - percentile(metric, N) e.g. percentile(@duration, 99)\n\nSupported functions: %s",
			compute, strings.Join(validFunctions, ", "))
	}

	aggregation = strings.ToLower(matches[1])
	metric = strings.TrimSpace(matches[2])

	// Validate aggregation function
	isValid := false
	for _, valid := range validFunctions {
		if aggregation == valid {
			isValid = true
			break
		}
	}
	// Also allow percentile (pcNN format)
	if strings.HasPrefix(aggregation, "pc") || aggregation == "percentile" {
		isValid = true
	}

	if !isValid {
		return "", "", fmt.Errorf("unknown aggregation function: %q\n\nSupported functions: %s",
			aggregation, strings.Join(validFunctions, ", "))
	}

	return aggregation, metric, nil
}

// Implementation functions

func runLogsSearch(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	fromTime, err := parseTimeString(logsFrom)
	if err != nil {
		return fmt.Errorf("invalid --from time: %w", err)
	}

	toTime, err := parseTimeString(logsTo)
	if err != nil {
		return fmt.Errorf("invalid --to time: %w", err)
	}

	// Use v2 API instead of deprecated v1 API
	api := datadogV2.NewLogsApi(client.V2())

	query := logsQuery
	from := fmt.Sprintf("%d", fromTime)
	to := fmt.Sprintf("%d", toTime)
	limit := int32(logsLimit)

	// Convert v1 sort values (asc/desc) to v2 format (timestamp/-timestamp)
	v2Sort := datadogV2.LogsSort("-timestamp") // default: descending
	if logsSort == "asc" {
		v2Sort = datadogV2.LogsSort("timestamp")
	}

	body := datadogV2.LogsListRequest{
		Filter: &datadogV2.LogsQueryFilter{
			Query: &query,
			From:  &from,
			To:    &to,
		},
		Page: &datadogV2.LogsListRequestPage{
			Limit: &limit,
		},
		Sort: &v2Sort,
	}

	// Note: v2 API doesn't support the index parameter the same way v1 did
	// If index filtering is needed, it should be included in the query string

	opts := datadogV2.ListLogsOptionalParameters{
		Body: &body,
	}

	// Fetch first page
	resp, r, err := api.ListLogs(client.Context(), opts)
	if err != nil {
		if r != nil && r.Body != nil {
			bodyBytes, readErr := io.ReadAll(r.Body)
			if readErr == nil && len(bodyBytes) > 0 {
				fromTimeObj := time.UnixMilli(fromTime)
				toTimeObj := time.UnixMilli(toTime)
				return fmt.Errorf("failed to search logs: %w\nStatus: %d\nAPI Response: %s\n\nRequest Details:\n- Query: %s\n- From: %s (parsed from: %s)\n- To: %s (parsed from: %s)\n- Limit: %d\n\nTroubleshooting:\n- Verify your time range is valid\n- Check that your query syntax is correct\n- Ensure you have proper permissions",
					err, r.StatusCode, string(bodyBytes),
					logsQuery,
					fromTimeObj.Format(time.RFC3339), logsFrom,
					toTimeObj.Format(time.RFC3339), logsTo,
					logsLimit)
			}
			return fmt.Errorf("failed to search logs: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to search logs: %w", err)
	}

	// Collect logs up to the requested limit
	allLogs := resp.GetData()
	pageCount := 1

	// Follow pagination until we hit the limit or run out of pages
	for logsLimit > 0 && len(allLogs) < logsLimit {
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

		remaining := logsLimit - len(allLogs)
		if remaining <= 0 {
			break
		}
		remainingLimit := int32(remaining)
		if remainingLimit < limit {
			body.Page.Limit = &remainingLimit
		}
		body.Page.Cursor = cursor
		opts.Body = &body
		resp, r, err = api.ListLogs(client.Context(), opts)
		if err != nil {
			printOutput("Warning: Failed to fetch page %d: %v\n", pageCount+1, err)
			break
		}

		allLogs = append(allLogs, resp.GetData()...)
		pageCount++
	}

	if logsLimit > 0 && len(allLogs) > logsLimit {
		allLogs = allLogs[:logsLimit]
	}

	// Show helpful message if no logs found
	if len(allLogs) == 0 {
		printOutput("No logs found matching your query.\n\n")
		printOutput("Tips:\n")
		printOutput("- Try a broader time range (e.g., --from=\"30d\")\n")
		printOutput("- Verify the service name exists in your logs\n")
		printOutput("- Check your query syntax: https://docs.datadoghq.com/logs/explorer/search_syntax/\n")
		printOutput("- Try a simpler query like --query=\"*\" to see any logs\n")
		return nil
	}

	finalResp := resp
	if pageCount > 1 {
		finalResp.SetData(allLogs)
		printOutput("Fetched %d logs across %d pages\n\n", len(allLogs), pageCount)
	}

	output, err := formatter.FormatOutput(finalResp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}

	printOutput("%s\n", output)
	return nil
}

func runLogsList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	fromTime, err := parseTimeString(logsFrom)
	if err != nil {
		return fmt.Errorf("invalid --from time: %w", err)
	}

	toTime, err := parseTimeString(logsTo)
	if err != nil {
		return fmt.Errorf("invalid --to time: %w", err)
	}

	api := datadogV2.NewLogsApi(client.V2())

	query := logsQuery
	from := fmt.Sprintf("%d", fromTime)
	to := fmt.Sprintf("%d", toTime)
	limit := int32(logsLimit)
	sort := datadogV2.LogsSort(logsSort)

	opts := datadogV2.ListLogsOptionalParameters{
		Body: &datadogV2.LogsListRequest{
			Filter: &datadogV2.LogsQueryFilter{
				Query: &query,
				From:  &from,
				To:    &to,
			},
			Page: &datadogV2.LogsListRequestPage{
				Limit: &limit,
			},
			Sort: &sort,
		},
	}

	resp, r, err := api.ListLogs(client.Context(), opts)
	if err != nil {
		if r != nil && r.Body != nil {
			bodyBytes, readErr := io.ReadAll(r.Body)
			if readErr == nil && len(bodyBytes) > 0 {
				return fmt.Errorf("failed to list logs: %w\nStatus: %d\nAPI Response: %s", err, r.StatusCode, string(bodyBytes))
			}
			return fmt.Errorf("failed to list logs: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to list logs: %w", err)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}

	printOutput("%s\n", output)
	return nil
}

func runLogsQuery(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	fromTime, err := parseTimeString(logsFrom)
	if err != nil {
		return fmt.Errorf("invalid --from time: %w", err)
	}

	toTime, err := parseTimeString(logsTo)
	if err != nil {
		return fmt.Errorf("invalid --to time: %w", err)
	}

	api := datadogV2.NewLogsApi(client.V2())

	query := logsQuery
	from := fmt.Sprintf("%d", fromTime)
	to := fmt.Sprintf("%d", toTime)
	limit := int32(logsLimit)
	sort := datadogV2.LogsSort(logsSort)

	body := datadogV2.LogsListRequest{
		Filter: &datadogV2.LogsQueryFilter{
			Query: &query,
			From:  &from,
			To:    &to,
		},
		Page: &datadogV2.LogsListRequestPage{
			Limit: &limit,
		},
		Sort: &sort,
	}

	opts := datadogV2.ListLogsOptionalParameters{
		Body: &body,
	}

	resp, r, err := api.ListLogs(client.Context(), opts)
	if err != nil {
		if r != nil && r.Body != nil {
			bodyBytes, readErr := io.ReadAll(r.Body)
			if readErr == nil && len(bodyBytes) > 0 {
				return fmt.Errorf("failed to query logs: %w\nStatus: %d\nAPI Response: %s", err, r.StatusCode, string(bodyBytes))
			}
			return fmt.Errorf("failed to query logs: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to query logs: %w", err)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}

	printOutput("%s\n", output)
	return nil
}

func runLogsAggregate(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	fromTime, err := parseTimeString(logsFrom)
	if err != nil {
		return fmt.Errorf("invalid --from time: %w", err)
	}

	toTime, err := parseTimeString(logsTo)
	if err != nil {
		return fmt.Errorf("invalid --to time: %w", err)
	}

	// Parse the compute string to extract aggregation and metric
	aggregation, metric, err := parseComputeString(logsCompute)
	if err != nil {
		return fmt.Errorf("invalid --compute value: %w", err)
	}

	api := datadogV2.NewLogsApi(client.V2())

	// Build compute aggregation
	compute := datadogV2.LogsCompute{
		Aggregation: datadogV2.LogsAggregationFunction(aggregation),
	}

	// Add metric field if present
	if metric != "" {
		compute.Metric = &metric
	}

	query := logsQuery
	from := fmt.Sprintf("%d", fromTime)
	to := fmt.Sprintf("%d", toTime)

	body := datadogV2.LogsAggregateRequest{
		Compute: []datadogV2.LogsCompute{compute},
		Filter: &datadogV2.LogsQueryFilter{
			Query: &query,
			From:  &from,
			To:    &to,
		},
	}

	// Add group by if specified
	if logsGroupBy != "" {
		limit := int64(logsLimit)
		body.GroupBy = []datadogV2.LogsGroupBy{
			{
				Facet: logsGroupBy,
				Limit: &limit,
			},
		}
	}

	resp, r, err := api.AggregateLogs(client.Context(), body)
	if err != nil {
		if r != nil && r.Body != nil {
			bodyBytes, readErr := io.ReadAll(r.Body)
			if readErr == nil && len(bodyBytes) > 0 {
				fromTimeObj := time.UnixMilli(fromTime)
				toTimeObj := time.UnixMilli(toTime)
				return fmt.Errorf("failed to aggregate logs: %w\nStatus: %d\nAPI Response: %s\n\nRequest Details:\n- Query: %s\n- Compute: %s (parsed as: aggregation=%q, metric=%q)\n- Group By: %s\n- From: %s (parsed from: %s)\n- To: %s (parsed from: %s)\n- Limit: %d\n\nTroubleshooting:\n- Verify the aggregation function is supported\n- Ensure the metric field exists in your logs (e.g., @duration, @bytes)\n- Check your query syntax\n- Verify your time range is valid",
					err, r.StatusCode, string(bodyBytes),
					logsQuery,
					logsCompute, aggregation, metric,
					logsGroupBy,
					fromTimeObj.Format(time.RFC3339), logsFrom,
					toTimeObj.Format(time.RFC3339), logsTo,
					logsLimit)
			}
			return fmt.Errorf("failed to aggregate logs: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to aggregate logs: %w", err)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}

	printOutput("%s\n", output)
	return nil
}

func runLogsArchivesList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewLogsArchivesApi(client.V2())

	resp, r, err := api.ListLogsArchives(client.Context())
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to list log archives: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to list log archives: %w", err)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}

	printOutput("%s\n", output)
	return nil
}

func runLogsArchivesGet(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	archiveID := args[0]
	api := datadogV2.NewLogsArchivesApi(client.V2())

	resp, r, err := api.GetLogsArchive(client.Context(), archiveID)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to get log archive: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to get log archive: %w", err)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}

	printOutput("%s\n", output)
	return nil
}

func runLogsArchivesDelete(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	archiveID := args[0]

	if !cfg.AutoApprove {
		fmt.Printf("⚠️  WARNING: This will delete the archive configuration for %s\n", archiveID)
		fmt.Println("Note: Archived data in storage will NOT be deleted.")
		fmt.Print("Are you sure you want to continue? (y/N): ")

		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			// User cancelled or error reading input
			fmt.Println("\nOperation cancelled")
			return nil
		}
		if response != "y" && response != "Y" {
			fmt.Println("Operation cancelled")
			return nil
		}
	}

	api := datadogV2.NewLogsArchivesApi(client.V2())

	r, err := api.DeleteLogsArchive(client.Context(), archiveID)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to delete log archive: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to delete log archive: %w", err)
	}

	fmt.Printf("Successfully deleted archive: %s\n", archiveID)
	return nil
}

func runLogsCustomDestinationsList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewLogsCustomDestinationsApi(client.V2())

	resp, r, err := api.ListLogsCustomDestinations(client.Context())
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to list custom destinations: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to list custom destinations: %w", err)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}

	printOutput("%s\n", output)
	return nil
}

func runLogsCustomDestinationsGet(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	destinationID := args[0]
	api := datadogV2.NewLogsCustomDestinationsApi(client.V2())

	resp, r, err := api.GetLogsCustomDestination(client.Context(), destinationID)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to get custom destination: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to get custom destination: %w", err)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}

	printOutput("%s\n", output)
	return nil
}

func runLogsMetricsList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewLogsMetricsApi(client.V2())

	resp, r, err := api.ListLogsMetrics(client.Context())
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to list log-based metrics: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to list log-based metrics: %w", err)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}

	printOutput("%s\n", output)
	return nil
}

func runLogsMetricsGet(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	metricID := args[0]
	api := datadogV2.NewLogsMetricsApi(client.V2())

	resp, r, err := api.GetLogsMetric(client.Context(), metricID)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to get log-based metric: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to get log-based metric: %w", err)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}

	printOutput("%s\n", output)
	return nil
}

func runLogsMetricsDelete(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	metricID := args[0]

	if !cfg.AutoApprove {
		fmt.Printf("⚠️  WARNING: This will permanently delete log-based metric %s\n", metricID)
		fmt.Print("Are you sure you want to continue? (y/N): ")

		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			// User cancelled or error reading input
			fmt.Println("\nOperation cancelled")
			return nil
		}
		if response != "y" && response != "Y" {
			fmt.Println("Operation cancelled")
			return nil
		}
	}

	api := datadogV2.NewLogsMetricsApi(client.V2())

	r, err := api.DeleteLogsMetric(client.Context(), metricID)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to delete log-based metric: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to delete log-based metric: %w", err)
	}

	fmt.Printf("Successfully deleted metric: %s\n", metricID)
	return nil
}

func runLogsRestrictionQueriesList(cmd *cobra.Command, args []string) error {
	// NOTE: LogsRestrictionQueriesApi is not available in datadog-api-client-go v2.30.0
	return fmt.Errorf("logs restriction queries API is not available in the current API client version")
}

func runLogsRestrictionQueriesGet(cmd *cobra.Command, args []string) error {
	// NOTE: LogsRestrictionQueriesApi is not available in datadog-api-client-go v2.30.0
	return fmt.Errorf("logs restriction queries API is not available in the current API client version")
}
