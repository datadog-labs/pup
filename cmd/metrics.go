// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/DataDog/pup/pkg/formatter"
	"github.com/spf13/cobra"
)

var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Query and manage metrics",
	Long: `Query time-series metrics, list available metrics, and manage metric metadata.

Metrics are the foundation of monitoring in Datadog. This command provides
comprehensive access to query metrics data, list available metrics, manage
metadata, and submit custom metrics.

CAPABILITIES:
  • Query time-series metrics data with flexible time ranges
  • List all available metrics with optional filtering
  • Get and update metric metadata (description, unit, type)
  • Submit custom metrics to Datadog
  • List metric tags and tag configurations

METRIC TYPES:
  • gauge: Point-in-time value (e.g., CPU usage, memory)
  • count: Cumulative count (e.g., request count, errors)
  • rate: Rate of change per second (e.g., requests per second)
  • distribution: Statistical distribution (e.g., latency percentiles)

TIME RANGES:
  Supports flexible time range specifications:
  • Relative: 1h, 30m, 7d, 1w (hours, minutes, days, weeks)
  • Absolute: Unix timestamps or ISO 8601 format
  • Special: now (current time)

EXAMPLES:
  # Query metrics
  pup metrics query --query="avg:system.cpu.user{*}" --from="1h" --to="now"
  pup metrics query --query="sum:app.requests{env:prod} by {service}" --from="4h"

  # List metrics
  pup metrics list
  pup metrics list --filter="system.*"

  # Get metric metadata
  pup metrics metadata get system.cpu.user
  pup metrics metadata get system.cpu.user --output=table

  # Update metric metadata
  pup metrics metadata update system.cpu.user \
    --description="CPU user time" \
    --unit="percent" \
    --type="gauge"

  # Submit custom metrics
  pup metrics submit --name="custom.metric" --value=123 --tags="env:prod,team:backend"
  pup metrics submit --name="custom.gauge" --value=99.5 --type="gauge" --timestamp=now

  # List metric tags
  pup metrics tags list system.cpu.user
  pup metrics tags list system.cpu.user --from="1h"

AUTHENTICATION:
  Requires either OAuth2 authentication (pup auth login) or API keys
  (DD_API_KEY and DD_APP_KEY environment variables).`,
}

// Query command
var metricsQueryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query time-series metrics data",
	Long: `Query time-series metrics data with flexible aggregation and filtering.

This command queries metrics data from Datadog using the metrics query language.
You can specify aggregation functions, filters, grouping, and time ranges.

QUERY SYNTAX:
  <aggregation>:<metric_name>{<filter>} [by {<group>}]

  Examples:
  • avg:system.cpu.user{*}
  • sum:app.requests{env:prod} by {service}
  • max:system.disk.used{host:web-*}
  • avg:system.load.1{availability-zone:us-east-1a} by {host}

AGGREGATIONS:
  • avg: Average value
  • sum: Sum of all values
  • min: Minimum value
  • max: Maximum value
  • count: Count of data points

TIME RANGES:
  • Relative: 1h, 30m, 7d, 1w, 1M (hours, minutes, days, weeks, months)
  • Absolute: Unix timestamp (seconds)
  • Special: now (current time)

EXAMPLES:
  # Query CPU usage for the last hour
  pup metrics query --query="avg:system.cpu.user{*}" --from="1h" --to="now"

  # Query request count by service for last 4 hours
  pup metrics query --query="sum:app.requests{env:prod} by {service}" --from="4h"

  # Query memory usage for specific hosts
  pup metrics query --query="avg:system.mem.used{host:web-*}" --from="2h"

  # Query with absolute timestamps
  pup metrics query --query="avg:system.load.1{*}" --from="1704067200" --to="1704153600"

OUTPUT:
  Returns time-series data including:
  • series: Array of time-series data points
  • from_date: Query start time (Unix timestamp seconds)
  • to_date: Query end time (Unix timestamp seconds)
  • query: The query string used
  • res_type: Response type
  • resp_version: Response version`,
	RunE: runMetricsQuery,
}

