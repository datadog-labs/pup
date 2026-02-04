// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/DataDog/pup/pkg/formatter"
	"github.com/spf13/cobra"
)

var integrationsCmd = &cobra.Command{
	Use:   "integrations",
	Short: "Manage third-party integrations",
	Long: `Manage third-party integrations with external services.

Integrations connect Datadog with external services like Slack, PagerDuty,
Jira, and many others for notifications and workflow automation.

CAPABILITIES:
  • List Slack integrations
  • Manage PagerDuty integrations
  • Configure webhook integrations
  • View integration status

EXAMPLES:
  # List Slack integrations
  pup integrations slack list

  # List PagerDuty integrations
  pup integrations pagerduty list

  # List webhooks
  pup integrations webhooks list

AUTHENTICATION:
  Requires either OAuth2 authentication or API keys.`,
}

var integrationsSlackCmd = &cobra.Command{
	Use:   "slack",
	Short: "Manage Slack integration",
}

var integrationsSlackListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Slack channels",
	RunE:  runIntegrationsSlackList,
}

var integrationsPagerDutyCmd = &cobra.Command{
	Use:   "pagerduty",
	Short: "Manage PagerDuty integration",
}

var integrationsPagerDutyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List PagerDuty services",
	RunE:  runIntegrationsPagerDutyList,
}

var integrationsWebhooksCmd = &cobra.Command{
	Use:   "webhooks",
	Short: "Manage webhooks",
}

var integrationsWebhooksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List webhooks",
	RunE:  runIntegrationsWebhooksList,
}

func init() {
	integrationsSlackCmd.AddCommand(integrationsSlackListCmd)
	integrationsPagerDutyCmd.AddCommand(integrationsPagerDutyListCmd)
	integrationsWebhooksCmd.AddCommand(integrationsWebhooksListCmd)
	integrationsCmd.AddCommand(integrationsSlackCmd, integrationsPagerDutyCmd, integrationsWebhooksCmd)
}

func runIntegrationsSlackList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV1.NewSlackIntegrationApi(client.V1())
	resp, r, err := api.GetSlackIntegrationChannels(client.Context(), "main")
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to list Slack channels: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to list Slack channels: %w", err)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func runIntegrationsPagerDutyList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV1.NewPagerDutyIntegrationApi(client.V1())
	resp, r, err := api.GetPagerDutyIntegrationServices(client.Context())
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to list PagerDuty services: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to list PagerDuty services: %w", err)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func runIntegrationsWebhooksList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV1.NewWebhooksIntegrationApi(client.V1())
	resp, r, err := api.GetWebhooksIntegration(client.Context(), "main")
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to list webhooks: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to list webhooks: %w", err)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}
