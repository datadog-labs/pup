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
  • Get test details
  • Get test results
  • List test locations
  • Manage global variables

EXAMPLES:
  # List all synthetic tests
  pup synthetics tests list

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
	syntheticsTestsCmd.AddCommand(syntheticsTestsListCmd, syntheticsTestsGetCmd)
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
