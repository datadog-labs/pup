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

var monitorsCmd = &cobra.Command{
	Use:   "monitors",
	Short: "Manage monitors",
	Long: `Manage Datadog monitors for alerting and notifications.

Monitors watch your metrics, logs, traces, and other data sources to alert you when
conditions are met. They support various monitor types including metric, log, trace,
composite, and more.

CAPABILITIES:
  • List all monitors with optional filtering by name or tags
  • Get detailed information about a specific monitor
  • Delete monitors (requires confirmation unless --yes flag is used)
  • View monitor configuration, thresholds, and notification settings

MONITOR TYPES:
  • metric alert: Alert on metric threshold
  • log alert: Alert on log query matches
  • trace-analytics alert: Alert on APM trace patterns
  • composite: Combine multiple monitors with boolean logic
  • service check: Alert on service check status
  • event alert: Alert on event patterns
  • process alert: Alert on process status

EXAMPLES:
  # List all monitors
  pup monitors list

  # Filter monitors by name
  pup monitors list --name="CPU"

  # Filter monitors by tags
  pup monitors list --tags="env:production,team:backend"

  # Get detailed information about a specific monitor
  pup monitors get 12345678

  # Delete a monitor with confirmation prompt
  pup monitors delete 12345678

  # Delete a monitor without confirmation (automation)
  pup monitors delete 12345678 --yes

OUTPUT FORMAT:
  All commands output JSON by default. Use --output flag for other formats.

AUTHENTICATION:
  Requires either OAuth2 authentication (pup auth login) or API keys
  (DD_API_KEY and DD_APP_KEY environment variables).`,
}

var monitorsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List monitors (limited results)",
	Long: `List monitors with optional filtering (returns up to limit).

This command retrieves monitors from your Datadog account. By default, it returns
up to 200 monitors. To see more monitors, use filters to narrow down results
or increase --limit (max 1000).

IMPORTANT: This command returns a LIMITED number of results (default 200, max 1000).
It does not return all monitors. Use filters to find specific monitors.

FILTERS:
  --name   Filter by monitor name (substring match)
  --tags   Filter by tags (comma-separated, e.g., "env:prod,team:backend")
  --limit  Maximum number of monitors to return (default: 200, max: 1000)

EXAMPLES:
  # List up to 200 monitors (default)
  pup monitors list

  # Find monitors with "CPU" in the name
  pup monitors list --name="CPU"

  # Find production monitors
  pup monitors list --tags="env:production"

  # Find monitors for a specific team
  pup monitors list --tags="team:backend"

  # Combine name and tag filters
  pup monitors list --name="Database" --tags="env:production"

  # Get up to 1000 monitors (maximum allowed)
  pup monitors list --limit=1000

  # Get only 50 monitors
  pup monitors list --limit=50

WORKING WITH LARGE SETS:
  This command returns a limited number of results. To work with large numbers of
  monitors, use filters (--name, --tags) to narrow down the results to find
  specific monitors rather than trying to retrieve all monitors.

OUTPUT FIELDS:
  • id: Monitor ID
  • name: Monitor name
  • type: Monitor type (metric, log, composite, etc.)
  • query: Monitor query
  • message: Alert message
  • tags: Monitor tags
  • options: Monitor configuration options
  • overall_state: Current state (Alert, Warn, No Data, OK)
  • created: Creation timestamp
  • modified: Last modification timestamp`,
	RunE: runMonitorsList,
}

