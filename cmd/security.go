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

var securityFindingsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List security findings",
	Long: `List security findings with optional filtering and pagination.

EXAMPLES:
  # List all findings
  pup security findings list

  # Filter by status
  pup security findings list --status=critical

  # Paginate results
  pup security findings list --page-size=50 --page-number=1`,
	RunE: runSecurityFindingsList,
}

var securityFindingsGetCmd = &cobra.Command{
	Use:   "get [finding-id]",
	Short: "Get security finding details",
	Long: `Get detailed information about a specific security finding.

EXAMPLES:
  # Get finding details
  pup security findings get finding-abc-123

  # Get finding with table output
  pup security findings get finding-abc-123 --output=table`,
	Args: cobra.ExactArgs(1),
	RunE: runSecurityFindingsGet,
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
	// Findings list flags
	findingsPageSize     int64
	findingsPageNumber   int64
	findingsStatus       string
	findingsEvaluation   string
	findingsRuleID       string
	findingsResourceType string

	// Findings search flags
	findingsQuery      string
	findingsLimit      int32
	findingsPageCursor string
	findingsSort       string
)

func init() {
	// Findings list flags
	securityFindingsListCmd.Flags().Int64Var(&findingsPageSize, "page-size", 100, "Number of findings per page (max: 1000)")
	securityFindingsListCmd.Flags().StringVar(&findingsPageCursor, "page-cursor", "", "Page cursor for pagination")
	securityFindingsListCmd.Flags().StringVar(&findingsStatus, "status", "", "Filter by status: critical, high, medium, low, info")
	securityFindingsListCmd.Flags().StringVar(&findingsEvaluation, "evaluation", "", "Filter by evaluation: pass, fail")
	securityFindingsListCmd.Flags().StringVar(&findingsRuleID, "rule-id", "", "Filter by rule ID")
	securityFindingsListCmd.Flags().StringVar(&findingsResourceType, "resource-type", "", "Filter by resource type")

	// Findings search flags
	securityFindingsSearchCmd.Flags().StringVar(&findingsQuery, "query", "", "Search query using log search syntax (required)")
	securityFindingsSearchCmd.Flags().Int32Var(&findingsLimit, "limit", 100, "Maximum results (1-1000)")
	securityFindingsSearchCmd.Flags().StringVar(&findingsSort, "sort", "", "Sort field: severity, status, timestamp")
	_ = securityFindingsSearchCmd.MarkFlagRequired("query")

	// Command hierarchy
	securityRulesCmd.AddCommand(securityRulesListCmd, securityRulesGetCmd)
	securitySignalsCmd.AddCommand(securitySignalsListCmd)
	securityFindingsCmd.AddCommand(securityFindingsListCmd, securityFindingsGetCmd, securityFindingsSearchCmd)
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

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
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

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
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

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func runSecurityFindingsList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewSecurityMonitoringApi(client.V2())

	// Build optional parameters with filtering
	opts := datadogV2.ListFindingsOptionalParameters{}

	if findingsPageSize > 0 {
		if findingsPageSize > 1000 {
			findingsPageSize = 1000
		}
		opts.WithPageLimit(findingsPageSize)
	}

	if findingsPageCursor != "" {
		opts.WithPageCursor(findingsPageCursor)
	}

	if findingsStatus != "" {
		status, err := datadogV2.NewFindingStatusFromValue(findingsStatus)
		if err != nil {
			return fmt.Errorf("invalid status value '%s': must be one of critical, high, medium, low, info", findingsStatus)
		}
		opts.WithFilterStatus(*status)
	}

	if findingsEvaluation != "" {
		evaluation, err := datadogV2.NewFindingEvaluationFromValue(findingsEvaluation)
		if err != nil {
			return fmt.Errorf("invalid evaluation value '%s': must be one of pass, fail", findingsEvaluation)
		}
		opts.WithFilterEvaluation(*evaluation)
	}

	if findingsRuleID != "" {
		opts.WithFilterRuleId(findingsRuleID)
	}

	if findingsResourceType != "" {
		opts.WithFilterResourceType(findingsResourceType)
	}

	resp, r, err := api.ListFindings(client.Context(), opts)
	if err != nil {
		return formatAPIError("list security findings", err, r)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}
	printOutput("%s\n", output)
	return nil
}

func runSecurityFindingsGet(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	findingID := args[0]
	api := datadogV2.NewSecurityMonitoringApi(client.V2())

	resp, r, err := api.GetFinding(client.Context(), findingID)
	if err != nil {
		return formatAPIError("get security finding", err, r)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}
	printOutput("%s\n", output)
	return nil
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

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}
	printOutput("%s\n", output)
	return nil
}
