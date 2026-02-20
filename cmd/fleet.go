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
	"github.com/spf13/cobra"
)

var fleetCmd = &cobra.Command{
	Use:   "fleet",
	Short: "Manage Fleet Automation",
	Long: `Manage Datadog Fleet Automation for remote agent management at scale.

Fleet Automation enables listing agents, deploying configuration changes,
upgrading agent packages, and scheduling automated operations.

NOTE: Fleet Automation APIs are in Preview and may introduce breaking changes.

CAPABILITIES:
  • List and inspect fleet agents and versions
  • Create and manage configuration and upgrade deployments
  • Schedule, trigger, and manage automated operations

EXAMPLES:
  # List fleet agents
  pup fleet agents list

  # Get agent details
  pup fleet agents get <agent-key>

  # List deployments
  pup fleet deployments list

  # Deploy a configuration change
  pup fleet deployments configure --file=config.json

  # List schedules
  pup fleet schedules list

AUTHENTICATION:
  Requires either OAuth2 authentication or API keys.`,
}

// Agents subcommands
var fleetAgentsCmd = &cobra.Command{
	Use:   "agents",
	Short: "Manage fleet agents",
}

var fleetAgentsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List fleet agents",
	RunE:  runFleetAgentsList,
}

var fleetAgentsGetCmd = &cobra.Command{
	Use:   "get [agent-key]",
	Short: "Get fleet agent details",
	Args:  cobra.ExactArgs(1),
	RunE:  runFleetAgentsGet,
}

var fleetAgentsVersionsCmd = &cobra.Command{
	Use:   "versions",
	Short: "List available agent versions",
	RunE:  runFleetAgentsVersions,
}

// Deployments subcommands
var fleetDeploymentsCmd = &cobra.Command{
	Use:   "deployments",
	Short: "Manage fleet deployments",
}

var fleetDeploymentsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List fleet deployments",
	RunE:  runFleetDeploymentsList,
}

var fleetDeploymentsGetCmd = &cobra.Command{
	Use:   "get [deployment-id]",
	Short: "Get fleet deployment details",
	Args:  cobra.ExactArgs(1),
	RunE:  runFleetDeploymentsGet,
}

var fleetDeploymentsConfigureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Create a configuration deployment",
	RunE:  runFleetDeploymentsConfigure,
}

var fleetDeploymentsUpgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Create an upgrade deployment",
	RunE:  runFleetDeploymentsUpgrade,
}

var fleetDeploymentsCancelCmd = &cobra.Command{
	Use:   "cancel [deployment-id]",
	Short: "Cancel a fleet deployment",
	Args:  cobra.ExactArgs(1),
	RunE:  runFleetDeploymentsCancel,
}

// Schedules subcommands
var fleetSchedulesCmd = &cobra.Command{
	Use:   "schedules",
	Short: "Manage fleet schedules",
}

var fleetSchedulesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List fleet schedules",
	RunE:  runFleetSchedulesList,
}

var fleetSchedulesGetCmd = &cobra.Command{
	Use:   "get [schedule-id]",
	Short: "Get fleet schedule details",
	Args:  cobra.ExactArgs(1),
	RunE:  runFleetSchedulesGet,
}

var fleetSchedulesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a fleet schedule",
	RunE:  runFleetSchedulesCreate,
}

var fleetSchedulesUpdateCmd = &cobra.Command{
	Use:   "update [schedule-id]",
	Short: "Update a fleet schedule",
	Args:  cobra.ExactArgs(1),
	RunE:  runFleetSchedulesUpdate,
}

var fleetSchedulesDeleteCmd = &cobra.Command{
	Use:   "delete [schedule-id]",
	Short: "Delete a fleet schedule",
	Args:  cobra.ExactArgs(1),
	RunE:  runFleetSchedulesDelete,
}

var fleetSchedulesTriggerCmd = &cobra.Command{
	Use:   "trigger [schedule-id]",
	Short: "Trigger a fleet schedule",
	Args:  cobra.ExactArgs(1),
	RunE:  runFleetSchedulesTrigger,
}

var (
	fleetFile     string
	fleetPageSize int64
	fleetTags     string
	fleetFilter   string
	fleetSort     string
	fleetDesc     bool
)

