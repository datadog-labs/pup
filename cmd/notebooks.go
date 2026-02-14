// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

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

var notebooksCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new notebook",
	RunE:  runNotebooksCreate,
}

var notebooksUpdateCmd = &cobra.Command{
	Use:   "update [notebook-id]",
	Short: "Update a notebook",
	Args:  cobra.ExactArgs(1),
	RunE:  runNotebooksUpdate,
}

var notebooksDeleteCmd = &cobra.Command{
	Use:   "delete [notebook-id]",
	Short: "Delete a notebook",
	Args:  cobra.ExactArgs(1),
	RunE:  runNotebooksDelete,
}

func init() {
	notebooksCreateCmd.Flags().String("body", "", "JSON body (@filepath or - for stdin) (required)")
	if err := notebooksCreateCmd.MarkFlagRequired("body"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}

	notebooksUpdateCmd.Flags().String("body", "", "JSON body (@filepath or - for stdin) (required)")
	if err := notebooksUpdateCmd.MarkFlagRequired("body"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}

	notebooksCmd.AddCommand(notebooksListCmd, notebooksGetCmd, notebooksCreateCmd, notebooksUpdateCmd, notebooksDeleteCmd)
}

// readBody reads JSON body content from a file (@path) or stdin (-).
func readBody(value string) ([]byte, error) {
	var data []byte
	var err error

	switch {
	case value == "-":
		data, err = io.ReadAll(inputReader)
		if err != nil {
			return nil, fmt.Errorf("failed to read body from stdin: %w", err)
		}
	case strings.HasPrefix(value, "@"):
		path := strings.TrimPrefix(value, "@")
		data, err = os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read body file: %w", err)
		}
	default:
		return nil, fmt.Errorf("body must be @filepath or - for stdin")
	}

	if !json.Valid(data) {
		return nil, fmt.Errorf("invalid JSON in body")
	}

	return data, nil
}

func runNotebooksCreate(cmd *cobra.Command, args []string) error {
	client, err := getClientForEndpoint("POST", "/api/v1/notebooks")
	if err != nil {
		return err
	}

	bodyFlag, _ := cmd.Flags().GetString("body")
	data, err := readBody(bodyFlag)
	if err != nil {
		return err
	}

	var body datadogV1.NotebookCreateRequest
	if err := json.Unmarshal(data, &body); err != nil {
		return fmt.Errorf("failed to parse notebook create request: %w", err)
	}

	api := datadogV1.NewNotebooksApi(client.V1())
	resp, r, err := api.CreateNotebook(client.Context(), body)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to create notebook: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to create notebook: %w", err)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}
	printOutput("%s\n", output)
	return nil
}

func runNotebooksUpdate(cmd *cobra.Command, args []string) error {
	client, err := getClientForEndpoint("PUT", "/api/v1/notebooks/")
	if err != nil {
		return err
	}

	notebookID := parseInt64(args[0])

	bodyFlag, _ := cmd.Flags().GetString("body")
	data, err := readBody(bodyFlag)
	if err != nil {
		return err
	}

	var body datadogV1.NotebookUpdateRequest
	if err := json.Unmarshal(data, &body); err != nil {
		return fmt.Errorf("failed to parse notebook update request: %w", err)
	}

	api := datadogV1.NewNotebooksApi(client.V1())
	resp, r, err := api.UpdateNotebook(client.Context(), notebookID, body)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to update notebook: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to update notebook: %w", err)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}
	printOutput("%s\n", output)
	return nil
}

func runNotebooksList(cmd *cobra.Command, args []string) error {
	client, err := getClientForEndpoint("GET", "/api/v1/notebooks")
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

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func runNotebooksGet(cmd *cobra.Command, args []string) error {
	client, err := getClientForEndpoint("GET", "/api/v1/notebooks/")
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

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func runNotebooksDelete(cmd *cobra.Command, args []string) error {
	client, err := getClientForEndpoint("DELETE", "/api/v1/notebooks/")
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