var monitorsGetCmd = &cobra.Command{
	Use:   "get [monitor-id]",
	Short: "Get monitor details",
	Long: `Get detailed information about a specific monitor.

This command retrieves all configuration details for a monitor including
thresholds, notification settings, evaluation windows, and metadata.

ARGUMENTS:
  monitor-id    The numeric ID of the monitor

EXAMPLES:
  # Get monitor details
  pup monitors get 12345678

  # Get monitor and save to file
  pup monitors get 12345678 > monitor.json

  # Get monitor with table output
  pup monitors get 12345678 --output=table

OUTPUT INCLUDES:
  • id: Monitor ID
  • name: Monitor name
  • type: Monitor type
  • query: Monitor query/formula
  • message: Alert notification message with @mentions
  • tags: List of tags
  • options: Configuration options
    - thresholds: Alert and warning thresholds
    - notify_no_data: Whether to alert on no data
    - no_data_timeframe: Minutes before no data alert
    - renotify_interval: Minutes between re-notifications
    - timeout_h: Hours before auto-resolve
    - include_tags: Whether to include tags in notifications
    - require_full_window: Require full evaluation window
    - new_group_delay: Seconds to wait for new group
  • overall_state: Current state
  • overall_state_modified: When state last changed
  • created: Creation timestamp
  • creator: User who created the monitor
  • modified: Last modification timestamp`,
	Args: cobra.ExactArgs(1),
	RunE: runMonitorsGet,
}

var monitorsDeleteCmd = &cobra.Command{
	Use:   "delete [monitor-id]",
	Short: "Delete a monitor",
	Long: `Delete a monitor permanently.

This is a DESTRUCTIVE operation that permanently removes the monitor and all its
alert history. By default, this command will prompt for confirmation. Use the
--yes flag to skip confirmation (useful for automation).

ARGUMENTS:
  monitor-id    The numeric ID of the monitor to delete

FLAGS:
  --yes, -y     Skip confirmation prompt (auto-approve)

EXAMPLES:
  # Delete monitor with confirmation prompt
  pup monitors delete 12345678

  # Delete monitor without confirmation (automation)
  pup monitors delete 12345678 --yes

  # Delete monitor using global auto-approve
  DD_AUTO_APPROVE=true pup monitors delete 12345678

CONFIRMATION PROMPT:
  When run without --yes flag, you will see:

    ⚠️  WARNING: This will permanently delete monitor 12345678
    Are you sure you want to continue? (y/N):

  Type 'y' or 'Y' to confirm, or any other key to cancel.

AUTOMATION:
  For scripts and CI/CD pipelines, use one of:
  • --yes flag: pup monitors delete 12345678 --yes
  • -y flag: pup monitors delete 12345678 -y
  • Environment: DD_AUTO_APPROVE=true pup monitors delete 12345678

WARNING:
  Deletion is permanent and cannot be undone. The monitor and all its alert
  history will be removed from Datadog.`,
	Args: cobra.ExactArgs(1),
	RunE: runMonitorsDelete,
}

var monitorsSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search monitors",
	Long: `Search monitors using a query string.

Search allows more flexible querying than list filtering, supporting
advanced search syntax for finding specific monitors.

FLAGS:
  --query        Search query string
  --page         Page number (default: 0)
  --per-page     Results per page (default: 30)
  --sort         Sort order (e.g., "name,asc", "id,desc")

EXAMPLES:
  # Search for monitors by text
  pup monitors search --query="database"

  # Search with pagination
  pup monitors search --query="cpu" --page=1 --per-page=50

  # Search and sort
  pup monitors search --query="memory" --sort="name,asc"`,
	RunE: runMonitorsSearch,
}

var (
	monitorName   string
	monitorTags   string
	monitorLimit  int32
	searchQuery   string
	searchPage    int64
	searchPerPage int64
	searchSort    string
)

func init() {
	monitorsListCmd.Flags().StringVar(&monitorName, "name", "", "Filter monitors by name")
	monitorsListCmd.Flags().StringVar(&monitorTags, "tags", "", "Filter by monitor tags (comma-separated, e.g., team:backend,env:prod)")
	monitorsListCmd.Flags().Int32Var(&monitorLimit, "limit", 200, "Maximum number of monitors to return (default: 200, max: 1000)")

	monitorsSearchCmd.Flags().StringVar(&searchQuery, "query", "", "Search query string")
	monitorsSearchCmd.Flags().Int64Var(&searchPage, "page", 0, "Page number")
	monitorsSearchCmd.Flags().Int64Var(&searchPerPage, "per-page", 30, "Results per page")
	monitorsSearchCmd.Flags().StringVar(&searchSort, "sort", "", "Sort order")

	monitorsCmd.AddCommand(
		monitorsListCmd,
		monitorsGetCmd,
		monitorsDeleteCmd,
		monitorsSearchCmd,
	)
}

