// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/spf13/cobra"
)

var cicdCmd = &cobra.Command{
	Use:   "cicd",
	Short: "Manage CI/CD visibility",
	Long: `Manage Datadog CI/CD visibility for pipeline and test monitoring.

CI/CD Visibility provides insights into your CI/CD pipelines, tracking pipeline
performance, test results, and failure patterns.

CAPABILITIES:
  • List and search CI pipelines with filtering
  • Get detailed pipeline execution information
  • Aggregate pipeline events for analytics
  • Track pipeline performance metrics

EXAMPLES:
  # List recent pipelines
  pup cicd pipelines list

  # Get pipeline details
  pup cicd pipelines get --pipeline-id="abc-123"

  # Search for failed pipelines
  pup cicd events search --query="@ci.status:error" --from="1h"

  # Aggregate by status
  pup cicd events aggregate --query="*" --compute="count" --group-by="@ci.status"

AUTHENTICATION:
  Requires either OAuth2 authentication (pup auth login) or API keys.`,
}

var cicdPipelinesCmd = &cobra.Command{
	Use:   "pipelines",
	Short: "Manage CI pipelines",
}

var cicdEventsCmd = &cobra.Command{
	Use:   "events",
	Short: "Query CI/CD events",
}

var cicdPipelinesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List CI pipelines",
	RunE:  runCICDPipelinesList,
}

var cicdPipelinesGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get pipeline details",
	RunE:  runCICDPipelinesGet,
}

var cicdEventsSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search CI/CD events",
	RunE:  runCICDEventsSearch,
}

var cicdEventsAggregateCmd = &cobra.Command{
	Use:   "aggregate",
	Short: "Aggregate CI/CD events",
	RunE:  runCICDEventsAggregate,
}

// DORA subcommands
var cicdDoraCmd = &cobra.Command{
	Use:   "dora",
	Short: "Manage DORA metrics",
}

var cicdDoraPatchDeploymentCmd = &cobra.Command{
	Use:   "patch-deployment [deployment-id]",
	Short: "Patch a DORA deployment",
	Args:  cobra.ExactArgs(1),
	RunE:  runCICDDoraPatchDeployment,
}

// Flaky Tests subcommands
var cicdFlakyTestsCmd = &cobra.Command{
	Use:   "flaky-tests",
	Short: "Manage flaky tests",
}

var cicdFlakyTestsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update flaky tests",
	RunE:  runCICDFlakyTestsUpdate,
}

var (
	pipelineID   string
	pipelineName string
	branch       string
	cicdQuery    string
	cicdFrom     string
	cicdTo       string
	cicdLimit    int32
	cicdSort     string
	cicdCompute  string
	cicdGroupBy  string
	cicdFile     string
)