func init() {
	// Agents list flags
	fleetAgentsListCmd.Flags().Int64Var(&fleetPageSize, "page-size", 0, "Number of results per page")
	fleetAgentsListCmd.Flags().StringVar(&fleetTags, "tags", "", "Comma-separated tags to filter by")
	fleetAgentsListCmd.Flags().StringVar(&fleetFilter, "filter", "", "Filter query")
	fleetAgentsListCmd.Flags().StringVar(&fleetSort, "sort", "", "Sort attribute")
	fleetAgentsListCmd.Flags().BoolVar(&fleetDesc, "desc", false, "Sort in descending order")

	// Deployments list flags
	fleetDeploymentsListCmd.Flags().Int64Var(&fleetPageSize, "page-size", 0, "Number of results per page")

	// Deployments file flags
	fleetDeploymentsConfigureCmd.Flags().StringVar(&fleetFile, "file", "", "JSON file with request body (required)")
	_ = fleetDeploymentsConfigureCmd.MarkFlagRequired("file")
	fleetDeploymentsUpgradeCmd.Flags().StringVar(&fleetFile, "file", "", "JSON file with request body (required)")
	_ = fleetDeploymentsUpgradeCmd.MarkFlagRequired("file")

	// Schedules file flags
	fleetSchedulesCreateCmd.Flags().StringVar(&fleetFile, "file", "", "JSON file with request body (required)")
	_ = fleetSchedulesCreateCmd.MarkFlagRequired("file")
	fleetSchedulesUpdateCmd.Flags().StringVar(&fleetFile, "file", "", "JSON file with request body (required)")
	_ = fleetSchedulesUpdateCmd.MarkFlagRequired("file")

	// Build command hierarchy
	fleetAgentsCmd.AddCommand(fleetAgentsListCmd, fleetAgentsGetCmd, fleetAgentsVersionsCmd)
	fleetDeploymentsCmd.AddCommand(
		fleetDeploymentsListCmd,
		fleetDeploymentsGetCmd,
		fleetDeploymentsConfigureCmd,
		fleetDeploymentsUpgradeCmd,
		fleetDeploymentsCancelCmd,
	)
	fleetSchedulesCmd.AddCommand(
		fleetSchedulesListCmd,
		fleetSchedulesGetCmd,
		fleetSchedulesCreateCmd,
		fleetSchedulesUpdateCmd,
		fleetSchedulesDeleteCmd,
		fleetSchedulesTriggerCmd,
	)
	fleetCmd.AddCommand(fleetAgentsCmd, fleetDeploymentsCmd, fleetSchedulesCmd)
}

// Agents implementations

