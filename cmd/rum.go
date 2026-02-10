// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"fmt"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/DataDog/pup/pkg/formatter"
	"github.com/spf13/cobra"
)

var rumCmd = &cobra.Command{
	Use:   "rum",
	Short: "Manage Real User Monitoring (RUM)",
	Long: `Manage Datadog Real User Monitoring (RUM) for frontend application performance.

RUM provides visibility into real user experiences across web and mobile applications,
capturing frontend performance metrics, user sessions, errors, and user journeys.

CAPABILITIES:
  • Manage RUM applications (web, mobile, browser)
  • Configure RUM metrics and custom metrics
  • Set up retention filters for session replay and data
  • Query session replay data and playlists
  • Analyze user interaction heatmaps

RUM DATA TYPES:
  • Views: Page views and screen loads
  • Actions: User interactions (clicks, taps, scrolls)
  • Errors: Frontend errors and crashes
  • Resources: Network requests and asset loading
  • Long Tasks: Performance bottlenecks

APPLICATION TYPES:
  • browser: Web applications
  • ios: iOS mobile applications
  • android: Android mobile applications
  • react-native: React Native applications
  • flutter: Flutter applications

EXAMPLES:
  # List all RUM applications
  pup rum apps list

  # Get RUM application details
  pup rum apps get --app-id="abc-123-def"

  # Create a new browser RUM application
  pup rum apps create --name="my-web-app" --type="browser"

  # List RUM custom metrics
  pup rum metrics list

  # List retention filters
  pup rum retention-filters list

  # Query session replay data
  pup rum sessions list --from="1h"

AUTHENTICATION:
  Requires either OAuth2 authentication (pup auth login) or API keys
  (DD_API_KEY and DD_APP_KEY environment variables).`,
}

// RUM Applications Commands
var rumAppsCmd = &cobra.Command{
	Use:   "apps",
	Short: "Manage RUM applications",
	Long: `Manage RUM applications for web and mobile monitoring.

RUM applications represent your frontend applications (web, iOS, Android, etc.)
and provide the context for collecting user experience data.`,
}

var rumAppsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all RUM applications",
	RunE:  runRumAppsList,
}

var rumAppsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get RUM application details",
	RunE:  runRumAppsGet,
}

var rumAppsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new RUM application",
	RunE:  runRumAppsCreate,
}

var rumAppsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a RUM application",
	RunE:  runRumAppsUpdate,
}

var rumAppsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a RUM application",
	RunE:  runRumAppsDelete,
}

// RUM Metrics Commands
var rumMetricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Manage RUM custom metrics",
}

var rumMetricsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all RUM custom metrics",
	RunE:  runRumMetricsList,
}

var rumMetricsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get RUM custom metric details",
	RunE:  runRumMetricsGet,
}

var rumMetricsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a RUM custom metric",
	RunE:  runRumMetricsCreate,
}

var rumMetricsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a RUM custom metric",
	RunE:  runRumMetricsUpdate,
}

var rumMetricsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a RUM custom metric",
	RunE:  runRumMetricsDelete,
}

// RUM Retention Filters Commands
var rumRetentionFiltersCmd = &cobra.Command{
	Use:   "retention-filters",
	Short: "Manage RUM retention filters",
}

var rumRetentionFiltersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all retention filters",
	RunE:  runRumRetentionFiltersList,
}

var rumRetentionFiltersGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get retention filter details",
	RunE:  runRumRetentionFiltersGet,
}

var rumRetentionFiltersCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a retention filter",
	RunE:  runRumRetentionFiltersCreate,
}

var rumRetentionFiltersUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a retention filter",
	RunE:  runRumRetentionFiltersUpdate,
}

var rumRetentionFiltersDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a retention filter",
	RunE:  runRumRetentionFiltersDelete,
}

// RUM Sessions Commands
var rumSessionsCmd = &cobra.Command{
	Use:   "sessions",
	Short: "Query RUM session replay data",
}

var rumSessionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List RUM sessions",
	RunE:  runRumSessionsList,
}

var rumSessionsSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search RUM sessions",
	RunE:  runRumSessionsSearch,
}

// RUM Playlists Commands
var rumPlaylistsCmd = &cobra.Command{
	Use:   "playlists",
	Short: "Manage session replay playlists",
}

var rumPlaylistsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List session replay playlists",
	RunE:  runRumPlaylistsList,
}

var rumPlaylistsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get playlist details",
	RunE:  runRumPlaylistsGet,
}

// RUM Heatmaps Commands
var rumHeatmapsCmd = &cobra.Command{
	Use:   "heatmaps",
	Short: "Query RUM interaction heatmaps",
}

var rumHeatmapsQueryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query heatmap data",
	RunE:  runRumHeatmapsQuery,
}

// Flags for RUM commands
var (
	rumAppID         string
	rumAppName       string
	rumAppType       string
	rumMetricID      string
	rumMetricName    string
	rumEventType     string
	rumFilter        string
	rumCompute       string
	rumGroupBy       string
	rumFilterID      string
	rumFilterName    string
	rumFilterQuery   string
	rumFilterRate    int
	rumFilterType    string
	rumFilterEnabled bool
	rumPlaylistID    string
	rumView          string
	rumQuery         string
	rumFrom          string
	rumTo            string
	rumLimit         int
)