func init() {
	cicdPipelinesListCmd.Flags().StringVar(&pipelineName, "pipeline-name", "", "Filter by pipeline name")
	cicdPipelinesListCmd.Flags().StringVar(&branch, "branch", "", "Filter by git branch")
	cicdPipelinesListCmd.Flags().StringVar(&cicdFrom, "from", "1h", "Start time")
	cicdPipelinesListCmd.Flags().StringVar(&cicdTo, "to", "now", "End time")

	cicdPipelinesGetCmd.Flags().StringVar(&pipelineID, "pipeline-id", "", "Pipeline ID (required)")
	if err := cicdPipelinesGetCmd.MarkFlagRequired("pipeline-id"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}

	cicdEventsSearchCmd.Flags().StringVar(&cicdQuery, "query", "", "Search query (required)")
	cicdEventsSearchCmd.Flags().StringVar(&cicdFrom, "from", "1h", "Start time")
	cicdEventsSearchCmd.Flags().StringVar(&cicdTo, "to", "now", "End time")
	cicdEventsSearchCmd.Flags().Int32Var(&cicdLimit, "limit", 50, "Maximum results")
	cicdEventsSearchCmd.Flags().StringVar(&cicdSort, "sort", "desc", "Sort order (asc or desc)")
	if err := cicdEventsSearchCmd.MarkFlagRequired("query"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}

	cicdEventsAggregateCmd.Flags().StringVar(&cicdQuery, "query", "", "Search query (required)")
	cicdEventsAggregateCmd.Flags().StringVar(&cicdFrom, "from", "1h", "Start time")
	cicdEventsAggregateCmd.Flags().StringVar(&cicdTo, "to", "now", "End time")
	cicdEventsAggregateCmd.Flags().StringVar(&cicdCompute, "compute", "count", "Aggregation function")
	cicdEventsAggregateCmd.Flags().StringVar(&cicdGroupBy, "group-by", "", "Group by field(s)")
	cicdEventsAggregateCmd.Flags().Int32Var(&cicdLimit, "limit", 10, "Maximum groups")
	if err := cicdEventsAggregateCmd.MarkFlagRequired("query"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}

	// DORA flags
	cicdDoraPatchDeploymentCmd.Flags().StringVar(&cicdFile, "file", "", "JSON file with patch data (required)")
	_ = cicdDoraPatchDeploymentCmd.MarkFlagRequired("file")

	// Flaky tests flags
	cicdFlakyTestsUpdateCmd.Flags().StringVar(&cicdFile, "file", "", "JSON file with flaky tests data (required)")
	_ = cicdFlakyTestsUpdateCmd.MarkFlagRequired("file")

	cicdPipelinesCmd.AddCommand(cicdPipelinesListCmd, cicdPipelinesGetCmd)
	cicdEventsCmd.AddCommand(cicdEventsSearchCmd, cicdEventsAggregateCmd)
	cicdDoraCmd.AddCommand(cicdDoraPatchDeploymentCmd)
	cicdFlakyTestsCmd.AddCommand(cicdFlakyTestsUpdateCmd)
	cicdCmd.AddCommand(cicdPipelinesCmd, cicdEventsCmd, cicdDoraCmd, cicdFlakyTestsCmd)
}

func runCICDPipelinesList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewCIVisibilityPipelinesApi(client.V2())
	query := "*"
	if pipelineName != "" {
		query = fmt.Sprintf("@ci.pipeline.name:%s", pipelineName)
	}
	if branch != "" {
		if query != "*" {
			query = fmt.Sprintf("%s AND @git.branch:%s", query, branch)
		} else {
			query = fmt.Sprintf("@git.branch:%s", branch)
		}
	}

	filter := datadogV2.CIAppPipelinesQueryFilter{
		From:  &cicdFrom,
		To:    &cicdTo,
		Query: &query,
	}

	body := datadogV2.NewCIAppPipelineEventsRequest()
	body.SetFilter(filter)

	opts := datadogV2.NewSearchCIAppPipelineEventsOptionalParameters()
	opts = opts.WithBody(*body)

	resp, r, err := api.SearchCIAppPipelineEvents(client.Context(), *opts)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to list pipelines: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to list pipelines: %w", err)
	}

	return formatAndPrint(resp, nil)
}

func runCICDPipelinesGet(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewCIVisibilityPipelinesApi(client.V2())

	// Search for the specific pipeline ID using filter
	filter := datadogV2.CIAppPipelinesQueryFilter{
		Query: &pipelineID,
	}
	body := datadogV2.NewCIAppPipelineEventsRequest()
	body.SetFilter(filter)

	opts := datadogV2.NewSearchCIAppPipelineEventsOptionalParameters()
	opts = opts.WithBody(*body)

	resp, r, err := api.SearchCIAppPipelineEvents(client.Context(), *opts)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to get pipeline: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to get pipeline: %w", err)
	}

	return formatAndPrint(resp, nil)
}

