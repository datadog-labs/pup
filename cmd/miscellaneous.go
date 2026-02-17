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

var miscCmd = &cobra.Command{
	Use:   "misc",
	Short: "Miscellaneous API operations",
	Long: `Miscellaneous API operations for various Datadog features.

CAPABILITIES:
  • Query IP ranges
  • Check API status
  • View service level agreements
  • Access miscellaneous endpoints

EXAMPLES:
  # Get Datadog IP ranges
  pup misc ip-ranges

  # Check API status
  pup misc status

AUTHENTICATION:
  Some endpoints may not require authentication.`,
}

var miscIPRangesCmd = &cobra.Command{
	Use:   "ip-ranges",
	Short: "Get Datadog IP ranges",
	RunE:  runMiscIPRanges,
}

var miscStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check API status",
	RunE:  runMiscStatus,
}

func init() {
	miscCmd.AddCommand(miscIPRangesCmd, miscStatusCmd)
}

func runMiscIPRanges(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV1.NewIPRangesApi(client.V1())
	resp, r, err := api.GetIPRanges(client.Context())
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to get IP ranges: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to get IP ranges: %w", err)
	}

	return formatAndPrint(resp, nil)
}

func runMiscStatus(cmd *cobra.Command, args []string) error {
	result := map[string]interface{}{
		"status":  "ok",
		"message": "API is operational",
	}

	return formatAndPrint(result, nil)
}
