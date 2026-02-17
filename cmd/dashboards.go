// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/spf13/cobra"
)

var dashboardsCmd = &cobra.Command{
	Use:   "dashboards",
	Short: "Manage dashboards",
	Long: `Manage Datadog dashboards for data visualization and monitoring.

Dashboards provide customizable views of your metrics, logs, traces, and other
observability data through various widget types including timeseries, heatmaps,
tables, and more.

CAPABILITIES:
  • List all dashboards with metadata
  • Get detailed dashboard configuration including all widgets
  • Delete dashboards (requires confirmation unless --yes flag is used)
  • View dashboard layouts, templates, and template variables

DASHBOARD TYPES:
  • Timeboard: Grid-based layout with synchronized timeseries graphs
  • Screenboard: Flexible free-form layout with any widget placement

WIDGET TYPES:
  • Timeseries: Line, area, or bar graphs over time
  • Query value: Single numeric value with thresholds
  • Table: Tabular data with columns
  • Heatmap: Heat map visualization
  • Toplist: Top N values
  • Change: Value change over time
  • Event timeline: Event stream
  • Free text: Markdown text and images
  • Group: Container for organizing widgets
  • Note: Text annotations
  • Service map: Service dependency visualization
  • And many more...

EXAMPLES:
  # List all dashboards
  pup dashboards list

  # Get detailed dashboard configuration
  pup dashboards get abc-def-123

  # Get dashboard and save to file
  pup dashboards get abc-def-123 > dashboard.json

  # Delete a dashboard with confirmation
  pup dashboards delete abc-def-123

  # Delete a dashboard without confirmation (automation)
  pup dashboards delete abc-def-123 --yes

TEMPLATE VARIABLES:
  Dashboards can include template variables for dynamic filtering:
  • $env: Environment filter
  • $service: Service filter
  • $host: Host filter
  • Custom variables based on tags

AUTHENTICATION:
  Requires either OAuth2 authentication (pup auth login) or API keys
  (DD_API_KEY and DD_APP_KEY environment variables).`,
}

var dashboardsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all dashboards",
	Long: `List all dashboards in your Datadog account.

This command retrieves summary information for all dashboards including their
IDs, titles, descriptions, and metadata.

EXAMPLES:
  # List all dashboards
  pup dashboards list

  # List dashboards with table output
  pup dashboards list --output=table

  # Save dashboard list to file
  pup dashboards list > dashboards.json

OUTPUT FIELDS:
  • id: Dashboard ID (used for get/delete operations)
  • title: Dashboard title/name
  • description: Dashboard description
  • author_handle: Email of dashboard creator
  • created_at: Creation timestamp (ISO 8601)
  • modified_at: Last modification timestamp (ISO 8601)
  • url: Dashboard URL (relative path)
  • is_read_only: Whether dashboard is read-only
  • layout_type: "ordered" (timeboard) or "free" (screenboard)
  • popularity: Popularity score based on views
  • tags: Dashboard tags

FILTERING:
  Currently no filtering is available in the list command. To search:
  • Use jq: pup dashboards list | jq '.dashboards[] | select(.title | contains("API"))'
  • Use grep: pup dashboards list | grep -i "production"

SORTING:
  Dashboards are returned sorted by popularity (most viewed first).`,
	RunE: runDashboardsList,
}

var dashboardsGetCmd = &cobra.Command{
	Use:   "get [dashboard-id]",
	Short: "Get dashboard details",
	Long: `Get complete configuration for a specific dashboard.

This command retrieves the full dashboard definition including all widgets,
layout configuration, template variables, and metadata. The output can be used
to backup, clone, or programmatically modify dashboards.

ARGUMENTS:
  dashboard-id    The dashboard ID (format: xxx-xxx-xxx)

EXAMPLES:
  # Get dashboard configuration
  pup dashboards get abc-def-123

  # Save dashboard to file for backup
  pup dashboards get abc-def-123 > my-dashboard-backup.json

  # Get dashboard with pretty JSON output
  pup dashboards get abc-def-123 | jq .

  # Extract just the widgets
  pup dashboards get abc-def-123 | jq '.widgets'

  # Get dashboard title
  pup dashboards get abc-def-123 | jq -r '.title'

OUTPUT STRUCTURE:
  • id: Dashboard ID
  • title: Dashboard title
  • description: Dashboard description
  • layout_type: "ordered" or "free"
  • widgets: Array of widget configurations
    - definition: Widget definition (queries, visualization)
    - id: Widget ID
    - layout: Widget position and size
  • template_variables: Array of template variable definitions
    - name: Variable name (e.g., "env", "service")
    - prefix: Tag prefix (e.g., "env")
    - default: Default value
    - available_values: List of available values
  • notify_list: List of users/teams to notify on changes
  • reflow_type: Reflow behavior ("auto" or "fixed")
  • created_at: Creation timestamp
  • modified_at: Last modification timestamp
  • author_handle: Dashboard creator

WIDGET DEFINITION FIELDS:
  Each widget contains:
  • type: Widget type (timeseries, query_value, toplist, etc.)
  • requests: Data queries (metrics, logs, traces, etc.)
  • title: Widget title
  • time: Time configuration
  • custom_links: Custom action links
  • markers: Event markers
  • yaxis: Y-axis configuration

USE CASES:
  • Backup dashboards before making changes
  • Clone dashboards to different accounts
  • Version control dashboard definitions
  • Programmatic dashboard generation
  • Extract widget configurations for reuse`,
	Args: cobra.ExactArgs(1),
	RunE: runDashboardsGet,
}

var dashboardsDeleteCmd = &cobra.Command{
	Use:   "delete [dashboard-id]",
	Short: "Delete a dashboard",
	Args:  cobra.ExactArgs(1),
	RunE:  runDashboardsDelete,
}

func init() {
	dashboardsCmd.AddCommand(dashboardsListCmd)
	dashboardsCmd.AddCommand(dashboardsGetCmd)
	dashboardsCmd.AddCommand(dashboardsDeleteCmd)
}

func runDashboardsList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV1.NewDashboardsApi(client.V1())

	resp, r, err := api.ListDashboards(client.Context())
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to list dashboards: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to list dashboards: %w", err)
	}

	return formatAndPrint(resp, nil)
}

func runDashboardsGet(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	dashboardID := args[0]
	api := datadogV1.NewDashboardsApi(client.V1())

	resp, r, err := api.GetDashboard(client.Context(), dashboardID)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to get dashboard: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to get dashboard: %w", err)
	}

	return formatAndPrint(resp, nil)
}

func runDashboardsDelete(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	dashboardID := args[0]

	// Check if auto-approve is enabled
	if !cfg.AutoApprove {
		printOutput("⚠️  WARNING: This will permanently delete dashboard %s\n", dashboardID)
		printOutput("Are you sure you want to continue? (y/N): ")

		response, err := readConfirmation()
		if err != nil {
			// User cancelled or error reading input
			printOutput("\nOperation cancelled\n")
			return nil
		}
		if response != "y" && response != "Y" {
			printOutput("Operation cancelled\n")
			return nil
		}
	}

	api := datadogV1.NewDashboardsApi(client.V1())

	resp, r, err := api.DeleteDashboard(client.Context(), dashboardID)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to delete dashboard: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to delete dashboard: %w", err)
	}

	return formatAndPrint(resp, nil)
}
