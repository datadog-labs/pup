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
	saRepository string
	saBranch     string
	saLanguage   string
	saFrom       string
	saTo         string
	saSeverity   string
	saStatus     string
)

func init() {
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

	staticAnalysisASTCmd.AddCommand(staticAnalysisASTListCmd, staticAnalysisASTGetCmd)
	staticAnalysisCustomRulesetsCmd.AddCommand(staticAnalysisCustomRulesetsListCmd, staticAnalysisCustomRulesetsGetCmd)
	staticAnalysisSCACmd.AddCommand(staticAnalysisSCAListCmd, staticAnalysisSCAGetCmd)
	staticAnalysisCoverageCmd.AddCommand(staticAnalysisCoverageListCmd, staticAnalysisCoverageGetCmd)
	staticAnalysisCmd.AddCommand(staticAnalysisASTCmd, staticAnalysisCustomRulesetsCmd, staticAnalysisSCACmd, staticAnalysisCoverageCmd)
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

	return formatAndPrint(result, nil)
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

	return formatAndPrint(result, nil)
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

	return formatAndPrint(resp, nil)
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

	return formatAndPrint(resp, nil)
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

	return formatAndPrint(result, nil)
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

	return formatAndPrint(result, nil)
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

	return formatAndPrint(result, nil)
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

	return formatAndPrint(result, nil)
}
