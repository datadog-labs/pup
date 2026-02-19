// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var statusPagesCmd = &cobra.Command{
	Use:   "status-pages",
	Short: "Manage status pages",
	Long: `Manage Datadog Status Pages for communicating service status.

Status Pages provide a public-facing view of your service health, including
components, degradations, and incident updates.

CAPABILITIES:
  • Manage status pages (list, create, update, delete)
  • Manage page components
  • Manage degradation events

EXAMPLES:
  # List status pages
  pup status-pages pages list

  # Create a status page
  pup status-pages pages create --file=page.json

  # List components for a page
  pup status-pages components list <page-id>

AUTHENTICATION:
  Requires either OAuth2 authentication or API keys.`,
}

// Pages subcommands
var statusPagesPagesCmd = &cobra.Command{
	Use:   "pages",
	Short: "Manage status pages",
}

var statusPagesPagesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all status pages",
	RunE:  runStatusPagesList,
}

var statusPagesPagesGetCmd = &cobra.Command{
	Use:   "get [page-id]",
	Short: "Get status page details",
	Args:  cobra.ExactArgs(1),
	RunE:  runStatusPagesGet,
}

var statusPagesPagesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a status page",
	RunE:  runStatusPagesCreate,
}

var statusPagesPagesUpdateCmd = &cobra.Command{
	Use:   "update [page-id]",
	Short: "Update a status page",
	Args:  cobra.ExactArgs(1),
	RunE:  runStatusPagesUpdate,
}

var statusPagesPagesDeleteCmd = &cobra.Command{
	Use:   "delete [page-id]",
	Short: "Delete a status page",
	Args:  cobra.ExactArgs(1),
	RunE:  runStatusPagesDelete,
}

// Components subcommands
var statusPagesComponentsCmd = &cobra.Command{
	Use:   "components",
	Short: "Manage status page components",
}

var statusPagesComponentsListCmd = &cobra.Command{
	Use:   "list [page-id]",
	Short: "List components for a page",
	Args:  cobra.ExactArgs(1),
	RunE:  runStatusPagesComponentsList,
}

var statusPagesComponentsGetCmd = &cobra.Command{
	Use:   "get [page-id] [component-id]",
	Short: "Get component details",
	Args:  cobra.ExactArgs(2),
	RunE:  runStatusPagesComponentsGet,
}

var statusPagesComponentsCreateCmd = &cobra.Command{
	Use:   "create [page-id]",
	Short: "Create a component",
	Args:  cobra.ExactArgs(1),
	RunE:  runStatusPagesComponentsCreate,
}

var statusPagesComponentsUpdateCmd = &cobra.Command{
	Use:   "update [page-id] [component-id]",
	Short: "Update a component",
	Args:  cobra.ExactArgs(2),
	RunE:  runStatusPagesComponentsUpdate,
}

var statusPagesComponentsDeleteCmd = &cobra.Command{
	Use:   "delete [page-id] [component-id]",
	Short: "Delete a component",
	Args:  cobra.ExactArgs(2),
	RunE:  runStatusPagesComponentsDelete,
}

// Degradations subcommands
var statusPagesDegradationsCmd = &cobra.Command{
	Use:   "degradations",
	Short: "Manage status page degradations",
}

var statusPagesDegradationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List degradations",
	RunE:  runStatusPagesDegradationsList,
}

var statusPagesDegradationsGetCmd = &cobra.Command{
	Use:   "get [page-id] [degradation-id]",
	Short: "Get degradation details",
	Args:  cobra.ExactArgs(2),
	RunE:  runStatusPagesDegradationsGet,
}

var statusPagesDegradationsCreateCmd = &cobra.Command{
	Use:   "create [page-id]",
	Short: "Create a degradation",
	Args:  cobra.ExactArgs(1),
	RunE:  runStatusPagesDegradationsCreate,
}

var statusPagesDegradationsUpdateCmd = &cobra.Command{
	Use:   "update [page-id] [degradation-id]",
	Short: "Update a degradation",
	Args:  cobra.ExactArgs(2),
	RunE:  runStatusPagesDegradationsUpdate,
}

var statusPagesDegradationsDeleteCmd = &cobra.Command{
	Use:   "delete [page-id] [degradation-id]",
	Short: "Delete a degradation",
	Args:  cobra.ExactArgs(2),
	RunE:  runStatusPagesDegradationsDelete,
}

var (
	statusPagesFile string
)