func runFleetAgentsList(cmd *cobra.Command, args []string) error {
	client, err := getClientForEndpoint("GET", "/api/v2/fleet/agents")
	if err != nil {
		return err
	}

	api := datadogV2.NewFleetAutomationApi(client.V2())
	opts := datadogV2.NewListFleetAgentsOptionalParameters()
	if fleetPageSize > 0 {
		opts = opts.WithPageSize(fleetPageSize)
	}
	if fleetTags != "" {
		opts = opts.WithTags(fleetTags)
	}
	if fleetFilter != "" {
		opts = opts.WithFilter(fleetFilter)
	}
	if fleetSort != "" {
		opts = opts.WithSortAttribute(fleetSort)
	}
	if fleetDesc {
		opts = opts.WithSortDescending(true)
	}

	resp, r, err := api.ListFleetAgents(client.Context(), *opts)
	if err != nil {
		return formatAPIError("list fleet agents", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runFleetAgentsGet(cmd *cobra.Command, args []string) error {
	client, err := getClientForEndpoint("GET", "/api/v2/fleet/agents/")
	if err != nil {
		return err
	}

	api := datadogV2.NewFleetAutomationApi(client.V2())
	resp, r, err := api.GetFleetAgentInfo(client.Context(), args[0])
	if err != nil {
		return formatAPIError("get fleet agent info", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runFleetAgentsVersions(cmd *cobra.Command, args []string) error {
	client, err := getClientForEndpoint("GET", "/api/v2/fleet/agents/versions")
	if err != nil {
		return err
	}

	api := datadogV2.NewFleetAutomationApi(client.V2())
	resp, r, err := api.ListFleetAgentVersions(client.Context())
	if err != nil {
		return formatAPIError("list fleet agent versions", err, r)
	}

	return formatAndPrint(resp, nil)
}

// Deployments implementations

func runFleetDeploymentsList(cmd *cobra.Command, args []string) error {
	client, err := getClientForEndpoint("GET", "/api/v2/fleet/deployments")
	if err != nil {
		return err
	}

	api := datadogV2.NewFleetAutomationApi(client.V2())
	opts := datadogV2.NewListFleetDeploymentsOptionalParameters()
	if fleetPageSize > 0 {
		opts = opts.WithPageSize(fleetPageSize)
	}

	resp, r, err := api.ListFleetDeployments(client.Context(), *opts)
	if err != nil {
		return formatAPIError("list fleet deployments", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runFleetDeploymentsGet(cmd *cobra.Command, args []string) error {
	client, err := getClientForEndpoint("GET", "/api/v2/fleet/deployments/")
	if err != nil {
		return err
	}

	api := datadogV2.NewFleetAutomationApi(client.V2())
	resp, r, err := api.GetFleetDeployment(client.Context(), args[0])
	if err != nil {
		return formatAPIError("get fleet deployment", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runFleetDeploymentsConfigure(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile(fleetFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var body datadogV2.FleetDeploymentConfigureCreateRequest
	if err := json.Unmarshal(data, &body); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	client, err := getClientForEndpoint("POST", "/api/v2/fleet/deployments/configure")
	if err != nil {
		return err
	}

	api := datadogV2.NewFleetAutomationApi(client.V2())
	resp, r, err := api.CreateFleetDeploymentConfigure(client.Context(), body)
	if err != nil {
		return formatAPIError("create fleet deployment configure", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runFleetDeploymentsUpgrade(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile(fleetFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var body datadogV2.FleetDeploymentPackageUpgradeCreateRequest
	if err := json.Unmarshal(data, &body); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	client, err := getClientForEndpoint("POST", "/api/v2/fleet/deployments/upgrade")
	if err != nil {
		return err
	}

	api := datadogV2.NewFleetAutomationApi(client.V2())
	resp, r, err := api.CreateFleetDeploymentUpgrade(client.Context(), body)
	if err != nil {
		return formatAPIError("create fleet deployment upgrade", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runFleetDeploymentsCancel(cmd *cobra.Command, args []string) error {
	if !cfg.AutoApprove {
		printOutput("WARNING: This will cancel deployment '%s'.\n", args[0])
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

	client, err := getClientForEndpoint("DELETE", "/api/v2/fleet/deployments/")
	if err != nil {
		return err
	}

	api := datadogV2.NewFleetAutomationApi(client.V2())
	r, err := api.CancelFleetDeployment(client.Context(), args[0])
	if err != nil {
		return formatAPIError("cancel fleet deployment", err, r)
	}

	printOutput("Deployment '%s' cancelled successfully.\n", args[0])
	return nil
}

// Schedules implementations

func runFleetSchedulesList(cmd *cobra.Command, args []string) error {
	client, err := getClientForEndpoint("GET", "/api/v2/fleet/schedules")
	if err != nil {
		return err
	}

	api := datadogV2.NewFleetAutomationApi(client.V2())
	resp, r, err := api.ListFleetSchedules(client.Context())
	if err != nil {
		return formatAPIError("list fleet schedules", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runFleetSchedulesGet(cmd *cobra.Command, args []string) error {
	client, err := getClientForEndpoint("GET", "/api/v2/fleet/schedules/")
	if err != nil {
		return err
	}

	api := datadogV2.NewFleetAutomationApi(client.V2())
	resp, r, err := api.GetFleetSchedule(client.Context(), args[0])
	if err != nil {
		return formatAPIError("get fleet schedule", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runFleetSchedulesCreate(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile(fleetFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var body datadogV2.FleetScheduleCreateRequest
	if err := json.Unmarshal(data, &body); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	client, err := getClientForEndpoint("POST", "/api/v2/fleet/schedules")
	if err != nil {
		return err
	}

	api := datadogV2.NewFleetAutomationApi(client.V2())
	resp, r, err := api.CreateFleetSchedule(client.Context(), body)
	if err != nil {
		return formatAPIError("create fleet schedule", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runFleetSchedulesUpdate(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile(fleetFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var body datadogV2.FleetSchedulePatchRequest
	if err := json.Unmarshal(data, &body); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	client, err := getClientForEndpoint("PATCH", "/api/v2/fleet/schedules/")
	if err != nil {
		return err
	}

	api := datadogV2.NewFleetAutomationApi(client.V2())
	resp, r, err := api.UpdateFleetSchedule(client.Context(), args[0], body)
	if err != nil {
		return formatAPIError("update fleet schedule", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runFleetSchedulesDelete(cmd *cobra.Command, args []string) error {
	if !cfg.AutoApprove {
		printOutput("WARNING: This will permanently delete schedule '%s'.\n", args[0])
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

	client, err := getClientForEndpoint("DELETE", "/api/v2/fleet/schedules/")
	if err != nil {
		return err
	}

	api := datadogV2.NewFleetAutomationApi(client.V2())
	r, err := api.DeleteFleetSchedule(client.Context(), args[0])
	if err != nil {
		return formatAPIError("delete fleet schedule", err, r)
	}

	printOutput("Schedule '%s' deleted successfully.\n", args[0])
	return nil
}

func runFleetSchedulesTrigger(cmd *cobra.Command, args []string) error {
	client, err := getClientForEndpoint("POST", "/api/v2/fleet/schedules/")
	if err != nil {
		return err
	}

	api := datadogV2.NewFleetAutomationApi(client.V2())
	resp, r, err := api.TriggerFleetSchedule(client.Context(), args[0])
	if err != nil {
		return formatAPIError("trigger fleet schedule", err, r)
	}

	return formatAndPrint(resp, nil)
}
