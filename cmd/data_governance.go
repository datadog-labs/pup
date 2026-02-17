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

var dataGovernanceCmd = &cobra.Command{
	Use:   "data-governance",
	Short: "Manage data governance",
	Long: `Manage data governance, sensitive data scanning, and data deletion.

CAPABILITIES:
  • Manage sensitive data scanner
  • Configure data deletion policies
  • View scan results
  • Manage scanning rules

EXAMPLES:
  # List scanning rules
  pup data-governance scanner rules list

  # Get rule details
  pup data-governance scanner rules get rule-id

AUTHENTICATION:
  Requires either OAuth2 authentication or API keys.`,
}

var dataGovernanceScannerCmd = &cobra.Command{
	Use:   "scanner",
	Short: "Manage sensitive data scanner",
}

var dataGovernanceScannerRulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "Manage scanning rules",
}

var dataGovernanceScannerRulesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List scanning rules",
	RunE:  runDataGovernanceScannerRulesList,
}

func init() {
	dataGovernanceScannerRulesCmd.AddCommand(dataGovernanceScannerRulesListCmd)
	dataGovernanceScannerCmd.AddCommand(dataGovernanceScannerRulesCmd)
	dataGovernanceCmd.AddCommand(dataGovernanceScannerCmd)
}

func runDataGovernanceScannerRulesList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewSensitiveDataScannerApi(client.V2())
	resp, r, err := api.ListScanningGroups(client.Context())
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to list scanning rules: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to list scanning rules: %w", err)
	}

	return formatAndPrint(resp, nil)
}
