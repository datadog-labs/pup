// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/spf13/cobra"
)

var securityCmd = &cobra.Command{
	Use:   "security",
	Short: "Manage security monitoring",
	Long: `Manage security monitoring rules, signals, and findings.

CAPABILITIES:
  • List and manage security monitoring rules
  • View security signals and findings
  • Configure suppression rules
  • Manage security filters

EXAMPLES:
  # List security monitoring rules
  pup security rules list

  # Get rule details
  pup security rules get rule-id

  # List security signals
  pup security signals list

AUTHENTICATION:
  Requires either OAuth2 authentication or API keys.`,
}

var securityRulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "Manage security rules",
}

var securityRulesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List security rules",
	RunE:  runSecurityRulesList,
}

var securityRulesGetCmd = &cobra.Command{
	Use:   "get [rule-id]",
	Short: "Get rule details",
	Args:  cobra.ExactArgs(1),
	RunE:  runSecurityRulesGet,
}

var securitySignalsCmd = &cobra.Command{
	Use:   "signals",
	Short: "Manage security signals",
}

var securitySignalsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List security signals",
	RunE:  runSecuritySignalsList,
}

var securityFindingsCmd = &cobra.Command{
	Use:   "findings",
	Short: "Manage security findings",
	Long: `Manage security findings from Datadog's Security Findings API.

Security findings provide insights into security posture and vulnerabilities
across your infrastructure and applications.`,
}

var securityFindingsSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search security findings",
	Long: `Search security findings using log search syntax.

QUERY SYNTAX (using log search syntax):
  • @severity:(critical OR high) - Filter by severity level
  • @status:open - Filter by status
  • @attributes.resource_type:s3_bucket - Filter by resource type
  • team:platform - Filter by tags (no @ prefix)
  • AND, OR, NOT - Boolean operators

EXAMPLES:
  # Search critical or high severity findings
  pup security findings search --query="@severity:(critical OR high)"

  # Search open findings with specific resource type and team tag
  pup security findings search --query="@status:open @attributes.resource_type:s3_bucket team:platform"

  # Limit results
  pup security findings search --query="@severity:critical" --limit=50`,
	RunE: runSecurityFindingsSearch,
}

var (
	// Findings search flags
	findingsQuery string
	findingsLimit int32
	findingsSort  string
)

func init() {
	// Findings search flags
	securityFindingsSearchCmd.Flags().StringVar(&findingsQuery, "query", "", "Search query using log search syntax (required)")
	securityFindingsSearchCmd.Flags().Int32Var(&findingsLimit, "limit", 100, "Maximum results (1-1000)")
	securityFindingsSearchCmd.Flags().StringVar(&findingsSort, "sort", "", "Sort field: severity, status, timestamp")
	_ = securityFindingsSearchCmd.MarkFlagRequired("query")

	// Command hierarchy
	securityRulesCmd.AddCommand(securityRulesListCmd, securityRulesGetCmd)
	securitySignalsCmd.AddCommand(securitySignalsListCmd)
	securityFindingsCmd.AddCommand(securityFindingsSearchCmd)
	securityCmd.AddCommand(securityRulesCmd, securitySignalsCmd, securityFindingsCmd)
}

func runSecurityRulesList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewSecurityMonitoringApi(client.V2())
	resp, r, err := api.ListSecurityMonitoringRules(client.Context())
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to list security rules: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to list security rules: %w", err)
	}

	return formatAndPrint(resp, nil)
}

func runSecurityRulesGet(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	ruleID := args[0]
	api := datadogV2.NewSecurityMonitoringApi(client.V2())
	resp, r, err := api.GetSecurityMonitoringRule(client.Context(), ruleID)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to get security rule: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to get security rule: %w", err)
	}

	return formatAndPrint(resp, nil)
}

func runSecuritySignalsList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewSecurityMonitoringApi(client.V2())
	resp, r, err := api.ListSecurityMonitoringSignals(client.Context())
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to list security signals: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to list security signals: %w", err)
	}

	return formatAndPrint(resp, nil)
}

func runSecurityFindingsSearch(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewSecurityMonitoringApi(client.V2())

	// Build search request
	searchReq := datadogV2.NewSecurityFindingsSearchRequest()
	searchData := datadogV2.NewSecurityFindingsSearchRequestData()
	searchAttrs := datadogV2.NewSecurityFindingsSearchRequestDataAttributes()

	// Set filter query
	searchAttrs.SetFilter(findingsQuery)

	// Set pagination
	if findingsLimit > 0 {
		page := datadogV2.NewSecurityFindingsSearchRequestPage()
		page.SetLimit(int64(findingsLimit))
		searchAttrs.SetPage(*page)
	}

	// Set sort if specified
	if findingsSort != "" {
		sort, err := datadogV2.NewSecurityFindingsSortFromValue(findingsSort)
		if err != nil {
			return fmt.Errorf("invalid sort value '%s'", findingsSort)
		}
		searchAttrs.SetSort(*sort)
	}

	searchData.SetAttributes(*searchAttrs)
	searchReq.SetData(*searchData)

	resp, r, err := api.SearchSecurityFindings(client.Context(), *searchReq)
	if err != nil {
		return formatAPIError("search security findings", err, r)
	}

	return formatAndPrint(resp, nil)
}
