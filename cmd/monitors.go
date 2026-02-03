// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/DataDog/fetch/pkg/formatter"
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
  fetch monitors list

  # Filter monitors by name
  fetch monitors list --name="CPU"

  # Filter monitors by tags
  fetch monitors list --tags="env:production,team:backend"

  # Get detailed information about a specific monitor
  fetch monitors get 12345678

  # Delete a monitor with confirmation prompt
  fetch monitors delete 12345678

  # Delete a monitor without confirmation (automation)
  fetch monitors delete 12345678 --yes

OUTPUT FORMAT:
  All commands output JSON by default. Use --output flag for other formats.

AUTHENTICATION:
  Requires either OAuth2 authentication (fetch auth login) or API keys
  (DD_API_KEY and DD_APP_KEY environment variables).`,
}

var monitorsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all monitors",
	Long: `List all monitors with optional filtering.

This command retrieves all monitors from your Datadog account. You can filter
the results by monitor name or tags to narrow down the list.

FILTERS:
  --name      Filter by monitor name (substring match)
  --tags      Filter by tags (comma-separated, e.g., "env:prod,team:backend")

EXAMPLES:
  # List all monitors
  fetch monitors list

  # Find monitors with "CPU" in the name
  fetch monitors list --name="CPU"

  # Find production monitors
  fetch monitors list --tags="env:production"

  # Find monitors for a specific team
  fetch monitors list --tags="team:backend"

  # Combine name and tag filters
  fetch monitors list --name="Database" --tags="env:production"

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
	RunE:  runMonitorsList,
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
  fetch monitors get 12345678

  # Get monitor and save to file
  fetch monitors get 12345678 > monitor.json

  # Get monitor with table output
  fetch monitors get 12345678 --output=table

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
	Args:  cobra.ExactArgs(1),
	RunE:  runMonitorsGet,
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
  fetch monitors delete 12345678

  # Delete monitor without confirmation (automation)
  fetch monitors delete 12345678 --yes

  # Delete monitor using global auto-approve
  DD_AUTO_APPROVE=true fetch monitors delete 12345678

CONFIRMATION PROMPT:
  When run without --yes flag, you will see:

    ⚠️  WARNING: This will permanently delete monitor 12345678
    Are you sure you want to continue? (y/N):

  Type 'y' or 'Y' to confirm, or any other key to cancel.

AUTOMATION:
  For scripts and CI/CD pipelines, use one of:
  • --yes flag: fetch monitors delete 12345678 --yes
  • -y flag: fetch monitors delete 12345678 -y
  • Environment: DD_AUTO_APPROVE=true fetch monitors delete 12345678

WARNING:
  Deletion is permanent and cannot be undone. The monitor and all its alert
  history will be removed from Datadog.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runMonitorsDelete,
}

var (
	monitorName string
	monitorTags string
)

func init() {
	monitorsListCmd.Flags().StringVar(&monitorName, "name", "", "Filter monitors by name")
	monitorsListCmd.Flags().StringVar(&monitorTags, "tags", "", "Filter monitors by tags (comma-separated)")

	monitorsCmd.AddCommand(monitorsListCmd)
	monitorsCmd.AddCommand(monitorsGetCmd)
	monitorsCmd.AddCommand(monitorsDeleteCmd)
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
		opts.WithTags(monitorTags)
	}

	resp, r, err := api.ListMonitors(client.Context(), opts)
	if err != nil {
		return fmt.Errorf("failed to list monitors: %w (status: %d)", err, r.StatusCode)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}

	fmt.Println(output)
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
		return fmt.Errorf("failed to get monitor: %w (status: %d)", err, r.StatusCode)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}

	fmt.Println(output)
	return nil
}

func runMonitorsDelete(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	monitorID := args[0]

	// Check if auto-approve is enabled
	if !cfg.AutoApprove {
		fmt.Printf("⚠️  WARNING: This will permanently delete monitor %s\n", monitorID)
		fmt.Print("Are you sure you want to continue? (y/N): ")

		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Operation cancelled")
			return nil
		}
	}

	api := datadogV1.NewMonitorsApi(client.V1())

	resp, r, err := api.DeleteMonitor(client.Context(), parseInt64(monitorID))
	if err != nil {
		return fmt.Errorf("failed to delete monitor: %w (status: %d)", err, r.StatusCode)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}

	fmt.Println(output)
	return nil
}
