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

var slosCmd = &cobra.Command{
	Use:   "slos",
	Short: "Manage Service Level Objectives",
	Long: `Manage Datadog Service Level Objectives (SLOs) for tracking service reliability.

SLOs help you define and track service reliability targets based on Service Level
Indicators (SLIs). They support various calculation types and target windows.

CAPABILITIES:
  • List all SLOs with status and error budget
  • Get detailed SLO configuration and history
  • Delete SLOs (requires confirmation unless --yes flag is used)
  • View SLO status, error budget burn rate, and target compliance

SLO TYPES:
  • Metric-based: Based on metric queries (e.g., success rate, latency)
  • Monitor-based: Based on monitor uptime
  • Time slice: Based on time slices meeting criteria

TARGET WINDOWS:
  • 7 days (7d)
  • 30 days (30d)
  • 90 days (90d)
  • Custom rolling windows

CALCULATION METHODS:
  • by_count: Count of good events / total events
  • by_uptime: Percentage of time in good state

EXAMPLES:
  # List all SLOs
  fetch slos list

  # Get detailed SLO information
  fetch slos get abc-123-def

  # Get SLO history and status
  fetch slos get abc-123-def | jq '.data'

  # Delete an SLO with confirmation
  fetch slos delete abc-123-def

  # Delete an SLO without confirmation (automation)
  fetch slos delete abc-123-def --yes

ERROR BUDGET:
  Error budget represents the allowed amount of unreliability before breaching
  the SLO target. It's calculated as (1 - target) * time_window.

  Example: 99.9% target over 30 days = 0.1% * 30 days = 43.2 minutes allowed downtime

AUTHENTICATION:
  Requires either OAuth2 authentication (fetch auth login) or API keys
  (DD_API_KEY and DD_APP_KEY environment variables).`,
}

var slosListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all SLOs",
	Long: `List all Service Level Objectives with current status.

This command retrieves all SLOs from your Datadog account including their
current status, error budget, and compliance percentage.

EXAMPLES:
  # List all SLOs
  fetch slos list

  # List SLOs with table output
  fetch slos list --output=table

  # Save SLO list to file
  fetch slos list > slos.json

  # Find SLOs by name with jq
  fetch slos list | jq '.data[] | select(.name | contains("API"))'

  # Check error budget for all SLOs
  fetch slos list | jq '.data[] | {name: .name, error_budget: .error_budget_remaining}'

OUTPUT FIELDS:
  • id: SLO ID
  • name: SLO name
  • description: SLO description
  • type: SLO type (metric, monitor)
  • type_id: Specific type identifier
  • tags: SLO tags
  • thresholds: Array of target thresholds
    - target: Target percentage (e.g., 99.9)
    - target_display: Display string (e.g., "99.9%")
    - timeframe: Target window (7d, 30d, 90d)
    - warning: Optional warning threshold
  • status: Current SLO status
    - state: "breaching", "ok", or "no_data"
    - error_budget_remaining: Percentage of error budget remaining
    - sli_value: Current SLI value
  • created_at: Creation timestamp
  • modified_at: Last modification timestamp
  • creator: User who created the SLO
  • monitor_ids: Associated monitor IDs (for monitor-based SLOs)
  • monitor_tags: Monitor tags used in query (for monitor-based SLOs)

SLO STATES:
  • ok: SLO is meeting target
  • breaching: SLO has breached target (error budget exhausted)
  • no_data: No data available to calculate SLO

FILTERING:
  Use jq to filter results:
  • Breaching SLOs: fetch slos list | jq '.data[] | select(.status.state == "breaching")'
  • High error budget: fetch slos list | jq '.data[] | select(.status.error_budget_remaining > 50)'
  • By tag: fetch slos list | jq '.data[] | select(.tags[] | contains("team:backend"))'`,
	RunE:  runSlosList,
}

