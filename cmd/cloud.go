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

var cloudCmd = &cobra.Command{
	Use:   "cloud",
	Short: "Manage cloud integrations",
	Long: `Manage cloud provider integrations (AWS, GCP, Azure).

Cloud integrations collect metrics and logs from your cloud providers
and provide insights into cloud resource usage and performance.

CAPABILITIES:
  • Manage AWS integrations
  • Manage GCP integrations
  • Manage Azure integrations
  • View cloud metrics

EXAMPLES:
  # List AWS integrations
  pup cloud aws list

  # List GCP integrations
  pup cloud gcp list

  # List Azure integrations
  pup cloud azure list

AUTHENTICATION:
  Requires either OAuth2 authentication or API keys.`,
}

var cloudAWSCmd = &cobra.Command{
	Use:   "aws",
	Short: "Manage AWS integrations",
}

var cloudAWSListCmd = &cobra.Command{
	Use:   "list",
	Short: "List AWS integrations",
	RunE:  runCloudAWSList,
}

var cloudGCPCmd = &cobra.Command{
	Use:   "gcp",
	Short: "Manage GCP integrations",
}

var cloudGCPListCmd = &cobra.Command{
	Use:   "list",
	Short: "List GCP integrations",
	RunE:  runCloudGCPList,
}

var cloudAzureCmd = &cobra.Command{
	Use:   "azure",
	Short: "Manage Azure integrations",
}

var cloudAzureListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Azure integrations",
	RunE:  runCloudAzureList,
}

func init() {
	cloudAWSCmd.AddCommand(cloudAWSListCmd)
	cloudGCPCmd.AddCommand(cloudGCPListCmd)
	cloudAzureCmd.AddCommand(cloudAzureListCmd)
	cloudCmd.AddCommand(cloudAWSCmd, cloudGCPCmd, cloudAzureCmd)
}

func runCloudAWSList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV1.NewAWSIntegrationApi(client.V1())
	resp, r, err := api.ListAWSAccounts(client.Context())
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to list AWS integrations: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to list AWS integrations: %w", err)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func runCloudGCPList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV1.NewGCPIntegrationApi(client.V1())
	resp, r, err := api.ListGCPIntegration(client.Context())
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to list GCP integrations: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to list GCP integrations: %w", err)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func runCloudAzureList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV1.NewAzureIntegrationApi(client.V1())
	resp, r, err := api.ListAzureIntegration(client.Context())
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to list Azure integrations: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to list Azure integrations: %w", err)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}
