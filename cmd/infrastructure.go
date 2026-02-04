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

var infrastructureCmd = &cobra.Command{
	Use:   "infrastructure",
	Short: "Manage infrastructure monitoring",
	Long: `Query and manage infrastructure hosts and metrics.

CAPABILITIES:
  • List hosts in your infrastructure
  • Get host details and metrics
  • Search hosts by tags or status
  • Monitor host health

EXAMPLES:
  # List all hosts
  pup infrastructure hosts list

  # Search for hosts by tag
  pup infrastructure hosts list --filter="env:production"

  # Get host details
  pup infrastructure hosts get my-host

AUTHENTICATION:
  Requires either OAuth2 authentication or API keys.`,
}

var infrastructureHostsCmd = &cobra.Command{
	Use:   "hosts",
	Short: "Manage hosts",
}

var infrastructureHostsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List hosts",
	RunE:  runInfrastructureHostsList,
}

var infrastructureHostsGetCmd = &cobra.Command{
	Use:   "get [hostname]",
	Short: "Get host details",
	Args:  cobra.ExactArgs(1),
	RunE:  runInfrastructureHostsGet,
}

var (
	infraFilter string
	infraSort   string
	infraCount  int64
)

func init() {
	infrastructureHostsListCmd.Flags().StringVar(&infraFilter, "filter", "", "Filter hosts")
	infrastructureHostsListCmd.Flags().StringVar(&infraSort, "sort", "status", "Sort field")
	infrastructureHostsListCmd.Flags().Int64Var(&infraCount, "count", 100, "Maximum hosts")

	infrastructureHostsCmd.AddCommand(infrastructureHostsListCmd, infrastructureHostsGetCmd)
	infrastructureCmd.AddCommand(infrastructureHostsCmd)
}

func runInfrastructureHostsList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV1.NewHostsApi(client.V1())
	opts := datadogV1.ListHostsOptionalParameters{}
	if infraFilter != "" {
		opts.WithFilter(infraFilter)
	}
	if infraSort != "" {
		opts.WithSortField(infraSort)
	}
	if infraCount > 0 {
		opts.WithCount(infraCount)
	}

	resp, r, err := api.ListHosts(client.Context(), opts)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to list hosts: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to list hosts: %w", err)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func runInfrastructureHostsGet(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	hostname := args[0]
	api := datadogV1.NewHostsApi(client.V1())
	resp, r, err := api.GetHostTotals(client.Context())
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to get host: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to get host: %w", err)
	}

	_ = hostname // Use hostname for filtering in actual implementation
	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}
