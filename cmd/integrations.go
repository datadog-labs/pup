// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
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

// Jira subcommands
var integrationsJiraCmd = &cobra.Command{
	Use:   "jira",
	Short: "Manage Jira integration",
}

var integrationsJiraAccountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "Manage Jira accounts",
}

var integrationsJiraAccountsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Jira accounts",
	RunE:  runIntegrationsJiraAccountsList,
}

var integrationsJiraAccountsDeleteCmd = &cobra.Command{
	Use:   "delete [account-id]",
	Short: "Delete a Jira account",
	Args:  cobra.ExactArgs(1),
	RunE:  runIntegrationsJiraAccountsDelete,
}

var integrationsJiraTemplatesCmd = &cobra.Command{
	Use:   "templates",
	Short: "Manage Jira issue templates",
}

var integrationsJiraTemplatesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Jira issue templates",
	RunE:  runIntegrationsJiraTemplatesList,
}

var integrationsJiraTemplatesGetCmd = &cobra.Command{
	Use:   "get [template-id]",
	Short: "Get Jira issue template",
	Args:  cobra.ExactArgs(1),
	RunE:  runIntegrationsJiraTemplatesGet,
}

var integrationsJiraTemplatesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create Jira issue template",
	RunE:  runIntegrationsJiraTemplatesCreate,
}

var integrationsJiraTemplatesUpdateCmd = &cobra.Command{
	Use:   "update [template-id]",
	Short: "Update Jira issue template",
	Args:  cobra.ExactArgs(1),
	RunE:  runIntegrationsJiraTemplatesUpdate,
}

var integrationsJiraTemplatesDeleteCmd = &cobra.Command{
	Use:   "delete [template-id]",
	Short: "Delete Jira issue template",
	Args:  cobra.ExactArgs(1),
	RunE:  runIntegrationsJiraTemplatesDelete,
}

// ServiceNow subcommands
var integrationsServiceNowCmd = &cobra.Command{
	Use:   "servicenow",
	Short: "Manage ServiceNow integration",
}

var integrationsServiceNowInstancesCmd = &cobra.Command{
	Use:   "instances",
	Short: "Manage ServiceNow instances",
}

var integrationsServiceNowInstancesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List ServiceNow instances",
	RunE:  runIntegrationsServiceNowInstancesList,
}

var integrationsServiceNowTemplatesCmd = &cobra.Command{
	Use:   "templates",
	Short: "Manage ServiceNow templates",
}

var integrationsServiceNowTemplatesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List ServiceNow templates",
	RunE:  runIntegrationsServiceNowTemplatesList,
}

var integrationsServiceNowTemplatesGetCmd = &cobra.Command{
	Use:   "get [template-id]",
	Short: "Get ServiceNow template",
	Args:  cobra.ExactArgs(1),
	RunE:  runIntegrationsServiceNowTemplatesGet,
}

var integrationsServiceNowTemplatesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create ServiceNow template",
	RunE:  runIntegrationsServiceNowTemplatesCreate,
}

var integrationsServiceNowTemplatesUpdateCmd = &cobra.Command{
	Use:   "update [template-id]",
	Short: "Update ServiceNow template",
	Args:  cobra.ExactArgs(1),
	RunE:  runIntegrationsServiceNowTemplatesUpdate,
}

var integrationsServiceNowTemplatesDeleteCmd = &cobra.Command{
	Use:   "delete [template-id]",
	Short: "Delete ServiceNow template",
	Args:  cobra.ExactArgs(1),
	RunE:  runIntegrationsServiceNowTemplatesDelete,
}

var integrationsServiceNowUsersCmd = &cobra.Command{
	Use:   "users",
	Short: "Manage ServiceNow users",
}

var integrationsServiceNowUsersListCmd = &cobra.Command{
	Use:   "list [instance-id]",
	Short: "List ServiceNow users",
	Args:  cobra.ExactArgs(1),
	RunE:  runIntegrationsServiceNowUsersList,
}

var integrationsServiceNowAssignmentGroupsCmd = &cobra.Command{
	Use:   "assignment-groups",
	Short: "Manage ServiceNow assignment groups",
}