var slosGetCmd = &cobra.Command{
	Use:   "get [slo-id]",
	Short: "Get SLO details",
	Long: `Get detailed configuration and status for a specific SLO.

This command retrieves complete information about an SLO including its
configuration, current status, historical performance, and error budget.

ARGUMENTS:
  slo-id    The SLO ID (format: xxx-xxx-xxx)

EXAMPLES:
  # Get SLO details
  fetch slos get abc-123-def

  # Get SLO and save to file
  fetch slos get abc-123-def > slo-backup.json

  # Check error budget remaining
  fetch slos get abc-123-def | jq '.data.error_budget_remaining'

  # Get current SLI value
  fetch slos get abc-123-def | jq '.data.sli_value'

  # View SLO target thresholds
  fetch slos get abc-123-def | jq '.data.thresholds'

OUTPUT STRUCTURE:
  • id: SLO ID
  • name: SLO name
  • description: Detailed description
  • type: SLO type
    - "metric": Based on metric queries
    - "monitor": Based on monitor uptime
    - "time_slice": Based on time slices
  • type_id: Type-specific identifier (0=metric, 1=monitor, 2=time_slice)
  • query: SLO query definition
    - numerator: Good events query (metric-based)
    - denominator: Total events query (metric-based)
  • monitor_ids: Array of monitor IDs (monitor-based)
  • monitor_search: Monitor query (monitor-based)
  • groups: Grouping dimensions
  • tags: SLO tags
  • thresholds: Target definitions
    - target: Target percentage
    - timeframe: Time window
    - warning: Warning threshold
  • created_at: Creation timestamp
  • modified_at: Last modification timestamp
  • creator: Creator information
  • team_tags: Team ownership tags

CURRENT STATUS:
  • state: Current state (ok, breaching, no_data)
  • sli_value: Current SLI percentage
  • error_budget_remaining: Remaining error budget percentage
  • error_budget_burn_rate: Current burn rate

HISTORICAL DATA:
  • history: Array of historical data points
  • uptime: Historical uptime percentages
  • corrections: Manual SLO corrections applied

USE CASES:
  • Monitor SLO compliance and error budget
  • Backup SLO configuration
  • Analyze historical SLO performance
  • Track error budget burn rate
  • Report on service reliability`,
	Args:  cobra.ExactArgs(1),
	RunE:  runSlosGet,
}

var slosDeleteCmd = &cobra.Command{
	Use:   "delete [slo-id]",
	Short: "Delete an SLO",
	Args:  cobra.ExactArgs(1),
	RunE:  runSlosDelete,
}

func init() {
	slosCmd.AddCommand(slosListCmd)
	slosCmd.AddCommand(slosGetCmd)
	slosCmd.AddCommand(slosDeleteCmd)
}

func runSlosList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV1.NewServiceLevelObjectivesApi(client.V1())

	resp, r, err := api.ListSLOs(client.Context())
	if err != nil {
		return fmt.Errorf("failed to list SLOs: %w (status: %d)", err, r.StatusCode)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}

	fmt.Println(output)
	return nil
}

func runSlosGet(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	sloID := args[0]
	api := datadogV1.NewServiceLevelObjectivesApi(client.V1())

	resp, r, err := api.GetSLO(client.Context(), sloID)
	if err != nil {
		return fmt.Errorf("failed to get SLO: %w (status: %d)", err, r.StatusCode)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}

	fmt.Println(output)
	return nil
}

func runSlosDelete(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	sloID := args[0]

	// Check if auto-approve is enabled
	if !cfg.AutoApprove {
		fmt.Printf("⚠️  WARNING: This will permanently delete SLO %s\n", sloID)
		fmt.Print("Are you sure you want to continue? (y/N): ")

		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Operation cancelled")
			return nil
		}
	}

	api := datadogV1.NewServiceLevelObjectivesApi(client.V1())

	resp, r, err := api.DeleteSLO(client.Context(), sloID)
	if err != nil {
		return fmt.Errorf("failed to delete SLO: %w (status: %d)", err, r.StatusCode)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}

	fmt.Println(output)
	return nil
}