func init() {
	// File flags for create/update operations
	statusPagesPagesCreateCmd.Flags().StringVar(&statusPagesFile, "file", "", "JSON file with request body (required)")
	_ = statusPagesPagesCreateCmd.MarkFlagRequired("file")
	statusPagesPagesUpdateCmd.Flags().StringVar(&statusPagesFile, "file", "", "JSON file with request body (required)")
	_ = statusPagesPagesUpdateCmd.MarkFlagRequired("file")
	statusPagesComponentsCreateCmd.Flags().StringVar(&statusPagesFile, "file", "", "JSON file with request body (required)")
	_ = statusPagesComponentsCreateCmd.MarkFlagRequired("file")
	statusPagesComponentsUpdateCmd.Flags().StringVar(&statusPagesFile, "file", "", "JSON file with request body (required)")
	_ = statusPagesComponentsUpdateCmd.MarkFlagRequired("file")
	statusPagesDegradationsCreateCmd.Flags().StringVar(&statusPagesFile, "file", "", "JSON file with request body (required)")
	_ = statusPagesDegradationsCreateCmd.MarkFlagRequired("file")
	statusPagesDegradationsUpdateCmd.Flags().StringVar(&statusPagesFile, "file", "", "JSON file with request body (required)")
	_ = statusPagesDegradationsUpdateCmd.MarkFlagRequired("file")

	// Build command hierarchy
	statusPagesPagesCmd.AddCommand(
		statusPagesPagesListCmd,
		statusPagesPagesGetCmd,
		statusPagesPagesCreateCmd,
		statusPagesPagesUpdateCmd,
		statusPagesPagesDeleteCmd,
	)
	statusPagesComponentsCmd.AddCommand(
		statusPagesComponentsListCmd,
		statusPagesComponentsGetCmd,
		statusPagesComponentsCreateCmd,
		statusPagesComponentsUpdateCmd,
		statusPagesComponentsDeleteCmd,
	)
	statusPagesDegradationsCmd.AddCommand(
		statusPagesDegradationsListCmd,
		statusPagesDegradationsGetCmd,
		statusPagesDegradationsCreateCmd,
		statusPagesDegradationsUpdateCmd,
		statusPagesDegradationsDeleteCmd,
	)
	statusPagesCmd.AddCommand(statusPagesPagesCmd, statusPagesComponentsCmd, statusPagesDegradationsCmd)
}

// parseUUID parses a string into a uuid.UUID
func parseUUID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("invalid UUID '%s': %w", s, err)
	}
	return id, nil
}

