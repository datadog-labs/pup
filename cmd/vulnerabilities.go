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

var vulnerabilitiesCmd = &cobra.Command{
	Use:   "vulnerabilities",
	Short: "Manage security vulnerabilities",
	Long: `Search and list security vulnerabilities in your applications.

CAPABILITIES:
  • Search vulnerabilities with custom queries
  • Filter by severity, status, service, and repository
  • Track vulnerability remediation status

EXAMPLES:
  # Search for critical vulnerabilities
  pup vulnerabilities search --query="severity:critical"

  # List all open vulnerabilities
  pup vulnerabilities list --severity="critical,high" --status="open"

AUTHENTICATION:
  Requires either OAuth2 authentication or API keys.`,
}

var vulnerabilitiesSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search vulnerabilities",
	RunE:  runVulnerabilitiesSearch,
}

var vulnerabilitiesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List vulnerabilities",
	RunE:  runVulnerabilitiesList,
}

var staticAnalysisCmd = &cobra.Command{
	Use:   "static-analysis",
	Short: "Manage static analysis",
	Long: `Manage static analysis for code security and quality.

CAPABILITIES:
  • AST analysis results
  • Custom security rulesets
  • Software Composition Analysis (SCA)
  • Code coverage analysis

EXAMPLES:
  # List custom rulesets
  pup static-analysis custom-rulesets list

  # Get ruleset details
  pup static-analysis custom-rulesets get abc-123`,
}

var staticAnalysisASTCmd = &cobra.Command{
	Use:   "ast",
	Short: "AST analysis",
}

var staticAnalysisCustomRulesetsCmd = &cobra.Command{
	Use:   "custom-rulesets",
	Short: "Custom security rulesets",
}

var staticAnalysisSCACmd = &cobra.Command{
	Use:   "sca",
	Short: "Software Composition Analysis",
}

var staticAnalysisCoverageCmd = &cobra.Command{
	Use:   "coverage",
	Short: "Code coverage analysis",
}

var staticAnalysisASTListCmd = &cobra.Command{
	Use:   "list",
	Short: "List AST analyses",
	RunE:  runStaticAnalysisASTList,
}

var staticAnalysisASTGetCmd = &cobra.Command{
	Use:   "get [analysis-id]",
	Short: "Get AST analysis details",
	Args:  cobra.ExactArgs(1),
	RunE:  runStaticAnalysisASTGet,
}

var staticAnalysisCustomRulesetsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List custom rulesets",
	RunE:  runStaticAnalysisCustomRulesetsList,
}

var staticAnalysisCustomRulesetsGetCmd = &cobra.Command{
	Use:   "get [ruleset-id]",
	Short: "Get custom ruleset details",
	Args:  cobra.ExactArgs(1),
	RunE:  runStaticAnalysisCustomRulesetsGet,
}

var staticAnalysisSCAListCmd = &cobra.Command{
	Use:   "list",
	Short: "List SCA results",
	RunE:  runStaticAnalysisSCAList,
}

var staticAnalysisSCAGetCmd = &cobra.Command{
	Use:   "get [scan-id]",
	Short: "Get SCA scan details",
	Args:  cobra.ExactArgs(1),
	RunE:  runStaticAnalysisSCAGet,
}

var staticAnalysisCoverageListCmd = &cobra.Command{
	Use:   "list",
	Short: "List coverage analyses",
	RunE:  runStaticAnalysisCoverageList,
}

var staticAnalysisCoverageGetCmd = &cobra.Command{
	Use:   "get [coverage-id]",
	Short: "Get coverage analysis details",
	Args:  cobra.ExactArgs(1),
	RunE:  runStaticAnalysisCoverageGet,
}

var (
	vulnQuery      string
	vulnSeverity   string
	vulnStatus     string
	vulnService    string
	vulnRepository string
	vulnFrom       string
	vulnTo         string
	vulnLimit      int32
	vulnOffset     int32
	saRepository   string
	saBranch       string
	saLanguage     string
	saFrom         string
	saTo           string
	saSeverity     string
	saStatus       string
)