func init() {
	// RUM Apps flags
	rumAppsGetCmd.Flags().StringVar(&rumAppID, "app-id", "", "Application ID (required)")
	if err := rumAppsGetCmd.MarkFlagRequired("app-id"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}

	rumAppsCreateCmd.Flags().StringVar(&rumAppName, "name", "", "Application name (required)")
	rumAppsCreateCmd.Flags().StringVar(&rumAppType, "type", "", "Application type (required)")
	if err := rumAppsCreateCmd.MarkFlagRequired("name"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}
	if err := rumAppsCreateCmd.MarkFlagRequired("type"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}

	rumAppsUpdateCmd.Flags().StringVar(&rumAppID, "app-id", "", "Application ID (required)")
	rumAppsUpdateCmd.Flags().StringVar(&rumAppName, "name", "", "Application name")
	rumAppsUpdateCmd.Flags().StringVar(&rumAppType, "type", "", "Application type")
	if err := rumAppsUpdateCmd.MarkFlagRequired("app-id"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}

	rumAppsDeleteCmd.Flags().StringVar(&rumAppID, "app-id", "", "Application ID (required)")
	if err := rumAppsDeleteCmd.MarkFlagRequired("app-id"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}

	// RUM Metrics flags
	rumMetricsGetCmd.Flags().StringVar(&rumMetricID, "metric-id", "", "Metric ID (required)")
	if err := rumMetricsGetCmd.MarkFlagRequired("metric-id"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}

	rumMetricsCreateCmd.Flags().StringVar(&rumMetricName, "name", "", "Metric name (required)")
	rumMetricsCreateCmd.Flags().StringVar(&rumEventType, "event-type", "", "Event type (required)")
	rumMetricsCreateCmd.Flags().StringVar(&rumFilter, "filter", "", "Query filter")
	rumMetricsCreateCmd.Flags().StringVar(&rumCompute, "compute", "", "Compute JSON (required)")
	rumMetricsCreateCmd.Flags().StringVar(&rumGroupBy, "group-by", "", "Group by JSON")
	if err := rumMetricsCreateCmd.MarkFlagRequired("name"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}
	if err := rumMetricsCreateCmd.MarkFlagRequired("event-type"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}
	if err := rumMetricsCreateCmd.MarkFlagRequired("compute"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}

	rumMetricsUpdateCmd.Flags().StringVar(&rumMetricID, "metric-id", "", "Metric ID (required)")
	rumMetricsUpdateCmd.Flags().StringVar(&rumFilter, "filter", "", "Query filter")
	rumMetricsUpdateCmd.Flags().StringVar(&rumGroupBy, "group-by", "", "Group by JSON")
	rumMetricsUpdateCmd.Flags().StringVar(&rumCompute, "compute", "", "Compute JSON")
	if err := rumMetricsUpdateCmd.MarkFlagRequired("metric-id"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}

	rumMetricsDeleteCmd.Flags().StringVar(&rumMetricID, "metric-id", "", "Metric ID (required)")
	if err := rumMetricsDeleteCmd.MarkFlagRequired("metric-id"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}

	// RUM Retention Filters flags
	rumRetentionFiltersGetCmd.Flags().StringVar(&rumFilterID, "filter-id", "", "Filter ID (required)")
	if err := rumRetentionFiltersGetCmd.MarkFlagRequired("filter-id"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}

	rumRetentionFiltersCreateCmd.Flags().StringVar(&rumFilterName, "name", "", "Filter name (required)")
	rumRetentionFiltersCreateCmd.Flags().StringVar(&rumFilterQuery, "query", "", "Filter query (required)")
	rumRetentionFiltersCreateCmd.Flags().IntVar(&rumFilterRate, "rate", 100, "Sample rate (0-100)")
	rumRetentionFiltersCreateCmd.Flags().StringVar(&rumFilterType, "type", "session-replay", "Filter type")
	rumRetentionFiltersCreateCmd.Flags().BoolVar(&rumFilterEnabled, "enabled", true, "Enable filter")
	if err := rumRetentionFiltersCreateCmd.MarkFlagRequired("name"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}
	if err := rumRetentionFiltersCreateCmd.MarkFlagRequired("query"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}

	rumRetentionFiltersUpdateCmd.Flags().StringVar(&rumFilterID, "filter-id", "", "Filter ID (required)")
	rumRetentionFiltersUpdateCmd.Flags().StringVar(&rumFilterName, "name", "", "Filter name")
	rumRetentionFiltersUpdateCmd.Flags().StringVar(&rumFilterQuery, "query", "", "Filter query")
	rumRetentionFiltersUpdateCmd.Flags().IntVar(&rumFilterRate, "rate", -1, "Sample rate (0-100)")
	rumRetentionFiltersUpdateCmd.Flags().BoolVar(&rumFilterEnabled, "enabled", true, "Enable filter")
	if err := rumRetentionFiltersUpdateCmd.MarkFlagRequired("filter-id"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}

	rumRetentionFiltersDeleteCmd.Flags().StringVar(&rumFilterID, "filter-id", "", "Filter ID (required)")
	if err := rumRetentionFiltersDeleteCmd.MarkFlagRequired("filter-id"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}

	// RUM Sessions flags
	rumSessionsListCmd.Flags().StringVar(&rumFrom, "from", "1h", "Time range start")
	rumSessionsListCmd.Flags().StringVar(&rumTo, "to", "now", "Time range end")
	rumSessionsListCmd.Flags().IntVar(&rumLimit, "limit", 100, "Maximum results")

	rumSessionsSearchCmd.Flags().StringVar(&rumQuery, "query", "", "Search query (required)")
	rumSessionsSearchCmd.Flags().StringVar(&rumFrom, "from", "1h", "Time range start")
	rumSessionsSearchCmd.Flags().StringVar(&rumTo, "to", "now", "Time range end")
	rumSessionsSearchCmd.Flags().IntVar(&rumLimit, "limit", 100, "Maximum results")
	if err := rumSessionsSearchCmd.MarkFlagRequired("query"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}

	// RUM Playlists flags
	rumPlaylistsGetCmd.Flags().StringVar(&rumPlaylistID, "playlist-id", "", "Playlist ID (required)")
	if err := rumPlaylistsGetCmd.MarkFlagRequired("playlist-id"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}

	// RUM Heatmaps flags
	rumHeatmapsQueryCmd.Flags().StringVar(&rumView, "view", "", "View/page name (required)")
	rumHeatmapsQueryCmd.Flags().StringVar(&rumFrom, "from", "24h", "Time range start")
	rumHeatmapsQueryCmd.Flags().StringVar(&rumTo, "to", "now", "Time range end")
	if err := rumHeatmapsQueryCmd.MarkFlagRequired("view"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}

	// Add subcommands
	rumAppsCmd.AddCommand(rumAppsListCmd, rumAppsGetCmd, rumAppsCreateCmd, rumAppsUpdateCmd, rumAppsDeleteCmd)
	rumMetricsCmd.AddCommand(rumMetricsListCmd, rumMetricsGetCmd, rumMetricsCreateCmd, rumMetricsUpdateCmd, rumMetricsDeleteCmd)
	rumRetentionFiltersCmd.AddCommand(rumRetentionFiltersListCmd, rumRetentionFiltersGetCmd, rumRetentionFiltersCreateCmd, rumRetentionFiltersUpdateCmd, rumRetentionFiltersDeleteCmd)
	rumSessionsCmd.AddCommand(rumSessionsListCmd, rumSessionsSearchCmd)
	rumPlaylistsCmd.AddCommand(rumPlaylistsListCmd, rumPlaylistsGetCmd)
	rumHeatmapsCmd.AddCommand(rumHeatmapsQueryCmd)
	rumCmd.AddCommand(rumAppsCmd, rumMetricsCmd, rumRetentionFiltersCmd, rumSessionsCmd, rumPlaylistsCmd, rumHeatmapsCmd)
}

