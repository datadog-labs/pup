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
}

var securityFindingsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List security findings",
	RunE:  runSecurityFindingsList,
}

func init() {
	securityRulesCmd.AddCommand(securityRulesListCmd, securityRulesGetCmd)
	securitySignalsCmd.AddCommand(securitySignalsListCmd)
	securityFindingsCmd.AddCommand(securityFindingsListCmd)
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

	output, err := formatter.ToJSON(resp)
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

	output, err := formatter.ToJSON(resp)
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

	output, err := formatter.ToJSON(resp)
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
	resp, r, err := api.ListFindings(client.Context())
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to list security findings: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to list security findings: %w", err)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}
