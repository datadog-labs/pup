// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

//go:build js

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication commands (limited in WASM)",
	Long:  `In WASM builds, OAuth2 browser login is not available. Set DD_ACCESS_TOKEN or DD_API_KEY + DD_APP_KEY instead.`,
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login via OAuth2 (not available in WASM)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("OAuth2 browser login is not available in WASM builds; set DD_ACCESS_TOKEN instead")
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check authentication status",
	RunE: func(cmd *cobra.Command, args []string) error {
		hasAccessToken := cfg.AccessToken != ""
		hasAPIKeys := cfg.APIKey != "" && cfg.AppKey != ""

		if hasAccessToken {
			fmt.Println("Authenticated via DD_ACCESS_TOKEN")
		} else if hasAPIKeys {
			fmt.Println("Authenticated via API keys (DD_API_KEY + DD_APP_KEY)")
		} else {
			fmt.Println("Not authenticated")
			fmt.Println("  Set DD_ACCESS_TOKEN or DD_API_KEY + DD_APP_KEY environment variables")
		}
		fmt.Printf("  Site: %s\n", cfg.Site)
		return nil
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout (not available in WASM)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("token storage is not available in WASM builds; unset DD_ACCESS_TOKEN to revoke access")
	},
}

var authRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh access token (not available in WASM)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("token refresh is not available in WASM builds; set a new DD_ACCESS_TOKEN instead")
	},
}

func init() {
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authLogoutCmd)
	authCmd.AddCommand(authRefreshCmd)
}