// List command
var metricsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available metrics",
	Long: `List all available metrics in your Datadog account.

This command retrieves the list of all metrics that have been submitted to
Datadog. You can optionally filter the list using a metric name pattern.

FILTERING:
  Use the --filter flag to search for metrics matching a pattern:
  • system.* - All system metrics
  • *.cpu.* - All CPU-related metrics
  • custom.* - All custom metrics
  • myapp.* - All metrics starting with myapp

EXAMPLES:
  # List all metrics
  pup metrics list

  # List system metrics
  pup metrics list --filter="system.*"

  # List CPU metrics
  pup metrics list --filter="*.cpu.*"

  # List custom metrics
  pup metrics list --filter="custom.*"

  # Search for specific metrics
  pup metrics list --filter="*request*"

OUTPUT:
  Returns an array of metric names. The response may be paginated for
  large metric sets.

PAGINATION:
  Currently returns all matching metrics. For very large metric sets,
  consider using more specific filters.`,
	RunE: runMetricsList,
}

// Metadata command group
var metricsMetadataCmd = &cobra.Command{
	Use:   "metadata",
	Short: "Manage metric metadata",
	Long: `Get and update metric metadata including description, unit, and type.

Metric metadata provides context about what a metric represents, its unit
of measurement, and its type. This information helps teams understand and
correctly interpret metrics.

METADATA FIELDS:
  • description: Human-readable description of the metric
  • unit: Unit of measurement (e.g., byte, percent, request)
  • type: Metric type (gauge, count, rate, distribution)
  • per_unit: Per-unit for rate metrics (e.g., second)
  • short_name: Short name for display

EXAMPLES:
  # Get metric metadata
  pup metrics metadata get system.cpu.user

  # Update metric metadata
  pup metrics metadata update system.cpu.user \
    --description="Percentage of CPU time spent in user space" \
    --unit="percent" \
    --type="gauge"

  # Update multiple fields
  pup metrics metadata update custom.response.time \
    --description="API response time" \
    --unit="millisecond" \
    --type="gauge" \
    --short-name="Response Time"`,
}

var metricsMetadataGetCmd = &cobra.Command{
	Use:   "get [metric-name]",
	Short: "Get metric metadata",
	Long: `Get metadata for a specific metric.

Retrieves all metadata associated with a metric including description,
unit, type, integration information, and more.

ARGUMENTS:
  metric-name    The name of the metric (e.g., system.cpu.user)

EXAMPLES:
  # Get metadata for a system metric
  pup metrics metadata get system.cpu.user

  # Get metadata for a custom metric
  pup metrics metadata get custom.api.latency

  # Get metadata with table output
  pup metrics metadata get system.cpu.user --output=table

OUTPUT:
  Returns metric metadata including:
  • description: Metric description
  • unit: Unit of measurement
  • type: Metric type
  • per_unit: Per-unit for rate metrics
  • short_name: Display name
  • integration: Integration name (if applicable)
  • statsd_interval: StatsD flush interval`,
	Args: cobra.ExactArgs(1),
	RunE: runMetricsMetadataGet,
}

var metricsMetadataUpdateCmd = &cobra.Command{
	Use:   "update [metric-name]",
	Short: "Update metric metadata",
	Long: `Update metadata for a specific metric.

Updates one or more metadata fields for a metric. Only specified fields
will be updated; other fields remain unchanged.

ARGUMENTS:
  metric-name    The name of the metric to update

FLAGS:
  --description    Metric description
  --unit          Unit of measurement
  --type          Metric type (gauge, count, rate, distribution)
  --per-unit      Per-unit for rate metrics
  --short-name    Short display name

EXAMPLES:
  # Update description only
  pup metrics metadata update custom.api.latency \
    --description="API endpoint response latency"

  # Update multiple fields
  pup metrics metadata update custom.request.rate \
    --description="Request rate per second" \
    --unit="request" \
    --type="rate" \
    --per-unit="second"

  # Update unit and type
  pup metrics metadata update custom.memory.used \
    --unit="byte" \
    --type="gauge"

OUTPUT:
  Returns success message with updated metadata.`,
	Args: cobra.ExactArgs(1),
	RunE: runMetricsMetadataUpdate,
}