// Pages implementations
func runStatusPagesList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewStatusPagesApi(client.V2())
	resp, r, err := api.ListStatusPages(client.Context())
	if err != nil {
		return formatAPIError("list status pages", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runStatusPagesGet(cmd *cobra.Command, args []string) error {
	pageID, err := parseUUID(args[0])
	if err != nil {
		return err
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewStatusPagesApi(client.V2())
	resp, r, err := api.GetStatusPage(client.Context(), pageID)
	if err != nil {
		return formatAPIError("get status page", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runStatusPagesCreate(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile(statusPagesFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var body datadogV2.CreateStatusPageRequest
	if err := json.Unmarshal(data, &body); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewStatusPagesApi(client.V2())
	resp, r, err := api.CreateStatusPage(client.Context(), body)
	if err != nil {
		return formatAPIError("create status page", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runStatusPagesUpdate(cmd *cobra.Command, args []string) error {
	pageID, err := parseUUID(args[0])
	if err != nil {
		return err
	}

	data, err := os.ReadFile(statusPagesFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var body datadogV2.PatchStatusPageRequest
	if err := json.Unmarshal(data, &body); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewStatusPagesApi(client.V2())
	resp, r, err := api.UpdateStatusPage(client.Context(), pageID, body)
	if err != nil {
		return formatAPIError("update status page", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runStatusPagesDelete(cmd *cobra.Command, args []string) error {
	pageID, err := parseUUID(args[0])
	if err != nil {
		return err
	}

	if !cfg.AutoApprove {
		printOutput("WARNING: This will permanently delete status page '%s'.\n", args[0])
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

	api := datadogV2.NewStatusPagesApi(client.V2())
	r, err := api.DeleteStatusPage(client.Context(), pageID)
	if err != nil {
		return formatAPIError("delete status page", err, r)
	}

	printOutput("Status page '%s' deleted successfully.\n", args[0])
	return nil
}

// Components implementations
func runStatusPagesComponentsList(cmd *cobra.Command, args []string) error {
	pageID, err := parseUUID(args[0])
	if err != nil {
		return err
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewStatusPagesApi(client.V2())
	resp, r, err := api.ListComponents(client.Context(), pageID)
	if err != nil {
		return formatAPIError("list components", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runStatusPagesComponentsGet(cmd *cobra.Command, args []string) error {
	pageID, err := parseUUID(args[0])
	if err != nil {
		return err
	}
	componentID, err := parseUUID(args[1])
	if err != nil {
		return err
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewStatusPagesApi(client.V2())
	resp, r, err := api.GetComponent(client.Context(), pageID, componentID)
	if err != nil {
		return formatAPIError("get component", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runStatusPagesComponentsCreate(cmd *cobra.Command, args []string) error {
	pageID, err := parseUUID(args[0])
	if err != nil {
		return err
	}

	data, err := os.ReadFile(statusPagesFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var body datadogV2.CreateComponentRequest
	if err := json.Unmarshal(data, &body); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewStatusPagesApi(client.V2())
	resp, r, err := api.CreateComponent(client.Context(), pageID, body)
	if err != nil {
		return formatAPIError("create component", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runStatusPagesComponentsUpdate(cmd *cobra.Command, args []string) error {
	pageID, err := parseUUID(args[0])
	if err != nil {
		return err
	}
	componentID, err := parseUUID(args[1])
	if err != nil {
		return err
	}

	data, err := os.ReadFile(statusPagesFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var body datadogV2.PatchComponentRequest
	if err := json.Unmarshal(data, &body); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewStatusPagesApi(client.V2())
	resp, r, err := api.UpdateComponent(client.Context(), pageID, componentID, body)
	if err != nil {
		return formatAPIError("update component", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runStatusPagesComponentsDelete(cmd *cobra.Command, args []string) error {
	pageID, err := parseUUID(args[0])
	if err != nil {
		return err
	}
	componentID, err := parseUUID(args[1])
	if err != nil {
		return err
	}

	if !cfg.AutoApprove {
		printOutput("WARNING: This will permanently delete component '%s'.\n", args[1])
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

	api := datadogV2.NewStatusPagesApi(client.V2())
	r, err := api.DeleteComponent(client.Context(), pageID, componentID)
	if err != nil {
		return formatAPIError("delete component", err, r)
	}

	printOutput("Component '%s' deleted successfully.\n", args[1])
	return nil
}

// Degradations implementations
func runStatusPagesDegradationsList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewStatusPagesApi(client.V2())
	resp, r, err := api.ListDegradations(client.Context())
	if err != nil {
		return formatAPIError("list degradations", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runStatusPagesDegradationsGet(cmd *cobra.Command, args []string) error {
	pageID, err := parseUUID(args[0])
	if err != nil {
		return err
	}
	degradationID, err := parseUUID(args[1])
	if err != nil {
		return err
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewStatusPagesApi(client.V2())
	resp, r, err := api.GetDegradation(client.Context(), pageID, degradationID)
	if err != nil {
		return formatAPIError("get degradation", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runStatusPagesDegradationsCreate(cmd *cobra.Command, args []string) error {
	pageID, err := parseUUID(args[0])
	if err != nil {
		return err
	}

	data, err := os.ReadFile(statusPagesFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var body datadogV2.CreateDegradationRequest
	if err := json.Unmarshal(data, &body); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewStatusPagesApi(client.V2())
	resp, r, err := api.CreateDegradation(client.Context(), pageID, body)
	if err != nil {
		return formatAPIError("create degradation", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runStatusPagesDegradationsUpdate(cmd *cobra.Command, args []string) error {
	pageID, err := parseUUID(args[0])
	if err != nil {
		return err
	}
	degradationID, err := parseUUID(args[1])
	if err != nil {
		return err
	}

	data, err := os.ReadFile(statusPagesFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var body datadogV2.PatchDegradationRequest
	if err := json.Unmarshal(data, &body); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewStatusPagesApi(client.V2())
	resp, r, err := api.UpdateDegradation(client.Context(), pageID, degradationID, body)
	if err != nil {
		return formatAPIError("update degradation", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runStatusPagesDegradationsDelete(cmd *cobra.Command, args []string) error {
	pageID, err := parseUUID(args[0])
	if err != nil {
		return err
	}
	degradationID, err := parseUUID(args[1])
	if err != nil {
		return err
	}

	if !cfg.AutoApprove {
		printOutput("WARNING: This will permanently delete degradation '%s'.\n", args[1])
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

	api := datadogV2.NewStatusPagesApi(client.V2())
	r, err := api.DeleteDegradation(client.Context(), pageID, degradationID)
	if err != nil {
		return formatAPIError("delete degradation", err, r)
	}

	printOutput("Degradation '%s' deleted successfully.\n", args[1])
	return nil
}
