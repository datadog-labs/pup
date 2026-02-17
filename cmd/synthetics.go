// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/spf13/cobra"
)

var syntheticsCmd = &cobra.Command{
	Use:   "synthetics",
	Short: "Manage synthetic monitoring",
	Long: `Manage synthetic tests for monitoring application availability.

Synthetic monitoring simulates user interactions and API requests to
monitor application performance and availability from various locations.

CAPABILITIES:
  • List synthetic tests
  • Search synthetic tests by text query
  • Get test details
  • Get test results
  • List test locations
  • Manage global variables

EXAMPLES:
  # List all synthetic tests
  pup synthetics tests list

  # Search tests by creator or team
  pup synthetics tests search --text='creator:"Jane Doe"'
  pup synthetics tests search --text="team:my-team"

  # Get test details
  pup synthetics tests get test-id

  # List available locations
  pup synthetics locations list

AUTHENTICATION:
  Requires either OAuth2 authentication or API keys.`,
}

var syntheticsTestsCmd = &cobra.Command{
	Use:   "tests",
	Short: "Manage synthetic tests",
}

var syntheticsTestsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List synthetic tests",
	RunE:  runSyntheticsTestsList,
}

var syntheticsTestsGetCmd = &cobra.Command{
	Use:   "get [test-id]",
	Short: "Get test details",
	Args:  cobra.ExactArgs(1),
	RunE:  runSyntheticsTestsGet,
}

var syntheticsTestsSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search synthetic tests",
	Long: `Search synthetic tests using a text query.

Search allows filtering tests by free text or by facet queries, returning
matching tests with pagination support.

QUERY SYNTAX:
  Free text matches against test names. Facet queries use the format
  facet:value or facet:"multi word value".

  Common facets:
    creator              Test creator name (e.g., creator:"Jane Doe")
    team                 Team tag (e.g., team:synthetics)
    env                  Environment tag (e.g., env:prod, env:staging)
    type                 Test type (api, api-multi, api-ssl, api-tcp, browser)
    state                Test state (live, paused)
    status               Monitor status (OK, Alert, "No Data")
    tag                  Any tag (e.g., tag:terraform:true)
    region               Test location (e.g., region:aws:us-east-2)
    domain               Target domain
    http_method          HTTP method (GET, POST, DELETE, PATCH)
    http_path            Request path
    muted                Muted state (0, 1)
    ci_execution_rule    CI rule (blocking, non_blocking, skipped)
    creation_source      How the test was created (e.g., terraform, templates)
    mobile_platform      Mobile platform (android, ios)
    notification         Notification handle
    endpoint             Full endpoint URL
    step_count           Number of test steps

  Use --facets-only to discover all available facets and their values
  for your organization.

FLAGS:
  --text                 Search text or facet query
  --include-full-config  Include full test configuration in results
  --facets-only          Return only facets (no test results)
  --start                Pagination offset (default: 0)
  --count                Number of results to return (default: 50)
  --sort                 Sort order (e.g., "name,asc")

EXAMPLES:
  # Search tests by name
  pup synthetics tests search --text="checkout"

  # Find tests by creator
  pup synthetics tests search --text='creator:"Jane Doe"'

  # Find tests for a team
  pup synthetics tests search --text="team:my-team"

  # Filter by type and environment
  pup synthetics tests search --text="type:browser env:prod"

  # Search with pagination
  pup synthetics tests search --text="api" --start=0 --count=100

  # Discover available facets and values for your org
  pup synthetics tests search --facets-only`,
	RunE: runSyntheticsTestsSearch,
}

var (
	syntheticsSearchText              string
	syntheticsSearchIncludeFullConfig bool
	syntheticsSearchFacetsOnly        bool
	syntheticsSearchStart             int64
	syntheticsSearchCount             int64
	syntheticsSearchSort              string
)

var syntheticsLocationsCmd = &cobra.Command{
	Use:   "locations",
	Short: "Manage test locations",
}

var syntheticsLocationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available locations",
	RunE:  runSyntheticsLocationsList,
}

func init() {
	syntheticsTestsSearchCmd.Flags().StringVar(&syntheticsSearchText, "text", "", "Search text query")
	syntheticsTestsSearchCmd.Flags().BoolVar(&syntheticsSearchIncludeFullConfig, "include-full-config", false, "Include full test configuration in results")
	syntheticsTestsSearchCmd.Flags().BoolVar(&syntheticsSearchFacetsOnly, "facets-only", false, "Return only facets (no test results)")
	syntheticsTestsSearchCmd.Flags().Int64Var(&syntheticsSearchStart, "start", 0, "Pagination offset")
	syntheticsTestsSearchCmd.Flags().Int64Var(&syntheticsSearchCount, "count", 50, "Number of results to return")
	syntheticsTestsSearchCmd.Flags().StringVar(&syntheticsSearchSort, "sort", "", "Sort order")

	syntheticsTestsCmd.AddCommand(syntheticsTestsListCmd, syntheticsTestsGetCmd, syntheticsTestsSearchCmd)
	syntheticsLocationsCmd.AddCommand(syntheticsLocationsListCmd)
	syntheticsCmd.AddCommand(syntheticsTestsCmd, syntheticsLocationsCmd)
}

func runSyntheticsTestsList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV1.NewSyntheticsApi(client.V1())
	resp, r, err := api.ListTests(client.Context())
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to list synthetic tests: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to list synthetic tests: %w", err)
	}

	return formatAndPrint(resp, nil)
}

func runSyntheticsTestsGet(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	testID := args[0]
	api := datadogV1.NewSyntheticsApi(client.V1())
	resp, r, err := api.GetTest(client.Context(), testID)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to get synthetic test: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to get synthetic test: %w", err)
	}

	return formatAndPrint(resp, nil)
}

func runSyntheticsTestsSearch(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV1.NewSyntheticsApi(client.V1())
	opts := datadogV1.SearchTestsOptionalParameters{}

	if syntheticsSearchText != "" {
		opts.WithText(syntheticsSearchText)
	}
	if syntheticsSearchIncludeFullConfig {
		opts.WithIncludeFullConfig(syntheticsSearchIncludeFullConfig)
	}
	if syntheticsSearchFacetsOnly {
		opts.WithFacetsOnly(syntheticsSearchFacetsOnly)
	}
	if cmd.Flags().Changed("start") {
		opts.WithStart(syntheticsSearchStart)
	}
	if cmd.Flags().Changed("count") {
		opts.WithCount(syntheticsSearchCount)
	}
	if syntheticsSearchSort != "" {
		opts.WithSort(syntheticsSearchSort)
	}

	resp, r, err := api.SearchTests(client.Context(), opts)
	if err != nil {
		return formatAPIError("search synthetic tests", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runSyntheticsLocationsList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV1.NewSyntheticsApi(client.V1())
	resp, r, err := api.ListLocations(client.Context())
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to list locations: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to list locations: %w", err)
	}

	return formatAndPrint(resp, nil)
}
