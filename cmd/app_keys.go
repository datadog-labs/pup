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

var appKeysCmd = &cobra.Command{
	Use:   "app-keys",
	Short: "Manage application keys",
	Long: `Manage Datadog application keys.

Application keys, in conjunction with your organization's API key, give you
full access to Datadog's API. Application keys are associated with the user
who created them and have the same permissions and scopes as the user.

CAPABILITIES:
  • List your application keys (or all org keys with --all)
  • Get application key details
  • Create new application keys (with optional scopes)
  • Update application key name or scopes
  • Delete application keys (requires confirmation)

EXAMPLES:
  # List your application keys
  pup app-keys list

  # List all application keys in the org (requires API keys)
  pup app-keys list --all

  # Get application key details
  pup app-keys get <app-key-id>

  # Create a new application key
  pup app-keys create --name="My Key"

  # Create a scoped application key
  pup app-keys create --name="Read Only" --scopes="dashboards_read,metrics_read"

  # Update an application key name
  pup app-keys update <app-key-id> --name="New Name"

  # Delete an application key
  pup app-keys delete <app-key-id>

AUTHENTICATION:
  Most commands use the current_user endpoints and support OAuth2 (via
  'pup auth login'). The 'list --all' command uses the org-wide endpoint
  and requires API + Application keys (DD_API_KEY + DD_APP_KEY).`,
}

var appKeysListCmd = &cobra.Command{
	Use:   "list",
	Short: "List application keys",
	Long: `List application keys.

By default, lists your own application keys (supports OAuth2).
Use --all to list all application keys in the org (requires API keys).`,
	RunE: runAppKeysList,
}

var appKeysGetCmd = &cobra.Command{
	Use:   "get [app-key-id]",
	Short: "Get application key details",
	Long: `Get details for a specific application key by its ID.

Returns full key details including the key value (if not using OTR mode).`,
	Args: cobra.ExactArgs(1),
	RunE: runAppKeysGet,
}

var appKeysCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new application key",
	Long: `Create a new application key for the current user.

The key value is returned in the response. If your org has One-Time Read (OTR)
mode enabled, this is the only time the full key value will be shown.

EXAMPLES:
  # Create a basic application key
  pup app-keys create --name="My Key"

  # Create a scoped application key
  pup app-keys create --name="Read Only" --scopes="dashboards_read,metrics_read"

  # Create and output as JSON
  pup app-keys create --name="CI Key" -o json`,
	RunE: runAppKeysCreate,
}

var appKeysUpdateCmd = &cobra.Command{
	Use:   "update [app-key-id]",
	Short: "Update an application key",
	Long: `Update an application key's name or scopes.

EXAMPLES:
  # Update the name
  pup app-keys update <app-key-id> --name="Updated Name"

  # Update scopes
  pup app-keys update <app-key-id> --scopes="dashboards_read"`,
	Args: cobra.ExactArgs(1),
	RunE: runAppKeysUpdate,
}

var appKeysDeleteCmd = &cobra.Command{
	Use:   "delete [app-key-id]",
	Short: "Delete an application key (DESTRUCTIVE)",
	Long: `Delete an application key permanently.

WARNING: This is a destructive operation that cannot be undone. Deleting an
application key will immediately revoke access for any applications or services
using it.

Before deleting, ensure:
  • No active services are using this key
  • You have alternative authentication configured

Use --auto-approve to skip the confirmation prompt (use with caution).`,
	Args: cobra.ExactArgs(1),
	RunE: runAppKeysDelete,
}

var (
	appKeysPageSize   int64
	appKeysPageNumber int64
	appKeysFilter     string
	appKeysSort       string
	appKeysListAll    bool
	appKeyName        string
	appKeyScopes      string
)

func init() {
	appKeysListCmd.Flags().Int64Var(&appKeysPageSize, "page-size", 10, "Number of results per page")
	appKeysListCmd.Flags().Int64Var(&appKeysPageNumber, "page-number", 0, "Page number to retrieve (0-indexed)")
	appKeysListCmd.Flags().StringVar(&appKeysFilter, "filter", "", "Filter by key name")
	appKeysListCmd.Flags().StringVar(&appKeysSort, "sort", "", "Sort field (name, -name, created_at, -created_at)")
	appKeysListCmd.Flags().BoolVar(&appKeysListAll, "all", false, "List all org keys (requires API keys, not OAuth)")

	appKeysCreateCmd.Flags().StringVar(&appKeyName, "name", "", "Application key name (required)")
	if err := appKeysCreateCmd.MarkFlagRequired("name"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}
	appKeysCreateCmd.Flags().StringVar(&appKeyScopes, "scopes", "", "Comma-separated authorization scopes (e.g. dashboards_read,metrics_read)")

	appKeysUpdateCmd.Flags().StringVar(&appKeyName, "name", "", "New name for the application key")
	appKeysUpdateCmd.Flags().StringVar(&appKeyScopes, "scopes", "", "Comma-separated authorization scopes")

	appKeysCmd.AddCommand(appKeysListCmd, appKeysGetCmd, appKeysCreateCmd, appKeysUpdateCmd, appKeysDeleteCmd)
}