var integrationsServiceNowAssignmentGroupsListCmd = &cobra.Command{
	Use:   "list [instance-id]",
	Short: "List ServiceNow assignment groups",
	Args:  cobra.ExactArgs(1),
	RunE:  runIntegrationsServiceNowAssignmentGroupsList,
}

var integrationsServiceNowBusinessServicesCmd = &cobra.Command{
	Use:   "business-services",
	Short: "Manage ServiceNow business services",
}

var integrationsServiceNowBusinessServicesListCmd = &cobra.Command{
	Use:   "list [instance-id]",
	Short: "List ServiceNow business services",
	Args:  cobra.ExactArgs(1),
	RunE:  runIntegrationsServiceNowBusinessServicesList,
}

var (
	integrationsFile string
)

func init() {
	// File flags
	integrationsJiraTemplatesCreateCmd.Flags().StringVar(&integrationsFile, "file", "", "JSON file with request body (required)")
	_ = integrationsJiraTemplatesCreateCmd.MarkFlagRequired("file")
	integrationsJiraTemplatesUpdateCmd.Flags().StringVar(&integrationsFile, "file", "", "JSON file with request body (required)")
	_ = integrationsJiraTemplatesUpdateCmd.MarkFlagRequired("file")
	integrationsServiceNowTemplatesCreateCmd.Flags().StringVar(&integrationsFile, "file", "", "JSON file with request body (required)")
	_ = integrationsServiceNowTemplatesCreateCmd.MarkFlagRequired("file")
	integrationsServiceNowTemplatesUpdateCmd.Flags().StringVar(&integrationsFile, "file", "", "JSON file with request body (required)")
	_ = integrationsServiceNowTemplatesUpdateCmd.MarkFlagRequired("file")

	// Existing subcommands
	integrationsSlackCmd.AddCommand(integrationsSlackListCmd)
	integrationsPagerDutyCmd.AddCommand(integrationsPagerDutyListCmd)
	integrationsWebhooksCmd.AddCommand(integrationsWebhooksListCmd)

	// Jira hierarchy
	integrationsJiraAccountsCmd.AddCommand(integrationsJiraAccountsListCmd, integrationsJiraAccountsDeleteCmd)
	integrationsJiraTemplatesCmd.AddCommand(integrationsJiraTemplatesListCmd, integrationsJiraTemplatesGetCmd, integrationsJiraTemplatesCreateCmd, integrationsJiraTemplatesUpdateCmd, integrationsJiraTemplatesDeleteCmd)
	integrationsJiraCmd.AddCommand(integrationsJiraAccountsCmd, integrationsJiraTemplatesCmd)

	// ServiceNow hierarchy
	integrationsServiceNowInstancesCmd.AddCommand(integrationsServiceNowInstancesListCmd)
	integrationsServiceNowTemplatesCmd.AddCommand(integrationsServiceNowTemplatesListCmd, integrationsServiceNowTemplatesGetCmd, integrationsServiceNowTemplatesCreateCmd, integrationsServiceNowTemplatesUpdateCmd, integrationsServiceNowTemplatesDeleteCmd)
	integrationsServiceNowUsersCmd.AddCommand(integrationsServiceNowUsersListCmd)
	integrationsServiceNowAssignmentGroupsCmd.AddCommand(integrationsServiceNowAssignmentGroupsListCmd)
	integrationsServiceNowBusinessServicesCmd.AddCommand(integrationsServiceNowBusinessServicesListCmd)
	integrationsServiceNowCmd.AddCommand(integrationsServiceNowInstancesCmd, integrationsServiceNowTemplatesCmd, integrationsServiceNowUsersCmd, integrationsServiceNowAssignmentGroupsCmd, integrationsServiceNowBusinessServicesCmd)

	integrationsCmd.AddCommand(integrationsSlackCmd, integrationsPagerDutyCmd, integrationsWebhooksCmd, integrationsJiraCmd, integrationsServiceNowCmd)
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

	return formatAndPrint(resp, nil)
}