// RUM Apps Implementation
func runRumAppsList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewRUMApi(client.V2())
	resp, r, err := api.GetRUMApplications(client.Context())
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to list RUM applications: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to list RUM applications: %w", err)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func runRumAppsGet(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewRUMApi(client.V2())
	resp, r, err := api.GetRUMApplication(client.Context(), rumAppID)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to get RUM application: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to get RUM application: %w", err)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func runRumAppsCreate(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	validTypes := []string{"browser", "ios", "android", "react-native", "flutter"}
	if !contains(validTypes, rumAppType) {
		return fmt.Errorf("invalid application type: %s (must be one of: %s)", rumAppType, strings.Join(validTypes, ", "))
	}

	api := datadogV2.NewRUMApi(client.V2())
	body := datadogV2.RUMApplicationCreateRequest{
		Data: datadogV2.RUMApplicationCreate{
			Attributes: datadogV2.RUMApplicationCreateAttributes{
				Name: rumAppName,
				Type: &rumAppType,
			},
			Type: datadogV2.RUMAPPLICATIONCREATETYPE_RUM_APPLICATION_CREATE,
		},
	}

	resp, r, err := api.CreateRUMApplication(client.Context(), body)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to create RUM application: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to create RUM application: %w", err)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func runRumAppsUpdate(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	attrs := datadogV2.RUMApplicationUpdateAttributes{}
	if rumAppName != "" {
		attrs.Name = &rumAppName
	}
	if rumAppType != "" {
		validTypes := []string{"browser", "ios", "android", "react-native", "flutter"}
		if !contains(validTypes, rumAppType) {
			return fmt.Errorf("invalid application type: %s", rumAppType)
		}
		attrs.Type = &rumAppType
	}

	api := datadogV2.NewRUMApi(client.V2())
	body := datadogV2.RUMApplicationUpdateRequest{
		Data: datadogV2.RUMApplicationUpdate{
			Attributes: &attrs,
			Id:         rumAppID,
			Type:       datadogV2.RUMAPPLICATIONUPDATETYPE_RUM_APPLICATION_UPDATE,
		},
	}

	resp, r, err := api.UpdateRUMApplication(client.Context(), rumAppID, body)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to update RUM application: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to update RUM application: %w", err)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func runRumAppsDelete(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	if !cfg.AutoApprove {
		fmt.Printf("⚠️  WARNING: This will permanently delete RUM application %s\n", rumAppID)
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

	api := datadogV2.NewRUMApi(client.V2())
	r, err := api.DeleteRUMApplication(client.Context(), rumAppID)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to delete RUM application: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to delete RUM application: %w", err)
	}

	fmt.Printf("Successfully deleted RUM application %s\n", rumAppID)
	return nil
}