// Submit command
var metricsSubmitCmd = &cobra.Command{
	Use:   "submit",
	Short: "Submit custom metrics to Datadog",
	Long: `Submit custom metric data points to Datadog.

This command allows you to submit custom metrics from the command line.
Useful for testing, scripting, and one-off metric submissions.

METRIC TYPES:
  • gauge: Current value (default)
  • count: Cumulative count
  • rate: Rate per second

REQUIRED FLAGS:
  --name         Metric name (e.g., custom.my.metric)
  --value        Metric value (numeric)

OPTIONAL FLAGS:
  --type         Metric type (gauge, count, rate) [default: gauge]
  --timestamp    Unix timestamp or "now" [default: now]
  --tags         Comma-separated tags (e.g., env:prod,team:api)
  --host         Host name to associate with metric
  --interval     Interval for rate/count metrics (seconds)

EXAMPLES:
  # Submit a gauge metric
  pup metrics submit --name="custom.temperature" --value=72.5

  # Submit with tags
  pup metrics submit \
    --name="custom.api.requests" \
    --value=1250 \
    --tags="env:prod,service:api,region:us-east-1"

  # Submit a count metric
  pup metrics submit \
    --name="custom.events.processed" \
    --value=100 \
    --type="count"

  # Submit with specific timestamp
  pup metrics submit \
    --name="custom.batch.size" \
    --value=5000 \
    --timestamp="1704067200"

  # Submit with host
  pup metrics submit \
    --name="custom.worker.queue.size" \
    --value=42 \
    --host="worker-01.example.com" \
    --tags="env:prod"

OUTPUT:
  Returns success message with submission details.

NOTES:
  • Metrics are submitted to the v2 metrics intake API
  • Values can be integers or floating-point numbers
  • Tags must follow the format key:value
  • Metric names should use lowercase with dots/underscores`,
	RunE: runMetricsSubmit,
}

// Tags command group
var metricsTagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "Manage metric tags",
	Long: `List and manage metric tag configurations.

Metric tags provide dimensions for filtering and grouping metrics.
This command allows you to explore available tags for metrics.

EXAMPLES:
  # List tags for a metric
  pup metrics tags list system.cpu.user

  # List tags for a specific time period
  pup metrics tags list system.cpu.user --from="1h"

  # List tags for custom metric
  pup metrics tags list custom.api.latency --from="24h"`,
}

var metricsTagsListCmd = &cobra.Command{
	Use:   "list [metric-name]",
	Short: "List tags for a metric",
	Long: `List all tag keys and values for a specific metric.

Retrieves all unique tag combinations that have been submitted with
a metric over the specified time period.

ARGUMENTS:
  metric-name    The name of the metric

FLAGS:
  --from         Start time (relative or absolute) [default: 1h]
  --to           End time (relative or absolute) [default: now]

EXAMPLES:
  # List tags for the last hour
  pup metrics tags list system.cpu.user

  # List tags for the last 24 hours
  pup metrics tags list system.cpu.user --from="24h"

  # List tags for custom metric
  pup metrics tags list custom.api.requests --from="7d"

OUTPUT:
  Returns array of tag strings in key:value format.`,
	Args: cobra.ExactArgs(1),
	RunE: runMetricsTagsList,
}

// Command flags
var (
	// Query flags
	queryString string
	fromTime    string
	toTime      string

	// List flags
	filterPattern string

	// Metadata update flags
	metadataDescription string
	metadataUnit        string
	metadataType        string
	metadataPerUnit     string
	metadataShortName   string

	// Submit flags
	submitName      string
	submitValue     float64
	submitType      string
	submitTimestamp string
	submitTags      string
	submitHost      string
	submitInterval  int64
)