func runMonitorsList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV1.NewMonitorsApi(client.V1())

	opts := datadogV1.ListMonitorsOptionalParameters{}
	if monitorName != "" {
		opts.WithName(monitorName)
	}
	if monitorTags != "" {
		opts.WithMonitorTags(monitorTags)
	}

	// In agent mode, use a larger default limit (500) unless explicitly set
	if isAgentMode() && !cmd.Flags().Changed("limit") {
		monitorLimit = 500
	}

	if monitorLimit > 1000 {
		monitorLimit = 1000
	}
	if monitorLimit < 1 {
		monitorLimit = 200
	}

	// Use limit as page size and request first page only
	opts.WithPageSize(monitorLimit)
	opts.WithPage(0)

	resp, r, err := api.ListMonitors(client.Context(), opts)
	if err != nil {
		return formatAPIError("list monitors", err, r)
	}

	// Show count of monitors found (helpful for debugging filters)
	if len(resp) == 0 {
		printOutput("No monitors found matching the specified criteria.\n")
		if monitorName != "" || monitorTags != "" {
			printOutput("Try adjusting your filters (--name or --tags) or removing them to see all monitors.\n")
		}
		return nil
	}

	// Enforce limit - only return up to requested number of items
	// API might return more items, so we truncate to the requested limit
	originalCount := len(resp)
	if len(resp) > int(monitorLimit) {
		resp = resp[:monitorLimit]
	}

	count := len(resp)
	truncated := originalCount > int(monitorLimit)
	var meta *formatter.Metadata
	if isAgentMode() {
		meta = &formatter.Metadata{
			Count:     &count,
			Truncated: truncated,
			Command:   "monitors list",
		}
		if truncated {
			meta.NextAction = fmt.Sprintf("Use --limit=%d or refine with --tags/--name filters", min(int(monitorLimit)*2, 1000))
		}
	}

	if err := formatAndPrint(resp, meta); err != nil {
		return err
	}

	// Show count info if we're truncating (human mode only)
	if !isAgentMode() && truncated {
		printOutput("\nShowing %d of %d monitors (use --limit to adjust)\n", monitorLimit, originalCount)
	}

	return nil
}

func runMonitorsGet(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	monitorID := args[0]
	api := datadogV1.NewMonitorsApi(client.V1())

	resp, r, err := api.GetMonitor(client.Context(), parseInt64(monitorID))
	if err != nil {
		return formatAPIError("get monitor", err, r)
	}

	var meta *formatter.Metadata
	if isAgentMode() {
		meta = &formatter.Metadata{Command: "monitors get"}
	}
	return formatAndPrint(resp, meta)
}

func runMonitorsDelete(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	monitorID := args[0]

	// Check if auto-approve is enabled
	if !cfg.AutoApprove {
		printOutput("⚠️  WARNING: This will permanently delete monitor %s\n", monitorID)
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

	api := datadogV1.NewMonitorsApi(client.V1())

	resp, r, err := api.DeleteMonitor(client.Context(), parseInt64(monitorID))
	if err != nil {
		return formatAPIError("delete monitor", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runMonitorsSearch(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV1.NewMonitorsApi(client.V1())
	opts := datadogV1.SearchMonitorsOptionalParameters{}

	if searchQuery != "" {
		opts.WithQuery(searchQuery)
	}
	if searchPage > 0 {
		opts.WithPage(searchPage)
	}
	if searchPerPage > 0 {
		opts.WithPerPage(searchPerPage)
	}
	if searchSort != "" {
		opts.WithSort(searchSort)
	}

	resp, r, err := api.SearchMonitors(client.Context(), opts)
	if err != nil {
		return formatAPIError("search monitors", err, r)
	}

	return formatAndPrint(resp, nil)
}
