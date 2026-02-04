// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/DataDog/pup/pkg/formatter"
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
	auditLogsSearchCmd.MarkFlagRequired("query")

	auditLogsCmd.AddCommand(auditLogsListCmd, auditLogsSearchCmd)
}

func runAuditLogsList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewAuditApi(client.V2())
	opts := datadogV2.ListAuditLogsOptionalParameters{}

	filter := datadogV2.AuditLogsQueryFilter{}
	filter.SetFrom(auditLogsFrom)
	filter.SetTo(auditLogsTo)

	page := datadogV2.AuditLogsQueryPageOptions{}
	page.SetLimit(auditLogsLimit)

	opts.WithFilterQuery("*")
	opts.WithPageLimit(auditLogsLimit)

	resp, r, err := api.ListAuditLogs(client.Context(), opts)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to list audit logs: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to list audit logs: %w", err)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func runAuditLogsSearch(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewAuditApi(client.V2())
	body := datadogV2.AuditLogsSearchEventsRequest{}

	filter := datadogV2.AuditLogsQueryFilter{}
	filter.SetQuery(auditLogsQuery)
	filter.SetFrom(auditLogsFrom)
	filter.SetTo(auditLogsTo)
	body.SetFilter(filter)

	page := datadogV2.AuditLogsQueryPageOptions{}
	page.SetLimit(auditLogsLimit)
	body.SetPage(page)

	resp, r, err := api.SearchAuditLogs(client.Context(), datadogV2.SearchAuditLogsOptionalParameters{}.WithBody(body))
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to search audit logs: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to search audit logs: %w", err)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}