func runIntegrationsPagerDutyList(cmd *cobra.Command, args []string) error {
	// NOTE: The Datadog API v2.30.0 does not support listing all PagerDuty services.
	// Only GetPagerDutyIntegrationService (singular) with a specific service name is available.
	return fmt.Errorf("listing PagerDuty services is not supported by the current API version - use 'get' with a specific service name instead")
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

	return formatAndPrint(resp, nil)
}

// Jira implementations
func runIntegrationsJiraAccountsList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewJiraIntegrationApi(client.V2())
	resp, r, err := api.ListJiraAccounts(client.Context())
	if err != nil {
		return formatAPIError("list Jira accounts", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runIntegrationsJiraAccountsDelete(cmd *cobra.Command, args []string) error {
	accountID, err := uuid.Parse(args[0])
	if err != nil {
		return fmt.Errorf("invalid account ID: %w", err)
	}

	if !cfg.AutoApprove {
		printOutput("WARNING: This will permanently delete Jira account '%s'.\n", args[0])
		printOutput("Are you sure you want to continue? [y/N]: ")
		response, err := readConfirmation()
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}
		if response != "y" && response != "Y" && response != "yes" {
			printOutput("Operation cancelled.\n")
			return nil
		}
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewJiraIntegrationApi(client.V2())
	r, err := api.DeleteJiraAccount(client.Context(), accountID)
	if err != nil {
		return formatAPIError("delete Jira account", err, r)
	}

	printOutput("Jira account '%s' deleted successfully.\n", args[0])
	return nil
}

func runIntegrationsJiraTemplatesList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewJiraIntegrationApi(client.V2())
	resp, r, err := api.ListJiraIssueTemplates(client.Context())
	if err != nil {
		return formatAPIError("list Jira templates", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runIntegrationsJiraTemplatesGet(cmd *cobra.Command, args []string) error {
	templateID, err := uuid.Parse(args[0])
	if err != nil {
		return fmt.Errorf("invalid template ID: %w", err)
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewJiraIntegrationApi(client.V2())
	resp, r, err := api.GetJiraIssueTemplate(client.Context(), templateID)
	if err != nil {
		return formatAPIError("get Jira template", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runIntegrationsJiraTemplatesCreate(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile(integrationsFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var body datadogV2.JiraIssueTemplateCreateRequest
	if err := json.Unmarshal(data, &body); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewJiraIntegrationApi(client.V2())
	resp, r, err := api.CreateJiraIssueTemplate(client.Context(), body)
	if err != nil {
		return formatAPIError("create Jira template", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runIntegrationsJiraTemplatesUpdate(cmd *cobra.Command, args []string) error {
	templateID, err := uuid.Parse(args[0])
	if err != nil {
		return fmt.Errorf("invalid template ID: %w", err)
	}

	data, err := os.ReadFile(integrationsFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var body datadogV2.JiraIssueTemplateUpdateRequest
	if err := json.Unmarshal(data, &body); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewJiraIntegrationApi(client.V2())
	resp, r, err := api.UpdateJiraIssueTemplate(client.Context(), templateID, body)
	if err != nil {
		return formatAPIError("update Jira template", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runIntegrationsJiraTemplatesDelete(cmd *cobra.Command, args []string) error {
	templateID, err := uuid.Parse(args[0])
	if err != nil {
		return fmt.Errorf("invalid template ID: %w", err)
	}

	if !cfg.AutoApprove {
		printOutput("WARNING: This will permanently delete Jira template '%s'.\n", args[0])
		printOutput("Are you sure you want to continue? [y/N]: ")
		response, err := readConfirmation()
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}
		if response != "y" && response != "Y" && response != "yes" {
			printOutput("Operation cancelled.\n")
			return nil
		}
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewJiraIntegrationApi(client.V2())
	r, err := api.DeleteJiraIssueTemplate(client.Context(), templateID)
	if err != nil {
		return formatAPIError("delete Jira template", err, r)
	}

	printOutput("Jira template '%s' deleted successfully.\n", args[0])
	return nil
}

// ServiceNow implementations
func runIntegrationsServiceNowInstancesList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewServiceNowIntegrationApi(client.V2())
	resp, r, err := api.ListServiceNowInstances(client.Context())
	if err != nil {
		return formatAPIError("list ServiceNow instances", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runIntegrationsServiceNowTemplatesList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewServiceNowIntegrationApi(client.V2())
	resp, r, err := api.ListServiceNowTemplates(client.Context())
	if err != nil {
		return formatAPIError("list ServiceNow templates", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runIntegrationsServiceNowTemplatesGet(cmd *cobra.Command, args []string) error {
	templateID, err := uuid.Parse(args[0])
	if err != nil {
		return fmt.Errorf("invalid template ID: %w", err)
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewServiceNowIntegrationApi(client.V2())
	resp, r, err := api.GetServiceNowTemplate(client.Context(), templateID)
	if err != nil {
		return formatAPIError("get ServiceNow template", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runIntegrationsServiceNowTemplatesCreate(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile(integrationsFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var body datadogV2.ServiceNowTemplateCreateRequest
	if err := json.Unmarshal(data, &body); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewServiceNowIntegrationApi(client.V2())
	resp, r, err := api.CreateServiceNowTemplate(client.Context(), body)
	if err != nil {
		return formatAPIError("create ServiceNow template", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runIntegrationsServiceNowTemplatesUpdate(cmd *cobra.Command, args []string) error {
	templateID, err := uuid.Parse(args[0])
	if err != nil {
		return fmt.Errorf("invalid template ID: %w", err)
	}

	data, err := os.ReadFile(integrationsFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var body datadogV2.ServiceNowTemplateUpdateRequest
	if err := json.Unmarshal(data, &body); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewServiceNowIntegrationApi(client.V2())
	resp, r, err := api.UpdateServiceNowTemplate(client.Context(), templateID, body)
	if err != nil {
		return formatAPIError("update ServiceNow template", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runIntegrationsServiceNowTemplatesDelete(cmd *cobra.Command, args []string) error {
	templateID, err := uuid.Parse(args[0])
	if err != nil {
		return fmt.Errorf("invalid template ID: %w", err)
	}

	if !cfg.AutoApprove {
		printOutput("WARNING: This will permanently delete ServiceNow template '%s'.\n", args[0])
		printOutput("Are you sure you want to continue? [y/N]: ")
		response, err := readConfirmation()
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}
		if response != "y" && response != "Y" && response != "yes" {
			printOutput("Operation cancelled.\n")
			return nil
		}
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewServiceNowIntegrationApi(client.V2())
	r, err := api.DeleteServiceNowTemplate(client.Context(), templateID)
	if err != nil {
		return formatAPIError("delete ServiceNow template", err, r)
	}

	printOutput("ServiceNow template '%s' deleted successfully.\n", args[0])
	return nil
}

func runIntegrationsServiceNowUsersList(cmd *cobra.Command, args []string) error {
	instanceID, err := uuid.Parse(args[0])
	if err != nil {
		return fmt.Errorf("invalid instance ID: %w", err)
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewServiceNowIntegrationApi(client.V2())
	resp, r, err := api.ListServiceNowUsers(client.Context(), instanceID)
	if err != nil {
		return formatAPIError("list ServiceNow users", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runIntegrationsServiceNowAssignmentGroupsList(cmd *cobra.Command, args []string) error {
	instanceID, err := uuid.Parse(args[0])
	if err != nil {
		return fmt.Errorf("invalid instance ID: %w", err)
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewServiceNowIntegrationApi(client.V2())
	resp, r, err := api.ListServiceNowAssignmentGroups(client.Context(), instanceID)
	if err != nil {
		return formatAPIError("list ServiceNow assignment groups", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runIntegrationsServiceNowBusinessServicesList(cmd *cobra.Command, args []string) error {
	instanceID, err := uuid.Parse(args[0])
	if err != nil {
		return fmt.Errorf("invalid instance ID: %w", err)
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewServiceNowIntegrationApi(client.V2())
	resp, r, err := api.ListServiceNowBusinessServices(client.Context(), instanceID)
	if err != nil {
		return formatAPIError("list ServiceNow business services", err, r)
	}

	return formatAndPrint(resp, nil)
}