func init() {
	vulnerabilitiesSearchCmd.Flags().StringVarP(&vulnQuery, "query", "q", "", "Search query (required)")
	vulnerabilitiesSearchCmd.Flags().StringVar(&vulnFrom, "from", "", "Start time")
	vulnerabilitiesSearchCmd.Flags().StringVar(&vulnTo, "to", "", "End time")
	vulnerabilitiesSearchCmd.Flags().Int32Var(&vulnLimit, "limit", 100, "Maximum results")
	vulnerabilitiesSearchCmd.Flags().Int32Var(&vulnOffset, "offset", 0, "Results offset")
	vulnerabilitiesSearchCmd.MarkFlagRequired("query")

	vulnerabilitiesListCmd.Flags().StringVar(&vulnSeverity, "severity", "", "Filter by severity")
	vulnerabilitiesListCmd.Flags().StringVar(&vulnStatus, "status", "", "Filter by status")
	vulnerabilitiesListCmd.Flags().StringVar(&vulnService, "service", "", "Filter by service")
	vulnerabilitiesListCmd.Flags().StringVar(&vulnRepository, "repository", "", "Filter by repository")
	vulnerabilitiesListCmd.Flags().Int32Var(&vulnLimit, "limit", 100, "Maximum results")
	vulnerabilitiesListCmd.Flags().Int32Var(&vulnOffset, "offset", 0, "Results offset")

	staticAnalysisASTListCmd.Flags().StringVar(&saRepository, "repository", "", "Filter by repository")
	staticAnalysisASTListCmd.Flags().StringVar(&saBranch, "branch", "", "Filter by branch")
	staticAnalysisASTListCmd.Flags().StringVar(&saLanguage, "language", "", "Filter by language")
	staticAnalysisASTListCmd.Flags().StringVar(&saFrom, "from", "", "Start time")
	staticAnalysisASTListCmd.Flags().StringVar(&saTo, "to", "", "End time")

	staticAnalysisSCAListCmd.Flags().StringVar(&saSeverity, "severity", "", "Filter by severity")
	staticAnalysisSCAListCmd.Flags().StringVar(&saStatus, "status", "", "Filter by status")
	staticAnalysisSCAListCmd.Flags().StringVar(&saRepository, "repository", "", "Filter by repository")

	staticAnalysisCoverageListCmd.Flags().StringVar(&saRepository, "repository", "", "Filter by repository")
	staticAnalysisCoverageListCmd.Flags().StringVar(&saBranch, "branch", "", "Filter by branch")
	staticAnalysisCoverageListCmd.Flags().StringVar(&saFrom, "from", "", "Start time")
	staticAnalysisCoverageListCmd.Flags().StringVar(&saTo, "to", "", "End time")

	vulnerabilitiesCmd.AddCommand(vulnerabilitiesSearchCmd, vulnerabilitiesListCmd)
	staticAnalysisASTCmd.AddCommand(staticAnalysisASTListCmd, staticAnalysisASTGetCmd)
	staticAnalysisCustomRulesetsCmd.AddCommand(staticAnalysisCustomRulesetsListCmd, staticAnalysisCustomRulesetsGetCmd)
	staticAnalysisSCACmd.AddCommand(staticAnalysisSCAListCmd, staticAnalysisSCAGetCmd)
	staticAnalysisCoverageCmd.AddCommand(staticAnalysisCoverageListCmd, staticAnalysisCoverageGetCmd)
	staticAnalysisCmd.AddCommand(staticAnalysisASTCmd, staticAnalysisCustomRulesetsCmd, staticAnalysisSCACmd, staticAnalysisCoverageCmd)
}

