// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/DataDog/pup/pkg/formatter"
	"github.com/spf13/cobra"
)

var organizationsCmd = &cobra.Command{
	Use:   "organizations",
	Short: "Manage organization settings",
	Long: `Manage organization-level settings and configuration.

CAPABILITIES:
  • View organization details
  • List child organizations
  • Manage organization settings
  • Configure billing and usage

EXAMPLES:
  # Get organization details
  pup organizations get

  # List child organizations
  pup organizations list

AUTHENTICATION:
  Requires either OAuth2 authentication or API keys with org management permissions.`,
}

var organizationsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get organization details",
	RunE:  runOrganizationsGet,
}

var organizationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List organizations",
	RunE:  runOrganizationsList,
}

func init() {
	organizationsCmd.AddCommand(organizationsGetCmd, organizationsListCmd)
}

func runOrganizationsGet(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV1.NewOrganizationsApi(client.V1())
	resp, r, err := api.GetOrg(client.Context(), "current")
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to get organization: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to get organization: %w", err)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func runOrganizationsList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV1.NewOrganizationsApi(client.V1())
	resp, r, err := api.ListOrgs(client.Context())
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to list organizations: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to list organizations: %w", err)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}
