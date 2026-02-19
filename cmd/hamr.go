// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/spf13/cobra"
)

var hamrCmd = &cobra.Command{
	Use:   "hamr",
	Short: "Manage High Availability Multi-Region (HAMR)",
	Long: `Manage Datadog High Availability Multi-Region (HAMR) connections.

HAMR provides high availability and multi-region failover capabilities
for your Datadog organization.

EXAMPLES:
  # Get HAMR connection status
  pup hamr connections get

  # Create a HAMR connection
  pup hamr connections create --file=connection.json

AUTHENTICATION:
  Requires either OAuth2 authentication or API keys.`,
}

var hamrConnectionsCmd = &cobra.Command{
	Use:   "connections",
	Short: "Manage HAMR organization connections",
}

var hamrConnectionsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get HAMR organization connection",
	RunE:  runHamrConnectionsGet,
}

var hamrConnectionsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create HAMR organization connection",
	RunE:  runHamrConnectionsCreate,
}

var (
	hamrFile string
)

func init() {
	hamrConnectionsCreateCmd.Flags().StringVar(&hamrFile, "file", "", "JSON file with request body (required)")
	_ = hamrConnectionsCreateCmd.MarkFlagRequired("file")

	hamrConnectionsCmd.AddCommand(hamrConnectionsGetCmd, hamrConnectionsCreateCmd)
	hamrCmd.AddCommand(hamrConnectionsCmd)
}

func runHamrConnectionsGet(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewHighAvailabilityMultiRegionApi(client.V2())
	resp, r, err := api.GetHamrOrgConnection(client.Context())
	if err != nil {
		return formatAPIError("get HAMR connection", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runHamrConnectionsCreate(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile(hamrFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var body datadogV2.HamrOrgConnectionRequest
	if err := json.Unmarshal(data, &body); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewHighAvailabilityMultiRegionApi(client.V2())
	resp, r, err := api.CreateHamrOrgConnection(client.Context(), body)
	if err != nil {
		return formatAPIError("create HAMR connection", err, r)
	}

	return formatAndPrint(resp, nil)
}