func runVulnerabilitiesSearch(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewSecurityMonitoringApi(client.V2())
	body := datadogV2.SecurityMonitoringSignalListRequest{
		Filter: &datadogV2.SecurityMonitoringSignalListRequestFilter{
			Query: &vulnQuery,
		},
		Page: &datadogV2.SecurityMonitoringSignalListRequestPage{},
	}

	if vulnFrom != "" {
		body.Filter.From = &vulnFrom
	}
	if vulnTo != "" {
		body.Filter.To = &vulnTo
	}
	if vulnLimit > 0 {
		body.Page.Limit = &vulnLimit
	}
	if vulnOffset > 0 {
		body.Page.Offset = &vulnOffset
	}

	resp, r, err := api.SearchSecurityMonitoringSignals(client.Context(), datadogV2.SearchSecurityMonitoringSignalsOptionalParameters{}.WithBody(body))
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to search vulnerabilities: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to search vulnerabilities: %w", err)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func runVulnerabilitiesList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewSecurityMonitoringApi(client.V2())
	opts := datadogV2.ListSecurityMonitoringSignalsOptionalParameters{}

	var queryParts []string
	if vulnSeverity != "" {
		queryParts = append(queryParts, fmt.Sprintf("severity:(%s)", vulnSeverity))
	}
	if vulnStatus != "" {
		queryParts = append(queryParts, fmt.Sprintf("status:(%s)", vulnStatus))
	}
	if vulnService != "" {
		queryParts = append(queryParts, fmt.Sprintf("service:%s", vulnService))
	}
	if vulnRepository != "" {
		queryParts = append(queryParts, fmt.Sprintf("repository:%s", vulnRepository))
	}

	if len(queryParts) > 0 {
		filterQuery := queryParts[0]
		for i := 1; i < len(queryParts); i++ {
			filterQuery = fmt.Sprintf("%s AND %s", filterQuery, queryParts[i])
		}
		opts = opts.WithFilterQuery(filterQuery)
	}

	if vulnLimit > 0 {
		opts = opts.WithPageLimit(int64(vulnLimit))
	}

	resp, r, err := api.ListSecurityMonitoringSignals(client.Context(), opts)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to list vulnerabilities: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to list vulnerabilities: %w", err)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func runStaticAnalysisASTList(cmd *cobra.Command, args []string) error {
	result := map[string]interface{}{
		"data": []map[string]interface{}{},
		"meta": map[string]interface{}{
			"message": "AST analysis list - API endpoint implementation pending",
			"filters": map[string]string{
				"repository": saRepository,
				"branch":     saBranch,
				"language":   saLanguage,
				"from":       saFrom,
				"to":         saTo,
			},
		},
	}

	output, err := formatter.ToJSON(result)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func runStaticAnalysisASTGet(cmd *cobra.Command, args []string) error {
	analysisID := args[0]
	result := map[string]interface{}{
		"data": map[string]interface{}{
			"id":   analysisID,
			"type": "ast_analysis",
			"attributes": map[string]interface{}{
				"message": "AST analysis details - API endpoint implementation pending",
			},
		},
	}

	output, err := formatter.ToJSON(result)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func runStaticAnalysisCustomRulesetsList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewSecurityMonitoringApi(client.V2())
	resp, r, err := api.ListSecurityMonitoringRules(client.Context())
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to list custom rulesets: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to list custom rulesets: %w", err)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func runStaticAnalysisCustomRulesetsGet(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	rulesetID := args[0]
	api := datadogV2.NewSecurityMonitoringApi(client.V2())
	resp, r, err := api.GetSecurityMonitoringRule(client.Context(), rulesetID)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to get custom ruleset: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to get custom ruleset: %w", err)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func runStaticAnalysisSCAList(cmd *cobra.Command, args []string) error {
	result := map[string]interface{}{
		"data": []map[string]interface{}{},
		"meta": map[string]interface{}{
			"message": "SCA results list - API endpoint implementation pending",
			"filters": map[string]string{
				"severity":   saSeverity,
				"status":     saStatus,
				"repository": saRepository,
			},
		},
	}

	output, err := formatter.ToJSON(result)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func runStaticAnalysisSCAGet(cmd *cobra.Command, args []string) error {
	scanID := args[0]
	result := map[string]interface{}{
		"data": map[string]interface{}{
			"id":   scanID,
			"type": "sca_scan",
			"attributes": map[string]interface{}{
				"message": "SCA scan details - API endpoint implementation pending",
			},
		},
	}

	output, err := formatter.ToJSON(result)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func runStaticAnalysisCoverageList(cmd *cobra.Command, args []string) error {
	result := map[string]interface{}{
		"data": []map[string]interface{}{},
		"meta": map[string]interface{}{
			"message": "Coverage analysis list - API endpoint implementation pending",
			"filters": map[string]string{
				"repository": saRepository,
				"branch":     saBranch,
				"from":       saFrom,
				"to":         saTo,
			},
		},
	}

	output, err := formatter.ToJSON(result)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func runStaticAnalysisCoverageGet(cmd *cobra.Command, args []string) error {
	coverageID := args[0]
	result := map[string]interface{}{
		"data": map[string]interface{}{
			"id":   coverageID,
			"type": "coverage_analysis",
			"attributes": map[string]interface{}{
				"message": "Coverage analysis details - API endpoint implementation pending",
			},
		},
	}

	output, err := formatter.ToJSON(result)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}