func init() {
	// Query command flags
	metricsQueryCmd.Flags().StringVar(&queryString, "query", "", "Metric query string (required)")
	metricsQueryCmd.Flags().StringVar(&fromTime, "from", "1h", "Start time (e.g., 1h, 30m, 7d, now, unix timestamp)")
	metricsQueryCmd.Flags().StringVar(&toTime, "to", "now", "End time (e.g., now, unix timestamp)")
	if err := metricsQueryCmd.MarkFlagRequired("query"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}

	// List command flags
	metricsListCmd.Flags().StringVar(&filterPattern, "filter", "", "Filter metrics by pattern (e.g., system.*)")

	// Metadata update flags
	metricsMetadataUpdateCmd.Flags().StringVar(&metadataDescription, "description", "", "Metric description")
	metricsMetadataUpdateCmd.Flags().StringVar(&metadataUnit, "unit", "", "Metric unit")
	metricsMetadataUpdateCmd.Flags().StringVar(&metadataType, "type", "", "Metric type (gauge, count, rate, distribution)")
	metricsMetadataUpdateCmd.Flags().StringVar(&metadataPerUnit, "per-unit", "", "Per-unit for rate metrics")
	metricsMetadataUpdateCmd.Flags().StringVar(&metadataShortName, "short-name", "", "Short display name")

	// Submit command flags
	metricsSubmitCmd.Flags().StringVar(&submitName, "name", "", "Metric name (required)")
	metricsSubmitCmd.Flags().Float64Var(&submitValue, "value", 0, "Metric value (required)")
	metricsSubmitCmd.Flags().StringVar(&submitType, "type", "gauge", "Metric type (gauge, count, rate)")
	metricsSubmitCmd.Flags().StringVar(&submitTimestamp, "timestamp", "now", "Timestamp (now or unix timestamp)")
	metricsSubmitCmd.Flags().StringVar(&submitTags, "tags", "", "Comma-separated tags (e.g., env:prod,team:api)")
	metricsSubmitCmd.Flags().StringVar(&submitHost, "host", "", "Host name")
	metricsSubmitCmd.Flags().Int64Var(&submitInterval, "interval", 0, "Interval in seconds for rate/count metrics")
	if err := metricsSubmitCmd.MarkFlagRequired("name"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}
	if err := metricsSubmitCmd.MarkFlagRequired("value"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}

	// Tags command flags
	metricsTagsListCmd.Flags().StringVar(&fromTime, "from", "1h", "Start time")
	metricsTagsListCmd.Flags().StringVar(&toTime, "to", "now", "End time")

	// Add subcommands to metadata
	metricsMetadataCmd.AddCommand(metricsMetadataGetCmd)
	metricsMetadataCmd.AddCommand(metricsMetadataUpdateCmd)

	// Add subcommands to tags
	metricsTagsCmd.AddCommand(metricsTagsListCmd)

	// Add subcommands to metrics
	metricsCmd.AddCommand(metricsQueryCmd)
	metricsCmd.AddCommand(metricsListCmd)
	metricsCmd.AddCommand(metricsMetadataCmd)
	metricsCmd.AddCommand(metricsSubmitCmd)
	metricsCmd.AddCommand(metricsTagsCmd)
}

