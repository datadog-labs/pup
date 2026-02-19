// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
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

// OCI subcommands
var cloudOCICmd = &cobra.Command{
	Use:   "oci",
	Short: "Manage OCI integrations",
}

var cloudOCITenanciesCmd = &cobra.Command{
	Use:   "tenancies",
	Short: "Manage OCI tenancy configurations",
}

var cloudOCITenanciesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List OCI tenancy configurations",
	RunE:  runCloudOCITenanciesList,
}

var cloudOCITenanciesGetCmd = &cobra.Command{
	Use:   "get [tenancy-ocid]",
	Short: "Get OCI tenancy configuration",
	Args:  cobra.ExactArgs(1),
	RunE:  runCloudOCITenanciesGet,
}

var cloudOCITenanciesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create OCI tenancy configuration",
	RunE:  runCloudOCITenanciesCreate,
}

var cloudOCITenanciesUpdateCmd = &cobra.Command{
	Use:   "update [tenancy-ocid]",
	Short: "Update OCI tenancy configuration",
	Args:  cobra.ExactArgs(1),
	RunE:  runCloudOCITenanciesUpdate,
}

var cloudOCITenanciesDeleteCmd = &cobra.Command{
	Use:   "delete [tenancy-ocid]",
	Short: "Delete OCI tenancy configuration",
	Args:  cobra.ExactArgs(1),
	RunE:  runCloudOCITenanciesDelete,
}

var cloudOCIProductsCmd = &cobra.Command{
	Use:   "products",
	Short: "Manage OCI products",
}

var cloudOCIProductsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List OCI tenancy products",
	RunE:  runCloudOCIProductsList,
}

var (
	cloudOCIFile        string
	cloudOCIProductKeys string
)

func init() {
	cloudOCITenanciesCreateCmd.Flags().StringVar(&cloudOCIFile, "file", "", "JSON file with request body (required)")
	_ = cloudOCITenanciesCreateCmd.MarkFlagRequired("file")
	cloudOCITenanciesUpdateCmd.Flags().StringVar(&cloudOCIFile, "file", "", "JSON file with request body (required)")
	_ = cloudOCITenanciesUpdateCmd.MarkFlagRequired("file")
	cloudOCIProductsListCmd.Flags().StringVar(&cloudOCIProductKeys, "product-keys", "", "Comma-separated product keys (required)")
	_ = cloudOCIProductsListCmd.MarkFlagRequired("product-keys")

	cloudAWSCmd.AddCommand(cloudAWSListCmd)
	cloudGCPCmd.AddCommand(cloudGCPListCmd)
	cloudAzureCmd.AddCommand(cloudAzureListCmd)

	cloudOCITenanciesCmd.AddCommand(cloudOCITenanciesListCmd, cloudOCITenanciesGetCmd, cloudOCITenanciesCreateCmd, cloudOCITenanciesUpdateCmd, cloudOCITenanciesDeleteCmd)
	cloudOCIProductsCmd.AddCommand(cloudOCIProductsListCmd)
	cloudOCICmd.AddCommand(cloudOCITenanciesCmd, cloudOCIProductsCmd)

	cloudCmd.AddCommand(cloudAWSCmd, cloudGCPCmd, cloudAzureCmd, cloudOCICmd)
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

	return formatAndPrint(resp, nil)
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

	return formatAndPrint(resp, nil)
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

	return formatAndPrint(resp, nil)
}

// OCI implementations
func runCloudOCITenanciesList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewOCIIntegrationApi(client.V2())
	resp, r, err := api.GetTenancyConfigs(client.Context())
	if err != nil {
		return formatAPIError("list OCI tenancies", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runCloudOCITenanciesGet(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewOCIIntegrationApi(client.V2())
	resp, r, err := api.GetTenancyConfig(client.Context(), args[0])
	if err != nil {
		return formatAPIError("get OCI tenancy", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runCloudOCITenanciesCreate(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile(cloudOCIFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var body datadogV2.CreateTenancyConfigRequest
	if err := json.Unmarshal(data, &body); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewOCIIntegrationApi(client.V2())
	resp, r, err := api.CreateTenancyConfig(client.Context(), body)
	if err != nil {
		return formatAPIError("create OCI tenancy", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runCloudOCITenanciesUpdate(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile(cloudOCIFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var body datadogV2.UpdateTenancyConfigRequest
	if err := json.Unmarshal(data, &body); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewOCIIntegrationApi(client.V2())
	resp, r, err := api.UpdateTenancyConfig(client.Context(), args[0], body)
	if err != nil {
		return formatAPIError("update OCI tenancy", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runCloudOCITenanciesDelete(cmd *cobra.Command, args []string) error {
	if !cfg.AutoApprove {
		printOutput("WARNING: This will permanently delete OCI tenancy '%s'.\n", args[0])
		printOutput("Are you sure you want to continue? [y/N]: ")
		response, err := readConfirmation()
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}
		if response != "y" && response != "Y" && response != "yes" {
			printOutput("Operation cancelled.\n")
			return nil
		}
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewOCIIntegrationApi(client.V2())
	r, err := api.DeleteTenancyConfig(client.Context(), args[0])
	if err != nil {
		return formatAPIError("delete OCI tenancy", err, r)
	}

	printOutput("OCI tenancy '%s' deleted successfully.\n", args[0])
	return nil
}

func runCloudOCIProductsList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewOCIIntegrationApi(client.V2())
	resp, r, err := api.ListTenancyProducts(client.Context(), cloudOCIProductKeys)
	if err != nil {
		return formatAPIError("list OCI products", err, r)
	}

	return formatAndPrint(resp, nil)
}
