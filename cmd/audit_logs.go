// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/datadog-labs/pup/pkg/util"
	"github.com/spf13/cobra"
)

var auditLogsCmd = &cobra.Command{
	Use:   "audit-logs",
	Short: "Query audit logs",
	Long: `Search and list audit logs for your Datadog organization.

Audit logs track all actions performed in your Datadog organization,
providing a complete audit trail for compliance and security.

CAPABILITIES:
  • Search audit logs with queries
  • List recent audit events
  • Filter by action, user, resource, outcome

EXAMPLES:
  # List recent audit logs
  pup audit-logs list

  # Search for specific user actions
  pup audit-logs search --query="@usr.name:admin@example.com"

  # Search for failed actions
  pup audit-logs search --query="@evt.outcome:error"

AUTHENTICATION:
  Requires either OAuth2 authentication or API keys.`,
}

var auditLogsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List recent audit logs",
	RunE:  runAuditLogsList,
}

var auditLogsSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search audit logs",
	RunE:  runAuditLogsSearch,
}

var (
	auditLogsQuery string
	auditLogsFrom  string
	auditLogsTo    string
	auditLogsLimit int32
)

func init() {
	auditLogsListCmd.Flags().StringVar(&auditLogsFrom, "from", "1h", "Start time")
	auditLogsListCmd.Flags().StringVar(&auditLogsTo, "to", "now", "End time")
	auditLogsListCmd.Flags().Int32Var(&auditLogsLimit, "limit", 100, "Maximum results")

	auditLogsSearchCmd.Flags().StringVar(&auditLogsQuery, "query", "", "Search query (required)")
	auditLogsSearchCmd.Flags().StringVar(&auditLogsFrom, "from", "1h", "Start time")
	auditLogsSearchCmd.Flags().StringVar(&auditLogsTo, "to", "now", "End time")
	auditLogsSearchCmd.Flags().Int32Var(&auditLogsLimit, "limit", 100, "Maximum results")
	if err := auditLogsSearchCmd.MarkFlagRequired("query"); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}

	auditLogsCmd.AddCommand(auditLogsListCmd, auditLogsSearchCmd)
}

func runAuditLogsList(cmd *cobra.Command, args []string) error {
	fromTime, err := util.ParseTimeParam(auditLogsFrom)
	if err != nil {
		return fmt.Errorf("invalid --from time: %w", err)
	}
	toTime, err := util.ParseTimeParam(auditLogsTo)
	if err != nil {
		return fmt.Errorf("invalid --to time: %w", err)
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewAuditApi(client.V2())
	opts := datadogV2.ListAuditLogsOptionalParameters{}

	opts.WithFilterQuery("*")
	opts.WithFilterFrom(fromTime)
	opts.WithFilterTo(toTime)
	opts.WithPageLimit(auditLogsLimit)

	resp, r, err := api.ListAuditLogs(client.Context(), opts)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to list audit logs: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to list audit logs: %w", err)
	}

	return formatAndPrint(resp, nil)
}

func runAuditLogsSearch(cmd *cobra.Command, args []string) error {
	fromTime, err := util.ParseTimeToUnixMilli(auditLogsFrom)
	if err != nil {
		return fmt.Errorf("invalid --from time: %w", err)
	}
	toTime, err := util.ParseTimeToUnixMilli(auditLogsTo)
	if err != nil {
		return fmt.Errorf("invalid --to time: %w", err)
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewAuditApi(client.V2())
	body := datadogV2.AuditLogsSearchEventsRequest{}

	from := fmt.Sprintf("%d", fromTime)
	to := fmt.Sprintf("%d", toTime)

	filter := datadogV2.AuditLogsQueryFilter{}
	filter.SetQuery(auditLogsQuery)
	filter.SetFrom(from)
	filter.SetTo(to)
	body.SetFilter(filter)

	page := datadogV2.AuditLogsQueryPageOptions{}
	page.SetLimit(auditLogsLimit)
	body.SetPage(page)

	opts := datadogV2.SearchAuditLogsOptionalParameters{
		Body: &body,
	}
	resp, r, err := api.SearchAuditLogs(client.Context(), opts)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to search audit logs: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to search audit logs: %w", err)
	}

	return formatAndPrint(resp, nil)
}
