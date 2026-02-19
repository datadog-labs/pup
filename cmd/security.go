// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"fmt"
	"io"
	"strings"

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

// Content Packs subcommands
var securityContentPacksCmd = &cobra.Command{
	Use:   "content-packs",
	Short: "Manage security content packs",
}

var securityContentPacksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List content pack states",
	RunE:  runSecurityContentPacksList,
}

var securityContentPacksActivateCmd = &cobra.Command{
	Use:   "activate [content-pack-id]",
	Short: "Activate a content pack",
	Args:  cobra.ExactArgs(1),
	RunE:  runSecurityContentPacksActivate,
}

var securityContentPacksDeactivateCmd = &cobra.Command{
	Use:   "deactivate [content-pack-id]",
	Short: "Deactivate a content pack",
	Args:  cobra.ExactArgs(1),
	RunE:  runSecurityContentPacksDeactivate,
}

// Bulk export subcommand
var securityRulesBulkExportCmd = &cobra.Command{
	Use:   "bulk-export",
	Short: "Bulk export security monitoring rules",
	RunE:  runSecurityRulesBulkExport,
}

// Risk Scores subcommands
var securityRiskScoresCmd = &cobra.Command{
	Use:   "risk-scores",
	Short: "Manage entity risk scores",
}

var securityRiskScoresListCmd = &cobra.Command{
	Use:   "list",
	Short: "List entity risk scores",
	RunE:  runSecurityRiskScoresList,
}

var (
	// Findings search flags
	findingsQuery string
	findingsLimit int32
	findingsSort  string

	// Bulk export flags
	securityRuleIDs string

	// Risk scores flags
	riskScoresQuery string
)

func init() {
	// Findings search flags
	securityFindingsSearchCmd.Flags().StringVar(&findingsQuery, "query", "", "Search query using log search syntax (required)")
	securityFindingsSearchCmd.Flags().Int32Var(&findingsLimit, "limit", 100, "Maximum results (1-1000)")
	securityFindingsSearchCmd.Flags().StringVar(&findingsSort, "sort", "", "Sort field: severity, status, timestamp")
	_ = securityFindingsSearchCmd.MarkFlagRequired("query")

	// Bulk export flags
	securityRulesBulkExportCmd.Flags().StringVar(&securityRuleIDs, "rule-ids", "", "Comma-separated rule IDs (required)")
	_ = securityRulesBulkExportCmd.MarkFlagRequired("rule-ids")

	// Risk scores flags
	securityRiskScoresListCmd.Flags().StringVar(&riskScoresQuery, "query", "", "Filter query")

	// Command hierarchy
	securityRulesCmd.AddCommand(securityRulesListCmd, securityRulesGetCmd, securityRulesBulkExportCmd)
	securitySignalsCmd.AddCommand(securitySignalsListCmd)
	securityFindingsCmd.AddCommand(securityFindingsSearchCmd)
	securityContentPacksCmd.AddCommand(securityContentPacksListCmd, securityContentPacksActivateCmd, securityContentPacksDeactivateCmd)
	securityRiskScoresCmd.AddCommand(securityRiskScoresListCmd)
	securityCmd.AddCommand(securityRulesCmd, securitySignalsCmd, securityFindingsCmd, securityContentPacksCmd, securityRiskScoresCmd)
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

// Content Packs implementations
func runSecurityContentPacksList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewSecurityMonitoringApi(client.V2())
	resp, r, err := api.GetContentPacksStates(client.Context())
	if err != nil {
		return formatAPIError("list content packs", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runSecurityContentPacksActivate(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewSecurityMonitoringApi(client.V2())
	r, err := api.ActivateContentPack(client.Context(), args[0])
	if err != nil {
		return formatAPIError("activate content pack", err, r)
	}

	printOutput("Content pack '%s' activated successfully.\n", args[0])
	return nil
}

func runSecurityContentPacksDeactivate(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewSecurityMonitoringApi(client.V2())
	r, err := api.DeactivateContentPack(client.Context(), args[0])
	if err != nil {
		return formatAPIError("deactivate content pack", err, r)
	}

	printOutput("Content pack '%s' deactivated successfully.\n", args[0])
	return nil
}

// Bulk Export implementation
func runSecurityRulesBulkExport(cmd *cobra.Command, args []string) error {
	ruleIDs := strings.Split(securityRuleIDs, ",")
	for i := range ruleIDs {
		ruleIDs[i] = strings.TrimSpace(ruleIDs[i])
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewSecurityMonitoringApi(client.V2())
	attrs := *datadogV2.NewSecurityMonitoringRuleBulkExportAttributes(ruleIDs)
	data := *datadogV2.NewSecurityMonitoringRuleBulkExportData(attrs, datadogV2.SECURITYMONITORINGRULEBULKEXPORTDATATYPE_SECURITY_MONITORING_RULES_BULK_EXPORT)
	body := *datadogV2.NewSecurityMonitoringRuleBulkExportPayload(data)

	resp, r, err := api.BulkExportSecurityMonitoringRules(client.Context(), body)
	if err != nil {
		return formatAPIError("bulk export security rules", err, r)
	}

	// resp is an io.Reader, read and output
	output, err := io.ReadAll(resp)
	if err != nil {
		return fmt.Errorf("failed to read export data: %w", err)
	}

	printOutput("%s\n", string(output))
	return nil
}

// Risk Scores implementation
func runSecurityRiskScoresList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewEntityRiskScoresApi(client.V2())
	opts := datadogV2.NewListEntityRiskScoresOptionalParameters()
	if riskScoresQuery != "" {
		opts = opts.WithFilterQuery(riskScoresQuery)
	}

	resp, r, err := api.ListEntityRiskScores(client.Context(), *opts)
	if err != nil {
		return formatAPIError("list entity risk scores", err, r)
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
