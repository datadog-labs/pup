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

var usersCmd = &cobra.Command{
	Use:   "users",
	Short: "Manage users and access",
	Long: `Manage users, roles, and access permissions.

CAPABILITIES:
  • List users in your organization
  • Get user details
  • Manage user roles and permissions
  • Invite new users
  • Disable users

EXAMPLES:
  # List all users
  pup users list

  # Get user details
  pup users get user-id

  # List roles
  pup users roles list

AUTHENTICATION:
  Requires either OAuth2 authentication or API keys.`,
}

var usersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List users",
	RunE:  runUsersList,
}

var usersGetCmd = &cobra.Command{
	Use:   "get [user-id]",
	Short: "Get user details",
	Args:  cobra.ExactArgs(1),
	RunE:  runUsersGet,
}

var usersRolesCmd = &cobra.Command{
	Use:   "roles",
	Short: "Manage roles",
}

var usersRolesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List roles",
	RunE:  runUsersRolesList,
}

func init() {
	usersRolesCmd.AddCommand(usersRolesListCmd)
	usersCmd.AddCommand(usersListCmd, usersGetCmd, usersRolesCmd)
}

func runUsersList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewUsersApi(client.V2())
	resp, r, err := api.ListUsers(client.Context())
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to list users: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to list users: %w", err)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func runUsersGet(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	userID := args[0]
	api := datadogV2.NewUsersApi(client.V2())
	resp, r, err := api.GetUser(client.Context(), userID)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to get user: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func runUsersRolesList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewRolesApi(client.V2())
	resp, r, err := api.ListRoles(client.Context())
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to list roles: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to list roles: %w", err)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}
