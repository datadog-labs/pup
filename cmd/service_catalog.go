// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/spf13/cobra"
)

var serviceCatalogCmd = &cobra.Command{
	Use:   "service-catalog",
	Short: "Manage service catalog",
	Long: `Manage services in the Datadog service catalog.

The service catalog provides a centralized registry of all services
in your infrastructure with ownership, dependencies, and documentation.

CAPABILITIES:
  • List services in the catalog
  • Get service details
  • View service definitions
  • Manage service metadata

EXAMPLES:
  # List all services
  pup service-catalog list

  # Get service details
  pup service-catalog get service-name

AUTHENTICATION:
  Requires either OAuth2 authentication or API keys.`,
}

var serviceCatalogListCmd = &cobra.Command{
	Use:   "list",
	Short: "List services",
	RunE:  runServiceCatalogList,
}

var serviceCatalogGetCmd = &cobra.Command{
	Use:   "get [service-name]",
	Short: "Get service details",
	Args:  cobra.ExactArgs(1),
	RunE:  runServiceCatalogGet,
}

func init() {
	serviceCatalogCmd.AddCommand(serviceCatalogListCmd, serviceCatalogGetCmd)
}

func runServiceCatalogList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewServiceDefinitionApi(client.V2())
	resp, r, err := api.ListServiceDefinitions(client.Context())
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to list services: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to list services: %w", err)
	}

	return formatAndPrint(resp, nil)
}

func runServiceCatalogGet(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	serviceName := args[0]
	api := datadogV2.NewServiceDefinitionApi(client.V2())
	resp, r, err := api.GetServiceDefinition(client.Context(), serviceName)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to get service: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to get service: %w", err)
	}

	return formatAndPrint(resp, nil)
}
