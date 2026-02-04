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

var apiKeysCmd = &cobra.Command{
	Use:   "api-keys",
	Short: "Manage API keys",
	Long: `Manage Datadog API keys and application keys.

API keys authenticate requests to Datadog APIs. Application keys provide
additional authentication for writing data.

CAPABILITIES:
  • List API keys
  • Get API key details
  • Create new API keys
  • Update API keys
  • Delete API keys

EXAMPLES:
  # List all API keys
  pup api-keys list

  # Get API key details
  pup api-keys get key-id

  # Create new API key
  pup api-keys create --name="Production Key"

AUTHENTICATION:
  Requires either OAuth2 authentication or existing API keys.`,
}

var apiKeysListCmd = &cobra.Command{
	Use:   "list",
	Short: "List API keys",
	RunE:  runAPIKeysList,
}

var apiKeysGetCmd = &cobra.Command{
	Use:   "get [key-id]",
	Short: "Get API key details",
	Args:  cobra.ExactArgs(1),
	RunE:  runAPIKeysGet,
}

var apiKeysCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create new API key",
	RunE:  runAPIKeysCreate,
}

var apiKeysDeleteCmd = &cobra.Command{
	Use:   "delete [key-id]",
	Short: "Delete an API key",
	Args:  cobra.ExactArgs(1),
	RunE:  runAPIKeysDelete,
}

var (
	apiKeyName string
)

func init() {
	apiKeysCreateCmd.Flags().StringVar(&apiKeyName, "name", "", "API key name (required)")
	apiKeysCreateCmd.MarkFlagRequired("name")

	apiKeysCmd.AddCommand(apiKeysListCmd, apiKeysGetCmd, apiKeysCreateCmd, apiKeysDeleteCmd)
}

func runAPIKeysList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewKeyManagementApi(client.V2())
	resp, r, err := api.ListAPIKeys(client.Context())
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to list API keys: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to list API keys: %w", err)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func runAPIKeysGet(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	keyID := args[0]
	api := datadogV2.NewKeyManagementApi(client.V2())
	resp, r, err := api.GetAPIKey(client.Context(), keyID)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to get API key: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to get API key: %w", err)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func runAPIKeysCreate(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewKeyManagementApi(client.V2())
	body := datadogV2.APIKeyCreateRequest{
		Data: datadogV2.APIKeyCreateData{
			Attributes: datadogV2.APIKeyCreateAttributes{
				Name: apiKeyName,
			},
			Type: datadogV2.APIKEYSTYPE_API_KEYS,
		},
	}

	resp, r, err := api.CreateAPIKey(client.Context(), body)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to create API key: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to create API key: %w", err)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func runAPIKeysDelete(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	keyID := args[0]
	if !cfg.AutoApprove {
		fmt.Printf("⚠️  WARNING: This will permanently delete API key %s\n", keyID)
		fmt.Print("Are you sure you want to continue? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Operation cancelled")
			return nil
		}
	}

	api := datadogV2.NewKeyManagementApi(client.V2())
	r, err := api.DeleteAPIKey(client.Context(), keyID)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to delete API key: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to delete API key: %w", err)
	}

	fmt.Printf("Successfully deleted API key %s\n", keyID)
	return nil
}