func runAppKeysList(cmd *cobra.Command, args []string) error {
	if appKeysListAll {
		return runAppKeysListAll(cmd, args)
	}
	return runAppKeysListCurrentUser(cmd, args)
}

func runAppKeysListCurrentUser(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewKeyManagementApi(client.V2())
	opts := datadogV2.NewListCurrentUserApplicationKeysOptionalParameters()

	if appKeysPageSize > 0 {
		opts = opts.WithPageSize(appKeysPageSize)
	}
	if appKeysPageNumber > 0 {
		opts = opts.WithPageNumber(appKeysPageNumber)
	}
	if appKeysFilter != "" {
		opts = opts.WithFilter(appKeysFilter)
	}
	if appKeysSort != "" {
		opts = opts.WithSort(datadogV2.ApplicationKeysSort(appKeysSort))
	}

	resp, r, err := api.ListCurrentUserApplicationKeys(client.Context(), *opts)
	if err != nil {
		return formatAPIError("list application keys", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runAppKeysListAll(cmd *cobra.Command, args []string) error {
	client, err := getClientForEndpoint("GET", "/api/v2/application_keys")
	if err != nil {
		return err
	}

	api := datadogV2.NewKeyManagementApi(client.V2())
	opts := datadogV2.NewListApplicationKeysOptionalParameters()

	if appKeysPageSize > 0 {
		opts = opts.WithPageSize(appKeysPageSize)
	}
	if appKeysPageNumber > 0 {
		opts = opts.WithPageNumber(appKeysPageNumber)
	}
	if appKeysFilter != "" {
		opts = opts.WithFilter(appKeysFilter)
	}
	if appKeysSort != "" {
		opts = opts.WithSort(datadogV2.ApplicationKeysSort(appKeysSort))
	}

	resp, r, err := api.ListApplicationKeys(client.Context(), *opts)
	if err != nil {
		return formatAPIError("list all application keys", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runAppKeysGet(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	appKeyID := args[0]
	api := datadogV2.NewKeyManagementApi(client.V2())

	resp, r, err := api.GetCurrentUserApplicationKey(client.Context(), appKeyID)
	if err != nil {
		return formatAPIError("get application key", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runAppKeysCreate(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewKeyManagementApi(client.V2())

	attrs := datadogV2.ApplicationKeyCreateAttributes{
		Name: appKeyName,
	}

	if appKeyScopes != "" {
		scopes := splitAndTrim(appKeyScopes)
		attrs.SetScopes(scopes)
	}

	body := datadogV2.ApplicationKeyCreateRequest{
		Data: datadogV2.ApplicationKeyCreateData{
			Attributes: attrs,
			Type:       datadogV2.APPLICATIONKEYSTYPE_APPLICATION_KEYS,
		},
	}

	resp, r, err := api.CreateCurrentUserApplicationKey(client.Context(), body)
	if err != nil {
		return formatAPIError("create application key", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runAppKeysUpdate(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	appKeyID := args[0]
	api := datadogV2.NewKeyManagementApi(client.V2())

	attrs := datadogV2.ApplicationKeyUpdateAttributes{}

	if appKeyName != "" {
		attrs.SetName(appKeyName)
	}
	if appKeyScopes != "" {
		scopes := splitAndTrim(appKeyScopes)
		attrs.SetScopes(scopes)
	}

	body := datadogV2.ApplicationKeyUpdateRequest{
		Data: datadogV2.ApplicationKeyUpdateData{
			Attributes: attrs,
			Id:         appKeyID,
			Type:       datadogV2.APPLICATIONKEYSTYPE_APPLICATION_KEYS,
		},
	}

	resp, r, err := api.UpdateCurrentUserApplicationKey(client.Context(), appKeyID, body)
	if err != nil {
		return formatAPIError("update application key", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runAppKeysDelete(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	appKeyID := args[0]
	if !cfg.AutoApprove {
		printOutput("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		printOutput("⚠️  DESTRUCTIVE OPERATION WARNING ⚠️\n")
		printOutput("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		printOutput("\nYou are about to PERMANENTLY DELETE application key: %s\n", appKeyID)
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
			printOutput("\n✓ Operation cancelled\n")
			return nil
		}
		if response != "yes" {
			printOutput("✓ Operation cancelled\n")
			return nil
		}
	}

	api := datadogV2.NewKeyManagementApi(client.V2())
	r, err := api.DeleteCurrentUserApplicationKey(client.Context(), appKeyID)
	if err != nil {
		return formatAPIError("delete application key", err, r)
	}

	printOutput("Successfully deleted application key %s\n", appKeyID)
	return nil
}