// runMetricsQuery executes the metrics query command
func runMetricsQuery(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	// Parse time ranges
	from, err := parseTimeParam(fromTime)
	if err != nil {
		return fmt.Errorf("invalid --from time: %w", err)
	}

	to, err := parseTimeParam(toTime)
	if err != nil {
		return fmt.Errorf("invalid --to time: %w", err)
	}

	// Use v2 API for timeseries query (more flexible)
	api := datadogV2.NewMetricsApi(client.V2())

	// Build query request
	metricsQuery := datadogV2.NewMetricsTimeseriesQuery(datadogV2.METRICSDATASOURCE_METRICS, queryString)
	timeseriesQuery := datadogV2.MetricsTimeseriesQueryAsTimeseriesQuery(metricsQuery)

	body := datadogV2.TimeseriesFormulaQueryRequest{
		Data: datadogV2.TimeseriesFormulaRequest{
			Attributes: datadogV2.TimeseriesFormulaRequestAttributes{
				From:    from.UTC().UnixMilli(),
				To:      to.UTC().UnixMilli(),
				Queries: []datadogV2.TimeseriesQuery{timeseriesQuery},
			},
			Type: datadogV2.TIMESERIESFORMULAREQUESTTYPE_TIMESERIES_REQUEST,
		},
	}

	resp, r, err := api.QueryTimeseriesData(client.Context(), body)
	if err != nil {
		if r != nil && r.Body != nil {
			bodyBytes, readErr := io.ReadAll(r.Body)
			if readErr == nil && len(bodyBytes) > 0 {
				return fmt.Errorf("failed to query metrics: %w\nStatus: %d\nAPI Response: %s\n\nRequest Details:\n- Query: %s\n- From: %s (Unix: %d)\n- To: %s (Unix: %d)\n\nTroubleshooting:\n- Verify your query syntax is correct (e.g., avg:metric.name{filter})\n- Check that the time range is valid\n- Ensure the metric exists and has data in the specified time range\n- Confirm you have proper permissions to access the metric",
					err, r.StatusCode, string(bodyBytes),
					queryString,
					from.Format(time.RFC3339), from.Unix(),
					to.Format(time.RFC3339), to.Unix())
			}
			return fmt.Errorf("failed to query metrics: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to query metrics: %w", err)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}

	printOutput("%s\n", output)
	return nil
}

// runMetricsList executes the metrics list command
func runMetricsList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV1.NewMetricsApi(client.V1())

	// From time defaults to 1 hour ago
	from := time.Now().Add(-1 * time.Hour).Unix()

	opts := datadogV1.NewListActiveMetricsOptionalParameters()
	if filterPattern != "" {
		opts = opts.WithTagFilter(filterPattern)
	}

	resp, r, err := api.ListActiveMetrics(client.Context(), from, *opts)
	if err != nil {
		if r != nil && r.Body != nil {
			bodyBytes, readErr := io.ReadAll(r.Body)
			if readErr == nil && len(bodyBytes) > 0 {
				return fmt.Errorf("failed to list metrics: %w\nStatus: %d\nAPI Response: %s\n\nRequest Details:\n- Filter: %s\n- From: %s (Unix: %d)\n\nTroubleshooting:\n- Check that your filter pattern is valid\n- Verify you have permissions to list metrics",
					err, r.StatusCode, string(bodyBytes),
					filterPattern,
					time.Unix(from, 0).Format(time.RFC3339), from)
			}
			return fmt.Errorf("failed to list metrics: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to list metrics: %w", err)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}

	printOutput("%s\n", output)
	return nil
}

// runMetricsMetadataGet executes the metadata get command
func runMetricsMetadataGet(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	metricName := args[0]
	api := datadogV1.NewMetricsApi(client.V1())

	resp, r, err := api.GetMetricMetadata(client.Context(), metricName)
	if err != nil {
		if r != nil && r.Body != nil {
			bodyBytes, readErr := io.ReadAll(r.Body)
			if readErr == nil && len(bodyBytes) > 0 {
				return fmt.Errorf("failed to get metric metadata: %w\nStatus: %d\nAPI Response: %s\n\nMetric: %s\n\nTroubleshooting:\n- Verify the metric name is correct\n- Ensure the metric exists in your account\n- Check that you have permissions to view metadata",
					err, r.StatusCode, string(bodyBytes), metricName)
			}
			return fmt.Errorf("failed to get metric metadata: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to get metric metadata: %w", err)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}

	printOutput("%s\n", output)
	return nil
}

// runMetricsMetadataUpdate executes the metadata update command
func runMetricsMetadataUpdate(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	metricName := args[0]
	api := datadogV1.NewMetricsApi(client.V1())

	// Build metadata update body
	body := datadogV1.MetricMetadata{}

	if metadataDescription != "" {
		body.SetDescription(metadataDescription)
	}
	if metadataUnit != "" {
		body.SetUnit(metadataUnit)
	}
	if metadataType != "" {
		body.SetType(metadataType)
	}
	if metadataPerUnit != "" {
		body.SetPerUnit(metadataPerUnit)
	}
	if metadataShortName != "" {
		body.SetShortName(metadataShortName)
	}

	// Check if at least one field is specified
	if !body.HasDescription() && !body.HasUnit() && !body.HasType() && !body.HasPerUnit() && !body.HasShortName() {
		return fmt.Errorf("at least one metadata field must be specified (--description, --unit, --type, --per-unit, --short-name)")
	}

	resp, r, err := api.UpdateMetricMetadata(client.Context(), metricName, body)
	if err != nil {
		if r != nil && r.Body != nil {
			bodyBytes, readErr := io.ReadAll(r.Body)
			if readErr == nil && len(bodyBytes) > 0 {
				return fmt.Errorf("failed to update metric metadata: %w\nStatus: %d\nAPI Response: %s\n\nMetric: %s\n\nTroubleshooting:\n- Verify the metric name is correct\n- Check that the metadata values are valid (unit, type, etc.)\n- Ensure you have permissions to update metadata",
					err, r.StatusCode, string(bodyBytes), metricName)
			}
			return fmt.Errorf("failed to update metric metadata: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to update metric metadata: %w", err)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}

	printOutput("%s\n", output)
	return nil
}

// runMetricsSubmit executes the metrics submit command
func runMetricsSubmit(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	// Parse timestamp
	var timestamp int64
	if submitTimestamp == "now" {
		timestamp = time.Now().Unix()
	} else {
		ts, err := strconv.ParseInt(submitTimestamp, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid timestamp: %w", err)
		}
		timestamp = ts
	}

	// Parse tags
	var tags []string
	if submitTags != "" {
		tags = strings.Split(submitTags, ",")
		// Trim whitespace from tags
		for i := range tags {
			tags[i] = strings.TrimSpace(tags[i])
		}
	}

	// Determine metric type
	var metricType datadogV2.MetricIntakeType
	switch strings.ToLower(submitType) {
	case "gauge":
		metricType = datadogV2.METRICINTAKETYPE_GAUGE
	case "count":
		metricType = datadogV2.METRICINTAKETYPE_COUNT
	case "rate":
		metricType = datadogV2.METRICINTAKETYPE_RATE
	default:
		return fmt.Errorf("invalid metric type: %s (must be gauge, count, or rate)", submitType)
	}

	// Build metric payload
	point := datadogV2.MetricPoint{
		Timestamp: &timestamp,
		Value:     &submitValue,
	}

	// Convert MetricIntakeType to string for resource type
	metricTypeStr := string(metricType)

	resource := datadogV2.MetricResource{
		Name: &submitName,
		Type: &metricTypeStr,
	}

	series := datadogV2.MetricSeries{
		Metric: submitName,
		Type:   &metricType,
		Points: []datadogV2.MetricPoint{point},
		Resources: []datadogV2.MetricResource{resource},
	}

	if len(tags) > 0 {
		series.Tags = tags
	}

	if submitInterval > 0 {
		series.Interval = &submitInterval
	}

	body := datadogV2.MetricPayload{
		Series: []datadogV2.MetricSeries{series},
	}

	// Submit using v2 API
	api := datadogV2.NewMetricsApi(client.V2())

	resp, r, err := api.SubmitMetrics(client.Context(), body, *datadogV2.NewSubmitMetricsOptionalParameters())
	if err != nil {
		if r != nil && r.Body != nil {
			bodyBytes, readErr := io.ReadAll(r.Body)
			if readErr == nil && len(bodyBytes) > 0 {
				return fmt.Errorf("failed to submit metrics: %w\nStatus: %d\nAPI Response: %s\n\nRequest Details:\n- Metric: %s\n- Value: %f\n- Type: %s\n- Timestamp: %d\n- Tags: %v\n\nTroubleshooting:\n- Verify the metric name follows naming conventions (lowercase, dots/underscores)\n- Check that the metric type is valid (gauge, count, rate)\n- Ensure your API key has permission to submit metrics\n- Verify tags are in key:value format",
					err, r.StatusCode, string(bodyBytes),
					submitName, submitValue, submitType, timestamp, tags)
			}
			return fmt.Errorf("failed to submit metrics: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to submit metrics: %w", err)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}

	printOutput("%s\n", output)
	return nil
}

// runMetricsTagsList executes the tags list command
func runMetricsTagsList(cmd *cobra.Command, args []string) error {
	// NOTE: ListTagsByMetricName is not available in datadog-api-client-go v2.30.0
	return fmt.Errorf("listing tags by metric name is not supported in the current API client version")
}

// parseTimeParam parses a time parameter (relative or absolute)
func parseTimeParam(timeStr string) (time.Time, error) {
	// Handle "now" keyword
	if strings.ToLower(timeStr) == "now" {
		return time.Now(), nil
	}

	// Try parsing as unix timestamp
	if timestamp, err := strconv.ParseInt(timeStr, 10, 64); err == nil {
		return time.Unix(timestamp, 0), nil
	}

	// Parse relative time (e.g., 1h, 30m, 7d, 1w)
	if len(timeStr) < 2 {
		return time.Time{}, fmt.Errorf("invalid time format: %s", timeStr)
	}

	valueStr := timeStr[:len(timeStr)-1]
	unit := timeStr[len(timeStr)-1:]

	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid time value: %s", timeStr)
	}

	now := time.Now()
	var duration time.Duration

	switch strings.ToLower(unit) {
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
		return time.Time{}, fmt.Errorf("invalid time unit: %s (use s, m, h, d, or w)", unit)
	}

	return now.Add(-duration), nil
}
