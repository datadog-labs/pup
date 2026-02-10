// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"fmt"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/DataDog/pup/pkg/formatter"
	"github.com/spf13/cobra"
)

var costCmd = &cobra.Command{
	Use:   "cost",
	Short: "Manage cost and billing data",
	Long: `Query cost management and billing information.

Access projected costs, cost attribution by tags, and organizational cost breakdowns.
Cost data is typically available with 12-24 hour delay.

CAPABILITIES:
  • View projected end-of-month costs
  • Get cost attribution by tags and teams
  • Query historical and estimated costs by organization

EXAMPLES:
  # Get projected costs for current month
  pup cost projected

  # Get cost attribution by team tag
  pup cost attribution --start-month=2024-01 --fields=team

  # Get actual costs for a specific month
  pup cost by-org --start-month=2024-01

AUTHENTICATION:
  Requires OAuth2 (via 'pup auth login') or valid API + Application keys.
  Cost management features require billing:read permissions.`,
}

var costProjectedCmd = &cobra.Command{
	Use:   "projected",
	Short: "Get projected end-of-month costs",
	Long: `Get projected costs for the current month based on usage trends.

Provides cost projections by product family to help estimate end-of-month bills.`,
	RunE: runCostProjected,
}

var costAttributionCmd = &cobra.Command{
	Use:   "attribution",
	Short: "Get cost attribution by tags",
	Long: `Get monthly cost attribution broken down by tag keys.

Shows how costs are distributed across different tag values (e.g., teams, services, environments).

REQUIRED FLAGS:
  --start-month    Start month in YYYY-MM format
  --fields         Comma-separated tag keys for breakdown (e.g., "team,env,service")

OPTIONAL FLAGS:
  --end-month      End month (defaults to start-month)

EXAMPLES:
  # Get cost by team for January 2024
  pup cost attribution --start-month=2024-01 --fields=team

  # Get cost by multiple dimensions
  pup cost attribution --start-month=2024-01 --fields=team,env,service

  # Get cost range
  pup cost attribution --start-month=2024-01 --end-month=2024-03 --fields=team`,
	RunE: runCostAttribution,
}

var costByOrgCmd = &cobra.Command{
	Use:   "by-org",
	Short: "Get costs by organization",
	Long: `Get cost breakdown by organization for a specific time period.

Provides actual, estimated, or historical cost data by organization.

REQUIRED FLAGS:
  --start-month    Start month in YYYY-MM format

OPTIONAL FLAGS:
  --end-month      End month (defaults to start-month)
  --view           View type: actual, estimated, historical (default: actual)

EXAMPLES:
  # Get actual costs for January 2024
  pup cost by-org --start-month=2024-01

  # Get estimated costs
  pup cost by-org --start-month=2024-01 --view=estimated

  # Get historical costs
  pup cost by-org --start-month=2024-01 --view=historical`,
	RunE: runCostByOrg,
}

var (
	costStartMonth string
	costEndMonth   string
	costFields     string
	costView       string
)

func init() {
	// Attribution flags
	costAttributionCmd.Flags().StringVar(&costStartMonth, "start-month", "", "Start month (YYYY-MM) (required)")
	costAttributionCmd.Flags().StringVar(&costEndMonth, "end-month", "", "End month (YYYY-MM)")
	costAttributionCmd.Flags().StringVar(&costFields, "fields", "", "Tag keys for breakdown (required)")
	_ = costAttributionCmd.MarkFlagRequired("start-month")
	_ = costAttributionCmd.MarkFlagRequired("fields")

	// By-org flags
	costByOrgCmd.Flags().StringVar(&costStartMonth, "start-month", "", "Start month (YYYY-MM) (required)")
	costByOrgCmd.Flags().StringVar(&costEndMonth, "end-month", "", "End month (YYYY-MM)")
	costByOrgCmd.Flags().StringVar(&costView, "view", "actual", "View type: actual, estimated, historical")
	_ = costByOrgCmd.MarkFlagRequired("start-month")

	// Command hierarchy
	costCmd.AddCommand(costProjectedCmd, costAttributionCmd, costByOrgCmd)
}

func runCostProjected(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewUsageMeteringApi(client.V2())
	opts := datadogV2.GetProjectedCostOptionalParameters{}

	resp, r, err := api.GetProjectedCost(client.Context(), opts)
	if err != nil {
		return formatAPIError("get projected cost", err, r)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}

	printOutput("%s\n", output)
	return nil
}

func runCostAttribution(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	// Parse start month
	startTime, err := time.Parse("2006-01", costStartMonth)
	if err != nil {
		return fmt.Errorf("invalid start month format (use YYYY-MM): %w", err)
	}

	// Parse end month if provided, otherwise use start month
	endTime := startTime
	if costEndMonth != "" {
		endTime, err = time.Parse("2006-01", costEndMonth)
		if err != nil {
			return fmt.Errorf("invalid end month format (use YYYY-MM): %w", err)
		}
	}

	api := datadogV2.NewUsageMeteringApi(client.V2())
	opts := datadogV2.GetMonthlyCostAttributionOptionalParameters{}
	if costEndMonth != "" {
		opts.WithEndMonth(endTime)
	}

	resp, r, err := api.GetMonthlyCostAttribution(client.Context(), startTime, costFields, opts)
	if err != nil {
		return formatAPIError("get cost attribution", err, r)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}

	printOutput("%s\n", output)
	return nil
}

func runCostByOrg(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	// Parse start month
	startTime, err := time.Parse("2006-01", costStartMonth)
	if err != nil {
		return fmt.Errorf("invalid start month format (use YYYY-MM): %w", err)
	}

	api := datadogV2.NewUsageMeteringApi(client.V2())

	// Call appropriate API based on view type
	var resp interface{}
	var r interface{}

	switch costView {
	case "actual":
		opts := datadogV2.GetCostByOrgOptionalParameters{}
		if costEndMonth != "" {
			endTime, err := time.Parse("2006-01", costEndMonth)
			if err != nil {
				return fmt.Errorf("invalid end month format (use YYYY-MM): %w", err)
			}
			opts.WithEndMonth(endTime)
		}
		resp, r, err = api.GetCostByOrg(client.Context(), startTime, opts)

	case "estimated":
		opts := datadogV2.GetEstimatedCostByOrgOptionalParameters{}
		opts.WithStartMonth(startTime)
		if costEndMonth != "" {
			endTime, err := time.Parse("2006-01", costEndMonth)
			if err != nil {
				return fmt.Errorf("invalid end month format (use YYYY-MM): %w", err)
			}
			opts.WithEndMonth(endTime)
		}
		resp, r, err = api.GetEstimatedCostByOrg(client.Context(), opts)

	case "historical":
		opts := datadogV2.GetHistoricalCostByOrgOptionalParameters{}
		if costEndMonth != "" {
			endTime, err := time.Parse("2006-01", costEndMonth)
			if err != nil {
				return fmt.Errorf("invalid end month format (use YYYY-MM): %w", err)
			}
			opts.WithEndMonth(endTime)
		}
		resp, r, err = api.GetHistoricalCostByOrg(client.Context(), startTime, opts)

	default:
		return fmt.Errorf("invalid view type '%s': must be actual, estimated, or historical", costView)
	}

	if err != nil {
		return formatAPIError(fmt.Sprintf("get %s cost by org", costView), err, r)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}

	printOutput("%s\n", output)
	return nil
}