func runCICDEventsSearch(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewCIVisibilityPipelinesApi(client.V2())
	var sort datadogV2.CIAppSort
	if cicdSort == "asc" {
		sort = datadogV2.CIAPPSORT_TIMESTAMP_ASCENDING
	} else {
		sort = datadogV2.CIAPPSORT_TIMESTAMP_DESCENDING
	}

	page := datadogV2.CIAppQueryPageOptions{
		Limit: &cicdLimit,
	}

	filter := datadogV2.CIAppPipelinesQueryFilter{
		From:  &cicdFrom,
		To:    &cicdTo,
		Query: &cicdQuery,
	}

	body := datadogV2.NewCIAppPipelineEventsRequest()
	body.SetFilter(filter)
	body.SetPage(page)
	body.SetSort(sort)

	opts := datadogV2.NewSearchCIAppPipelineEventsOptionalParameters()
	opts = opts.WithBody(*body)

	resp, r, err := api.SearchCIAppPipelineEvents(client.Context(), *opts)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to search events: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to search events: %w", err)
	}

	return formatAndPrint(resp, nil)
}

func runCICDEventsAggregate(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewCIVisibilityPipelinesApi(client.V2())
	compute, err := buildComputeAggregation(cicdCompute)
	if err != nil {
		return err
	}

	var groupBy []datadogV2.CIAppPipelinesGroupBy
	if cicdGroupBy != "" {
		fields := strings.Split(cicdGroupBy, ",")
		for _, field := range fields {
			field = strings.TrimSpace(field)
			gb := datadogV2.NewCIAppPipelinesGroupBy(field)
			limit := int64(cicdLimit)
			gb.SetLimit(limit)
			groupBy = append(groupBy, *gb)
		}
	}

	filter := datadogV2.CIAppPipelinesQueryFilter{
		From:  &cicdFrom,
		To:    &cicdTo,
		Query: &cicdQuery,
	}

	body := datadogV2.NewCIAppPipelinesAggregateRequest()
	body.SetCompute([]datadogV2.CIAppCompute{*compute})
	body.SetFilter(filter)

	if len(groupBy) > 0 {
		body.SetGroupBy(groupBy)
	}

	resp, r, err := api.AggregateCIAppPipelineEvents(client.Context(), *body)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to aggregate events: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to aggregate events: %w", err)
	}

	return formatAndPrint(resp, nil)
}

// DORA implementations
func runCICDDoraPatchDeployment(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile(cicdFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var body datadogV2.DORADeploymentPatchRequest
	if err := json.Unmarshal(data, &body); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewDORAMetricsApi(client.V2())
	r, err := api.PatchDORADeployment(client.Context(), args[0], body)
	if err != nil {
		return formatAPIError("patch DORA deployment", err, r)
	}

	printOutput("DORA deployment '%s' patched successfully.\n", args[0])
	return nil
}

// Flaky Tests implementations
func runCICDFlakyTestsUpdate(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile(cicdFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var body datadogV2.UpdateFlakyTestsRequest
	if err := json.Unmarshal(data, &body); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewTestOptimizationApi(client.V2())
	resp, r, err := api.UpdateFlakyTests(client.Context(), body)
	if err != nil {
		return formatAPIError("update flaky tests", err, r)
	}

	return formatAndPrint(resp, nil)
}

func buildComputeAggregation(compute string) (*datadogV2.CIAppCompute, error) {
	if compute == "" || compute == "count" {
		return &datadogV2.CIAppCompute{
			Aggregation: datadogV2.CIAPPAGGREGATIONFUNCTION_COUNT,
		}, nil
	}

	parts := strings.SplitN(compute, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid compute format: %s (expected format: function:field)", compute)
	}

	function := parts[0]
	field := parts[1]
	var aggType datadogV2.CIAppAggregationFunction

	switch function {
	case "count":
		aggType = datadogV2.CIAPPAGGREGATIONFUNCTION_COUNT
	case "cardinality":
		aggType = datadogV2.CIAPPAGGREGATIONFUNCTION_CARDINALITY
	default:
		return nil, fmt.Errorf("unsupported aggregation function: %s (supported: count, cardinality)", function)
	}

	return &datadogV2.CIAppCompute{
		Aggregation: aggType,
		Metric:      &field,
	}, nil
}
