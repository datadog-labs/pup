// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"fmt"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/DataDog/pup/pkg/formatter"
	"github.com/spf13/cobra"
)

var usageCmd = &cobra.Command{
	Use:   "usage",
	Short: "Query usage and billing information",
	Long: `Query usage metrics and billing information for your organization.

CAPABILITIES:
  • View usage summary
  • Get hourly usage
  • Track usage by product
  • Monitor cost attribution

EXAMPLES:
  # Get usage summary
  pup usage summary --start="2024-01-01" --end="2024-01-31"

  # Get hourly usage
  pup usage hourly --start="2024-01-01" --end="2024-01-02"

AUTHENTICATION:
  Requires either OAuth2 authentication or API keys with billing permissions.`,
}

var usageSummaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Get usage summary",
	RunE:  runUsageSummary,
}

var usageHourlyCmd = &cobra.Command{
	Use:   "hourly",
	Short: "Get hourly usage",
	RunE:  runUsageHourly,
}

var (
	usageStart string
	usageEnd   string
)

func init() {
	usageSummaryCmd.Flags().StringVar(&usageStart, "start", "", "Start date (YYYY-MM-DD)")
	usageSummaryCmd.Flags().StringVar(&usageEnd, "end", "", "End date (YYYY-MM-DD)")

	usageHourlyCmd.Flags().StringVar(&usageStart, "start", "", "Start date (YYYY-MM-DD)")
	usageHourlyCmd.Flags().StringVar(&usageEnd, "end", "", "End date (YYYY-MM-DD)")

	usageCmd.AddCommand(usageSummaryCmd, usageHourlyCmd)
}

func runUsageSummary(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	startTime, err := time.Parse("2006-01-02", usageStart)
	if err != nil {
		return fmt.Errorf("invalid start date: %w", err)
	}

	endTime, err := time.Parse("2006-01-02", usageEnd)
	if err != nil {
		return fmt.Errorf("invalid end date: %w", err)
	}

	api := datadogV1.NewUsageMeteringApi(client.V1())
	opts := datadogV1.NewGetUsageSummaryOptionalParameters()
	opts = opts.WithEndMonth(endTime)
	resp, r, err := api.GetUsageSummary(client.Context(), startTime, *opts)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to get usage summary: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to get usage summary: %w", err)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func runUsageHourly(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	startTime, err := time.Parse("2006-01-02", usageStart)
	if err != nil {
		return fmt.Errorf("invalid start date: %w", err)
	}

	endTime, err := time.Parse("2006-01-02", usageEnd)
	if err != nil {
		return fmt.Errorf("invalid end date: %w", err)
	}

	api := datadogV1.NewUsageMeteringApi(client.V1())
	opts := datadogV1.NewGetUsageHostsOptionalParameters()
	opts = opts.WithEndHr(endTime)
	resp, r, err := api.GetUsageHosts(client.Context(), startTime, *opts)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to get hourly usage: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to get hourly usage: %w", err)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}