// RUM Metrics Implementation
func runRumMetricsList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewRumMetricsApi(client.V2())
	resp, r, err := api.ListRumMetrics(client.Context())
	if err != nil {
		return formatAPIError("list RUM metrics", err, r)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}
	printOutput("%s\n", output)
	return nil
}

func runRumMetricsGet(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	metricID := args[0]
	api := datadogV2.NewRumMetricsApi(client.V2())
	resp, r, err := api.GetRumMetric(client.Context(), metricID)
	if err != nil {
		return formatAPIError("get RUM metric", err, r)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}
	printOutput("%s\n", output)
	return nil
}

func runRumMetricsCreate(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("RUM metrics create is not yet implemented - API client types require additional mapping")
}

func runRumMetricsUpdate(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("RUM metrics update is not yet implemented - API client types require additional mapping")
}

func runRumMetricsDelete(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("RUM metrics delete is not yet implemented - API client types require additional mapping")
}

// RUM Retention Filters Implementation
func runRumRetentionFiltersList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewRumRetentionFiltersApi(client.V2())
	resp, r, err := api.ListRetentionFilters(client.Context(), rumAppID)
	if err != nil {
		return formatAPIError("list RUM retention filters", err, r)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}
	printOutput("%s\n", output)
	return nil
}

func runRumRetentionFiltersGet(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	filterID := args[0]
	api := datadogV2.NewRumRetentionFiltersApi(client.V2())
	resp, r, err := api.GetRetentionFilter(client.Context(), rumAppID, filterID)
	if err != nil {
		return formatAPIError("get RUM retention filter", err, r)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}
	printOutput("%s\n", output)
	return nil
}

func runRumRetentionFiltersCreate(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("RUM retention filters create is not yet implemented - API client types require additional mapping")
}

func runRumRetentionFiltersUpdate(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("RUM retention filters update is not yet implemented - API client types require additional mapping")
}

func runRumRetentionFiltersDelete(cmd *cobra.Command, args []string) error {
	// NOTE: RUM Retention Filters API is not available in datadog-api-client-go v2.30.0
	return fmt.Errorf("RUM retention filters API is not available in the current API client version")
}

// RUM Sessions Implementation
func runRumSessionsList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewRUMApi(client.V2())
	body := datadogV2.RUMSearchEventsRequest{
		Filter: &datadogV2.RUMQueryFilter{
			From: &rumFrom,
			To:   &rumTo,
		},
		Page: &datadogV2.RUMQueryPageOptions{
			Limit: datadog.PtrInt32(int32(rumLimit)),
		},
	}

	resp, r, err := api.SearchRUMEvents(client.Context(), body)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to list RUM sessions: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to list RUM sessions: %w", err)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func runRumSessionsSearch(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewRUMApi(client.V2())
	body := datadogV2.RUMSearchEventsRequest{
		Filter: &datadogV2.RUMQueryFilter{
			Query: &rumQuery,
			From:  &rumFrom,
			To:    &rumTo,
		},
		Page: &datadogV2.RUMQueryPageOptions{
			Limit: datadog.PtrInt32(int32(rumLimit)),
		},
	}

	resp, r, err := api.SearchRUMEvents(client.Context(), body)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to search RUM sessions: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to search RUM sessions: %w", err)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

// RUM Playlists (Placeholder)
func runRumPlaylistsList(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("playlist functionality not yet implemented in Datadog API client")
}

func runRumPlaylistsGet(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("playlist functionality not yet implemented in Datadog API client")
}

// RUM Heatmaps (Placeholder)
func runRumHeatmapsQuery(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("heatmap functionality not yet implemented in Datadog API client")
}

// Helper function
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
