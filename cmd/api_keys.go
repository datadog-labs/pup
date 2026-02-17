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

var apiKeysCmd = &cobra.Command{
	Use:   "api-keys",
	Short: "Manage API keys",
	Long: `Manage Datadog API keys.

API keys authenticate requests to Datadog APIs. This command manages API keys
only (not application keys).

CAPABILITIES:
  • List API keys
  • Get API key details
  • Create new API keys
  • Update API keys (name only)
  • Delete API keys (requires confirmation)

EXAMPLES:
  # List all API keys
  pup api-keys list

  # Get API key details
  pup api-keys get key-id

  # Create new API key
  pup api-keys create --name="Production Key"

  # Delete an API key (with confirmation prompt)
  pup api-keys delete key-id

AUTHENTICATION:
  Requires OAuth2 (via 'pup auth login') or a valid API key + Application key
  combination. Note: You cannot use an API key to delete itself.`,
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
	Short: "Delete an API key (DESTRUCTIVE)",
	Long: `Delete an API key permanently.

WARNING: This is a destructive operation that cannot be undone. Deleting an API
key will immediately revoke access for any applications or services using it.

Before deleting, ensure:
  • No active services are using this key
  • You have alternative authentication configured
  • You cannot delete the API key currently being used for authentication

Use --auto-approve to skip the confirmation prompt (use with caution).`,
	Args: cobra.ExactArgs(1),
	RunE: runAPIKeysDelete,
}

var (
	apiKeyName string
)

func init() {
	apiKeysCreateCmd.Flags().StringVar(&apiKeyName, "name", "", "API key name (required)")
	if err := apiKeysCreateCmd.MarkFlagRequired("name"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}

	apiKeysCmd.AddCommand(apiKeysListCmd, apiKeysGetCmd, apiKeysCreateCmd, apiKeysDeleteCmd)
}

func runAPIKeysList(cmd *cobra.Command, args []string) error {
	// API Keys management doesn't support OAuth, use API keys
	client, err := getClientForEndpoint("GET", "/api/v2/api_keys")
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

	return formatAndPrint(resp, nil)
}

func runAPIKeysGet(cmd *cobra.Command, args []string) error {
	// API Keys management doesn't support OAuth, use API keys
	client, err := getClientForEndpoint("GET", "/api/v2/api_keys/")
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

	return formatAndPrint(resp, nil)
}

func runAPIKeysCreate(cmd *cobra.Command, args []string) error {
	// API Keys management doesn't support OAuth, use API keys
	client, err := getClientForEndpoint("POST", "/api/v2/api_keys")
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

	return formatAndPrint(resp, nil)
}

func runAPIKeysDelete(cmd *cobra.Command, args []string) error {
	// API Keys management doesn't support OAuth, use API keys
	client, err := getClientForEndpoint("DELETE", "/api/v2/api_keys/")
	if err != nil {
		return err
	}

	keyID := args[0]
	if !cfg.AutoApprove {
		printOutput("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		printOutput("⚠️  DESTRUCTIVE OPERATION WARNING ⚠️\n")
		printOutput("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		printOutput("\nYou are about to PERMANENTLY DELETE API key: %s\n", keyID)
		printOutput("\nThis action:\n")
		printOutput("  • Cannot be undone\n")
		printOutput("  • Will immediately revoke access for any services using this key\n")
		printOutput("  • May cause service disruptions if the key is in active use\n")
		printOutput("\nPlease confirm you have:\n")
		printOutput("  • Verified no active services depend on this key\n")
		printOutput("  • Documented or backed up the key information if needed\n")
		printOutput("\nType 'yes' to confirm deletion (or anything else to cancel): ")
		response, err := readConfirmation()
		if err != nil {
			// User cancelled or error reading input
			printOutput("\n✓ Operation cancelled\n")
			return nil
		}
		if response != "yes" {
			printOutput("✓ Operation cancelled\n")
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

	printOutput("Successfully deleted API key %s\n", keyID)
	return nil
}
