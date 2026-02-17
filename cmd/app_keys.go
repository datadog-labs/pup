// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/spf13/cobra"
)

var appKeysCmd = &cobra.Command{
	Use:   "app-keys",
	Short: "Manage app key registrations",
	Long: `Manage Datadog app key registrations for Action Connections.

App key registrations enable application keys to be used with Action Connections
and Workflow Automation features. This is separate from standard application key
management (see 'pup api-keys' for that).

CAPABILITIES:
  • List registered app keys
  • Get app key registration details
  • Register an application key for Action Connections
  • Unregister an application key from Action Connections

EXAMPLES:
  # List all registered app keys
  pup app-keys list

  # Get app key registration details
  pup app-keys get <app-key-id>

  # Register an application key
  pup app-keys register <app-key-id>

  # Unregister an application key
  pup app-keys unregister <app-key-id>

AUTHENTICATION:
  Requires OAuth2 (via 'pup auth login') or valid API + Application keys.`,
}

var appKeysListCmd = &cobra.Command{
	Use:   "list",
	Short: "List registered app keys",
	Long: `List all app keys registered for Action Connections.

Returns a paginated list of app key registrations with their IDs and types.`,
	RunE: runAppKeysList,
}

var appKeysGetCmd = &cobra.Command{
	Use:   "get [app-key-id]",
	Short: "Get app key registration details",
	Long: `Get details for a specific app key registration by its ID.

The app-key-id is the UUID of the registered application key.`,
	Args: cobra.ExactArgs(1),
	RunE: runAppKeysGet,
}

var appKeysRegisterCmd = &cobra.Command{
	Use:   "register [app-key-id]",
	Short: "Register an application key",
	Long: `Register an existing application key for use with Action Connections.

This enables the application key to be used in workflow automation and
Action Connection features. The app-key-id must be the ID of an existing
application key (see 'pup api-keys list' to view application keys).

EXAMPLES:
  # Register an application key
  pup app-keys register abc-123-def-456

  # Register with JSON output
  pup app-keys register abc-123-def-456 -o json`,
	Args: cobra.ExactArgs(1),
	RunE: runAppKeysRegister,
}

var appKeysUnregisterCmd = &cobra.Command{
	Use:   "unregister [app-key-id]",
	Short: "Unregister an application key",
	Long: `Unregister an application key from Action Connections (DESTRUCTIVE).

WARNING: This will remove the app key registration, preventing it from being
used with Action Connections and workflow automation features. The underlying
application key itself will NOT be deleted.

Before unregistering, ensure:
  • No active Action Connections are using this key
  • No workflows depend on this registration

Use --auto-approve to skip the confirmation prompt (use with caution).`,
	Args: cobra.ExactArgs(1),
	RunE: runAppKeysUnregister,
}

var (
	appKeysPageSize   int64
	appKeysPageNumber int64
)

func init() {
	appKeysListCmd.Flags().Int64Var(&appKeysPageSize, "page-size", 10, "Number of results per page")
	appKeysListCmd.Flags().Int64Var(&appKeysPageNumber, "page-number", 0, "Page number to retrieve (0-indexed)")

	appKeysCmd.AddCommand(appKeysListCmd, appKeysGetCmd, appKeysRegisterCmd, appKeysUnregisterCmd)
}

func runAppKeysList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewActionConnectionApi(client.V2())
	opts := datadogV2.ListAppKeyRegistrationsOptionalParameters{}

	if appKeysPageSize > 0 {
		opts.WithPageSize(appKeysPageSize)
	}
	if appKeysPageNumber > 0 {
		opts.WithPageNumber(appKeysPageNumber)
	}

	resp, r, err := api.ListAppKeyRegistrations(client.Context(), opts)
	if err != nil {
		return formatAPIError("list app key registrations", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runAppKeysGet(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	appKeyID := args[0]
	api := datadogV2.NewActionConnectionApi(client.V2())

	resp, r, err := api.GetAppKeyRegistration(client.Context(), appKeyID)
	if err != nil {
		return formatAPIError("get app key registration", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runAppKeysRegister(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	appKeyID := args[0]
	api := datadogV2.NewActionConnectionApi(client.V2())

	resp, r, err := api.RegisterAppKey(client.Context(), appKeyID)
	if err != nil {
		return formatAPIError("register app key", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runAppKeysUnregister(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	appKeyID := args[0]
	if !cfg.AutoApprove {
		printOutput("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		printOutput("⚠️  DESTRUCTIVE OPERATION WARNING ⚠️\n")
		printOutput("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		printOutput("\nYou are about to UNREGISTER app key: %s\n", appKeyID)
		printOutput("\nThis action:\n")
		printOutput("  • Will remove the app key registration\n")
		printOutput("  • May affect Action Connections using this key\n")
		printOutput("  • Cannot be undone (must re-register if needed)\n")
		printOutput("  • Does NOT delete the underlying application key\n")
		printOutput("\nType 'yes' to confirm unregistration: ")

		response, err := readConfirmation()
		if err != nil || response != "yes" {
			printOutput("\n✓ Operation cancelled\n")
			return nil
		}
	}

	api := datadogV2.NewActionConnectionApi(client.V2())
	r, err := api.UnregisterAppKey(client.Context(), appKeyID)
	if err != nil {
		return formatAPIError("unregister app key", err, r)
	}

	printOutput("Successfully unregistered app key %s\n", appKeyID)
	return nil
}
