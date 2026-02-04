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

var notebooksCmd = &cobra.Command{
	Use:   "notebooks",
	Short: "Manage notebooks",
	Long: `Manage Datadog notebooks for investigation and documentation.

Notebooks combine graphs, logs, and narrative text to document
investigations, share findings, and create runbooks.

CAPABILITIES:
  • List notebooks
  • Get notebook details
  • Create new notebooks
  • Update notebooks
  • Delete notebooks

EXAMPLES:
  # List all notebooks
  pup notebooks list

  # Get notebook details
  pup notebooks get notebook-id

AUTHENTICATION:
  Requires either OAuth2 authentication or API keys.`,
}

var notebooksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List notebooks",
	RunE:  runNotebooksList,
}

var notebooksGetCmd = &cobra.Command{
	Use:   "get [notebook-id]",
	Short: "Get notebook details",
	Args:  cobra.ExactArgs(1),
	RunE:  runNotebooksGet,
}

var notebooksDeleteCmd = &cobra.Command{
	Use:   "delete [notebook-id]",
	Short: "Delete a notebook",
	Args:  cobra.ExactArgs(1),
	RunE:  runNotebooksDelete,
}

func init() {
	notebooksCmd.AddCommand(notebooksListCmd, notebooksGetCmd, notebooksDeleteCmd)
}

func runNotebooksList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV1.NewNotebooksApi(client.V1())
	resp, r, err := api.ListNotebooks(client.Context())
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to list notebooks: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to list notebooks: %w", err)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func runNotebooksGet(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	notebookID := parseInt64(args[0])
	api := datadogV1.NewNotebooksApi(client.V1())
	resp, r, err := api.GetNotebook(client.Context(), notebookID)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to get notebook: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to get notebook: %w", err)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func runNotebooksDelete(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	notebookID := parseInt64(args[0])
	if !cfg.AutoApprove {
		fmt.Printf("⚠️  WARNING: This will permanently delete notebook %d\n", notebookID)
		fmt.Print("Are you sure you want to continue? (y/N): ")
		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			// User cancelled or error reading input
			fmt.Println("\nOperation cancelled")
			return nil
		}
		if response != "y" && response != "Y" {
			fmt.Println("Operation cancelled")
			return nil
		}
	}

	api := datadogV1.NewNotebooksApi(client.V1())
	r, err := api.DeleteNotebook(client.Context(), notebookID)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to delete notebook: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to delete notebook: %w", err)
	}

	fmt.Printf("Successfully deleted notebook %d\n", notebookID)
	return nil
}
