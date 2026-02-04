// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/DataDog/pup/pkg/formatter"
	"github.com/spf13/cobra"
)

var downtimeCmd = &cobra.Command{
	Use:   "downtime",
	Short: "Manage monitor downtimes",
	Long: `Manage downtimes to silence monitors during maintenance windows.

Downtimes prevent monitors from alerting during scheduled maintenance,
deployments, or other planned events.

CAPABILITIES:
  • List all downtimes
  • Get downtime details
  • Create new downtimes
  • Update existing downtimes
  • Cancel downtimes

EXAMPLES:
  # List all active downtimes
  pup downtime list

  # Get downtime details
  pup downtime get abc-123-def

  # Cancel a downtime
  pup downtime cancel abc-123-def

AUTHENTICATION:
  Requires either OAuth2 authentication or API keys.`,
}

var downtimeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all downtimes",
	RunE:  runDowntimeList,
}

var downtimeGetCmd = &cobra.Command{
	Use:   "get [downtime-id]",
	Short: "Get downtime details",
	Args:  cobra.ExactArgs(1),
	RunE:  runDowntimeGet,
}

var downtimeCancelCmd = &cobra.Command{
	Use:   "cancel [downtime-id]",
	Short: "Cancel a downtime",
	Args:  cobra.ExactArgs(1),
	RunE:  runDowntimeCancel,
}

func init() {
	downtimeCmd.AddCommand(downtimeListCmd, downtimeGetCmd, downtimeCancelCmd)
}

func runDowntimeList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewDowntimesApi(client.V2())
	resp, r, err := api.ListDowntimes(client.Context())
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to list downtimes: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to list downtimes: %w", err)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func runDowntimeGet(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	downtimeID := args[0]
	api := datadogV2.NewDowntimesApi(client.V2())
	resp, r, err := api.GetDowntime(client.Context(), downtimeID)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to get downtime: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to get downtime: %w", err)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func runDowntimeCancel(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	downtimeID := args[0]
	if !cfg.AutoApprove {
		fmt.Printf("⚠️  WARNING: This will cancel downtime %s\n", downtimeID)
		fmt.Print("Are you sure you want to continue? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Operation cancelled")
			return nil
		}
	}

	api := datadogV2.NewDowntimesApi(client.V2())
	r, err := api.CancelDowntime(client.Context(), downtimeID)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to cancel downtime: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to cancel downtime: %w", err)
	}

	fmt.Printf("Successfully cancelled downtime %s\n", downtimeID)
	return nil
}
