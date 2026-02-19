// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/spf13/cobra"
)

var casesCmd = &cobra.Command{
	Use:   "cases",
	Short: "Manage case management cases and projects",
	Long: `Manage Datadog Case Management for tracking and resolving issues.

Case Management provides structured workflows for handling customer issues,
bugs, and internal requests. Cases can be organized into projects with
custom attributes, priorities, and assignments.

CAPABILITIES:
  • Create and manage cases with custom attributes
  • Search and filter cases
  • Assign cases to users
  • Archive/unarchive cases
  • Manage projects
  • Add comments and track timelines

CASE PRIORITIES:
  • NOT_DEFINED: No priority set
  • P1: Critical priority
  • P2: High priority
  • P3: Medium priority
  • P4: Low priority
  • P5: Lowest priority

EXAMPLES:
  # Search cases
  pup cases search --query="bug"

  # Get case details
  pup cases get case-123

  # Create a new case
  pup cases create --title="Bug report" --type-id="type-uuid" --priority=P2

  # List projects
  pup cases projects list

AUTHENTICATION:
  Requires either OAuth2 authentication (pup auth login) or API keys.`,
}

var casesSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search cases",
	Long: `Search cases with optional filtering.

FLAGS:
  --query       Search query string
  --page-size   Results per page (default: 10)
  --page-number Page number (default: 0)

EXAMPLES:
  # Search all cases
  pup cases search

  # Search with query
  pup cases search --query="bug"

  # Search with pagination
  pup cases search --page-size=20 --page-number=1`,
	RunE: runCasesSearch,
}

var casesGetCmd = &cobra.Command{
	Use:   "get [case-id]",
	Short: "Get case details",
	Long: `Get detailed information about a specific case.

ARGUMENTS:
  case-id    The case ID

EXAMPLES:
  # Get case details
  pup cases get case-123`,
	Args: cobra.ExactArgs(1),
	RunE: runCasesGet,
}

var casesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new case",
	Long: `Create a new case with title and type.

REQUIRED FLAGS:
  --title       Case title
  --type-id     Case type UUID

OPTIONAL FLAGS:
  --description Case description
  --priority    Priority: NOT_DEFINED, P1, P2, P3, P4, P5 (default: NOT_DEFINED)

EXAMPLES:
  # Create basic case
  pup cases create --title="Bug report" --type-id="abc-123"

  # Create with priority and description
  pup cases create --title="Critical bug" --type-id="abc-123" --priority=P1 --description="Production issue"`,
	RunE: runCasesCreate,
}

var casesArchiveCmd = &cobra.Command{
	Use:   "archive [case-id]",
	Short: "Archive a case",
	Long: `Archive a case to mark it as completed.

ARGUMENTS:
  case-id    The case ID

EXAMPLES:
  # Archive case
  pup cases archive case-123`,
	Args: cobra.ExactArgs(1),
	RunE: runCasesArchive,
}

var casesUnarchiveCmd = &cobra.Command{
	Use:   "unarchive [case-id]",
	Short: "Unarchive a case",
	Long: `Unarchive a case to reopen it.

ARGUMENTS:
  case-id    The case ID

EXAMPLES:
  # Unarchive case
  pup cases unarchive case-123`,
	Args: cobra.ExactArgs(1),
	RunE: runCasesUnarchive,
}

var casesAssignCmd = &cobra.Command{
	Use:   "assign [case-id]",
	Short: "Assign a case to a user",
	Long: `Assign a case to a specific user.

REQUIRED FLAGS:
  --user-id    User UUID to assign the case to

ARGUMENTS:
  case-id    The case ID

EXAMPLES:
  # Assign case
  pup cases assign case-123 --user-id="user-uuid"`,
	Args: cobra.ExactArgs(1),
	RunE: runCasesAssign,
}

var casesUpdateTitleCmd = &cobra.Command{
	Use:   "update-title [case-id]",
	Short: "Update case title",
	Long: `Update the title of a case.

REQUIRED FLAGS:
  --title      New title

ARGUMENTS:
  case-id    The case ID

EXAMPLES:
  # Update title
  pup cases update-title case-123 --title="New title"`,
	Args: cobra.ExactArgs(1),
	RunE: runCasesUpdateTitle,
}

var casesUpdatePriorityCmd = &cobra.Command{
	Use:   "update-priority [case-id]",
	Short: "Update case priority",
	Long: `Update the priority of a case.

REQUIRED FLAGS:
  --priority   New priority: NOT_DEFINED, P1, P2, P3, P4, P5

ARGUMENTS:
  case-id    The case ID

EXAMPLES:
  # Update priority
  pup cases update-priority case-123 --priority=P1`,
	Args: cobra.ExactArgs(1),
	RunE: runCasesUpdatePriority,
}

// Projects subcommand
var casesProjectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Manage case projects",
	Long: `Create, list, get, and delete case management projects.

Projects organize cases into logical groups with shared settings.`,
}

var casesProjectsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	Long: `List all case management projects.

EXAMPLES:
  # List projects
  pup cases projects list`,
	RunE: runCasesProjectsList,
}

var casesProjectsGetCmd = &cobra.Command{
	Use:   "get [project-id]",
	Short: "Get project details",
	Long: `Get detailed information about a project.

ARGUMENTS:
  project-id    The project ID

EXAMPLES:
  # Get project details
  pup cases projects get project-123`,
	Args: cobra.ExactArgs(1),
	RunE: runCasesProjectsGet,
}

var casesProjectsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new project",
	Long: `Create a new case management project.

REQUIRED FLAGS:
  --name        Project name
  --key         Project key (short identifier)

EXAMPLES:
  # Create project
  pup cases projects create --name="Customer Issues" --key="CUST"`,
	RunE: runCasesProjectsCreate,
}

var casesProjectsDeleteCmd = &cobra.Command{
	Use:   "delete [project-id]",
	Short: "Delete a project",
	Long: `Delete a case management project.

ARGUMENTS:
  project-id    The project ID

FLAGS:
  --yes, -y    Skip confirmation prompt

EXAMPLES:
  # Delete with confirmation
  pup cases projects delete project-123

  # Delete without confirmation
  pup cases projects delete project-123 --yes`,
	Args: cobra.ExactArgs(1),
	RunE: runCasesProjectsDelete,
}

// Jira subcommands
var casesJiraCmd = &cobra.Command{
	Use:   "jira",
	Short: "Manage Jira integrations for cases",
}

var casesJiraCreateIssueCmd = &cobra.Command{
	Use:   "create-issue [case-id]",
	Short: "Create a Jira issue for a case",
	Args:  cobra.ExactArgs(1),
	RunE:  runCasesJiraCreateIssue,
}

var casesJiraLinkCmd = &cobra.Command{
	Use:   "link [case-id]",
	Short: "Link a Jira issue to a case",
	Args:  cobra.ExactArgs(1),
	RunE:  runCasesJiraLink,
}

var casesJiraUnlinkCmd = &cobra.Command{
	Use:   "unlink [case-id]",
	Short: "Unlink a Jira issue from a case",
	Args:  cobra.ExactArgs(1),
	RunE:  runCasesJiraUnlink,
}

// ServiceNow subcommands
var casesServiceNowCmd = &cobra.Command{
	Use:   "servicenow",
	Short: "Manage ServiceNow integrations for cases",
}

var casesServiceNowCreateTicketCmd = &cobra.Command{
	Use:   "create-ticket [case-id]",
	Short: "Create a ServiceNow ticket for a case",
	Args:  cobra.ExactArgs(1),
	RunE:  runCasesServiceNowCreateTicket,
}

// Move subcommand
var casesMoveCmd = &cobra.Command{
	Use:   "move [case-id]",
	Short: "Move a case to a different project",
	Args:  cobra.ExactArgs(1),
	RunE:  runCasesMove,
}

// Projects update subcommand
var casesProjectsUpdateCmd = &cobra.Command{
	Use:   "update [project-id]",
	Short: "Update a project",
	Args:  cobra.ExactArgs(1),
	RunE:  runCasesProjectsUpdate,
}

// Notification rules subcommands
var casesProjectsNotificationRulesCmd = &cobra.Command{
	Use:   "notification-rules",
	Short: "Manage project notification rules",
}

var casesProjectsNotificationRulesListCmd = &cobra.Command{
	Use:   "list [project-id]",
	Short: "List notification rules for a project",
	Args:  cobra.ExactArgs(1),
	RunE:  runCasesProjectsNotificationRulesList,
}

var casesProjectsNotificationRulesCreateCmd = &cobra.Command{
	Use:   "create [project-id]",
	Short: "Create a notification rule",
	Args:  cobra.ExactArgs(1),
	RunE:  runCasesProjectsNotificationRulesCreate,
}

var casesProjectsNotificationRulesUpdateCmd = &cobra.Command{
	Use:   "update [project-id] [rule-id]",
	Short: "Update a notification rule",
	Args:  cobra.ExactArgs(2),
	RunE:  runCasesProjectsNotificationRulesUpdate,
}

var casesProjectsNotificationRulesDeleteCmd = &cobra.Command{
	Use:   "delete [project-id] [rule-id]",
	Short: "Delete a notification rule",
	Args:  cobra.ExactArgs(2),
	RunE:  runCasesProjectsNotificationRulesDelete,
}

var (
	// Case flags
	caseTitle       string
	caseTypeID      string
	caseDescription string
	casePriority    string
	caseUserID      string
	caseQuery       string
	casePageSize    int64
	casePageNumber  int64

	// Project flags
	projectName string
	projectKey  string

	// Cases file flag (for create/update operations)
	casesFile     string
	caseProjectID string
)

func init() {
	// Search flags
	casesSearchCmd.Flags().StringVar(&caseQuery, "query", "", "Search query")
	casesSearchCmd.Flags().Int64Var(&casePageSize, "page-size", 10, "Results per page")
	casesSearchCmd.Flags().Int64Var(&casePageNumber, "page-number", 0, "Page number")

	// Create flags
	casesCreateCmd.Flags().StringVar(&caseTitle, "title", "", "Case title (required)")
	casesCreateCmd.Flags().StringVar(&caseTypeID, "type-id", "", "Case type UUID (required)")
	casesCreateCmd.Flags().StringVar(&caseDescription, "description", "", "Case description")
	casesCreateCmd.Flags().StringVar(&casePriority, "priority", "NOT_DEFINED", "Priority level")
	_ = casesCreateCmd.MarkFlagRequired("title")
	_ = casesCreateCmd.MarkFlagRequired("type-id")

	// Assign flags
	casesAssignCmd.Flags().StringVar(&caseUserID, "user-id", "", "User UUID (required)")
	_ = casesAssignCmd.MarkFlagRequired("user-id")

	// Update title flags
	casesUpdateTitleCmd.Flags().StringVar(&caseTitle, "title", "", "New title (required)")
	_ = casesUpdateTitleCmd.MarkFlagRequired("title")

	// Update priority flags
	casesUpdatePriorityCmd.Flags().StringVar(&casePriority, "priority", "", "New priority (required)")
	_ = casesUpdatePriorityCmd.MarkFlagRequired("priority")

	// Project create flags
	casesProjectsCreateCmd.Flags().StringVar(&projectName, "name", "", "Project name (required)")
	casesProjectsCreateCmd.Flags().StringVar(&projectKey, "key", "", "Project key (required)")
	_ = casesProjectsCreateCmd.MarkFlagRequired("name")
	_ = casesProjectsCreateCmd.MarkFlagRequired("key")

	// File flags for new subcommands
	casesJiraCreateIssueCmd.Flags().StringVar(&casesFile, "file", "", "JSON file with request body (required)")
	_ = casesJiraCreateIssueCmd.MarkFlagRequired("file")
	casesJiraLinkCmd.Flags().StringVar(&casesFile, "file", "", "JSON file with request body (required)")
	_ = casesJiraLinkCmd.MarkFlagRequired("file")
	casesServiceNowCreateTicketCmd.Flags().StringVar(&casesFile, "file", "", "JSON file with request body (required)")
	_ = casesServiceNowCreateTicketCmd.MarkFlagRequired("file")
	casesMoveCmd.Flags().StringVar(&caseProjectID, "project-id", "", "Target project ID (required)")
	_ = casesMoveCmd.MarkFlagRequired("project-id")
	casesProjectsUpdateCmd.Flags().StringVar(&casesFile, "file", "", "JSON file with request body (required)")
	_ = casesProjectsUpdateCmd.MarkFlagRequired("file")
	casesProjectsNotificationRulesCreateCmd.Flags().StringVar(&casesFile, "file", "", "JSON file with request body (required)")
	_ = casesProjectsNotificationRulesCreateCmd.MarkFlagRequired("file")
	casesProjectsNotificationRulesUpdateCmd.Flags().StringVar(&casesFile, "file", "", "JSON file with request body (required)")
	_ = casesProjectsNotificationRulesUpdateCmd.MarkFlagRequired("file")

	// Build command hierarchy
	casesJiraCmd.AddCommand(casesJiraCreateIssueCmd, casesJiraLinkCmd, casesJiraUnlinkCmd)
	casesServiceNowCmd.AddCommand(casesServiceNowCreateTicketCmd)
	casesProjectsNotificationRulesCmd.AddCommand(casesProjectsNotificationRulesListCmd, casesProjectsNotificationRulesCreateCmd, casesProjectsNotificationRulesUpdateCmd, casesProjectsNotificationRulesDeleteCmd)

	casesProjectsCmd.AddCommand(
		casesProjectsListCmd,
		casesProjectsGetCmd,
		casesProjectsCreateCmd,
		casesProjectsDeleteCmd,
		casesProjectsUpdateCmd,
		casesProjectsNotificationRulesCmd,
	)

	casesCmd.AddCommand(
		casesSearchCmd,
		casesGetCmd,
		casesCreateCmd,
		casesArchiveCmd,
		casesUnarchiveCmd,
		casesAssignCmd,
		casesUpdateTitleCmd,
		casesUpdatePriorityCmd,
		casesProjectsCmd,
		casesJiraCmd,
		casesServiceNowCmd,
		casesMoveCmd,
	)
}

func runCasesSearch(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewCaseManagementApi(client.V2())
	opts := datadogV2.SearchCasesOptionalParameters{}

	if caseQuery != "" {
		opts.WithFilter(caseQuery)
	}
	if casePageSize > 0 {
		opts.WithPageSize(casePageSize)
	}
	if casePageNumber > 0 {
		opts.WithPageNumber(casePageNumber)
	}

	resp, r, err := api.SearchCases(client.Context(), opts)
	if err != nil {
		return formatAPIError("search cases", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runCasesGet(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	caseID := args[0]
	api := datadogV2.NewCaseManagementApi(client.V2())

	resp, r, err := api.GetCase(client.Context(), caseID)
	if err != nil {
		return formatAPIError("get case", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runCasesCreate(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	// Build case create request
	attributes := datadogV2.NewCaseCreateAttributes(caseTitle, caseTypeID)

	if caseDescription != "" {
		attributes.SetDescription(caseDescription)
	}

	if casePriority != "" {
		priority, err := datadogV2.NewCasePriorityFromValue(casePriority)
		if err != nil {
			return fmt.Errorf("invalid priority: %w", err)
		}
		attributes.SetPriority(*priority)
	}

	caseData := datadogV2.NewCaseCreate(*attributes, datadogV2.CASERESOURCETYPE_CASE)
	body := datadogV2.NewCaseCreateRequest(*caseData)

	api := datadogV2.NewCaseManagementApi(client.V2())
	resp, r, err := api.CreateCase(client.Context(), *body)
	if err != nil {
		return formatAPIError("create case", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runCasesArchive(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	caseID := args[0]
	api := datadogV2.NewCaseManagementApi(client.V2())

	emptyData := datadogV2.NewCaseEmpty(datadogV2.CASERESOURCETYPE_CASE)
	body := *datadogV2.NewCaseEmptyRequest(*emptyData)
	resp, r, err := api.ArchiveCase(client.Context(), caseID, body)
	if err != nil {
		return formatAPIError("archive case", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runCasesUnarchive(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	caseID := args[0]
	api := datadogV2.NewCaseManagementApi(client.V2())

	emptyData := datadogV2.NewCaseEmpty(datadogV2.CASERESOURCETYPE_CASE)
	body := *datadogV2.NewCaseEmptyRequest(*emptyData)
	resp, r, err := api.UnarchiveCase(client.Context(), caseID, body)
	if err != nil {
		return formatAPIError("unarchive case", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runCasesAssign(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	caseID := args[0]
	api := datadogV2.NewCaseManagementApi(client.V2())

	// Build assign request (simplified - just assignee ID)
	attributes := datadogV2.NewCaseAssignAttributes(caseUserID)
	data := datadogV2.NewCaseAssign(*attributes, datadogV2.CASERESOURCETYPE_CASE)
	body := datadogV2.NewCaseAssignRequest(*data)

	resp, r, err := api.AssignCase(client.Context(), caseID, *body)
	if err != nil {
		return formatAPIError("assign case", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runCasesUpdateTitle(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	caseID := args[0]
	api := datadogV2.NewCaseManagementApi(client.V2())

	// Build update title request
	attributes := datadogV2.NewCaseUpdateTitleAttributes(caseTitle)
	data := datadogV2.NewCaseUpdateTitle(*attributes, datadogV2.CASERESOURCETYPE_CASE)
	body := datadogV2.NewCaseUpdateTitleRequest(*data)

	resp, r, err := api.UpdateCaseTitle(client.Context(), caseID, *body)
	if err != nil {
		return formatAPIError("update case title", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runCasesUpdatePriority(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	caseID := args[0]
	api := datadogV2.NewCaseManagementApi(client.V2())

	priority, err := datadogV2.NewCasePriorityFromValue(casePriority)
	if err != nil {
		return fmt.Errorf("invalid priority: %w", err)
	}

	// Build update priority request
	attributes := datadogV2.NewCaseUpdatePriorityAttributes(*priority)
	data := datadogV2.NewCaseUpdatePriority(*attributes, datadogV2.CASERESOURCETYPE_CASE)
	body := datadogV2.NewCaseUpdatePriorityRequest(*data)

	resp, r, err := api.UpdatePriority(client.Context(), caseID, *body)
	if err != nil {
		return formatAPIError("update case priority", err, r)
	}

	return formatAndPrint(resp, nil)
}

// Project implementations
func runCasesProjectsList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewCaseManagementApi(client.V2())
	resp, r, err := api.GetProjects(client.Context())
	if err != nil {
		return formatAPIError("list projects", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runCasesProjectsGet(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	projectID := args[0]
	api := datadogV2.NewCaseManagementApi(client.V2())

	resp, r, err := api.GetProject(client.Context(), projectID)
	if err != nil {
		return formatAPIError("get project", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runCasesProjectsCreate(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	// Build project create request
	attributes := datadogV2.NewProjectCreateAttributes(projectKey, projectName)
	data := datadogV2.NewProjectCreate(*attributes, datadogV2.PROJECTRESOURCETYPE_PROJECT)
	body := datadogV2.NewProjectCreateRequest(*data)

	api := datadogV2.NewCaseManagementApi(client.V2())
	resp, r, err := api.CreateProject(client.Context(), *body)
	if err != nil {
		return formatAPIError("create project", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runCasesProjectsUpdate(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile(casesFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var body datadogV2.ProjectUpdateRequest
	if err := json.Unmarshal(data, &body); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	projectID := args[0]
	api := datadogV2.NewCaseManagementApi(client.V2())
	resp, r, err := api.UpdateProject(client.Context(), projectID, body)
	if err != nil {
		return formatAPIError("update project", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runCasesProjectsDelete(cmd *cobra.Command, args []string) error {
	projectID := args[0]

	// Confirmation prompt unless --yes flag is set
	if !cfg.AutoApprove {
		printOutput("WARNING: This will permanently delete project '%s'.\n", projectID)
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

	api := datadogV2.NewCaseManagementApi(client.V2())
	r, err := api.DeleteProject(client.Context(), projectID)
	if err != nil {
		return formatAPIError("delete project", err, r)
	}

	printOutput("Project '%s' deleted successfully.\n", projectID)
	return nil
}

// Jira implementations
func runCasesJiraCreateIssue(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile(casesFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var body datadogV2.JiraIssueCreateRequest
	if err := json.Unmarshal(data, &body); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewCaseManagementApi(client.V2())
	r, err := api.CreateCaseJiraIssue(client.Context(), args[0], body)
	if err != nil {
		return formatAPIError("create Jira issue for case", err, r)
	}

	printOutput("Jira issue created for case '%s'.\n", args[0])
	return nil
}

func runCasesJiraLink(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile(casesFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var body datadogV2.JiraIssueLinkRequest
	if err := json.Unmarshal(data, &body); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewCaseManagementApi(client.V2())
	r, err := api.LinkJiraIssueToCase(client.Context(), args[0], body)
	if err != nil {
		return formatAPIError("link Jira issue to case", err, r)
	}

	printOutput("Jira issue linked to case '%s'.\n", args[0])
	return nil
}

func runCasesJiraUnlink(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewCaseManagementApi(client.V2())
	r, err := api.UnlinkJiraIssue(client.Context(), args[0])
	if err != nil {
		return formatAPIError("unlink Jira issue from case", err, r)
	}

	printOutput("Jira issue unlinked from case '%s'.\n", args[0])
	return nil
}

// ServiceNow implementations
func runCasesServiceNowCreateTicket(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile(casesFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var body datadogV2.ServiceNowTicketCreateRequest
	if err := json.Unmarshal(data, &body); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewCaseManagementApi(client.V2())
	r, err := api.CreateCaseServiceNowTicket(client.Context(), args[0], body)
	if err != nil {
		return formatAPIError("create ServiceNow ticket for case", err, r)
	}

	printOutput("ServiceNow ticket created for case '%s'.\n", args[0])
	return nil
}

// Move implementation
func runCasesMove(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	body := datadogV2.ProjectRelationship{
		Data: datadogV2.ProjectRelationshipData{
			Id:   caseProjectID,
			Type: datadogV2.PROJECTRESOURCETYPE_PROJECT,
		},
	}

	api := datadogV2.NewCaseManagementApi(client.V2())
	resp, r, err := api.MoveCaseToProject(client.Context(), args[0], body)
	if err != nil {
		return formatAPIError("move case to project", err, r)
	}

	return formatAndPrint(resp, nil)
}

// Notification Rules implementations
func runCasesProjectsNotificationRulesList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewCaseManagementApi(client.V2())
	resp, r, err := api.GetProjectNotificationRules(client.Context(), args[0])
	if err != nil {
		return formatAPIError("list notification rules", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runCasesProjectsNotificationRulesCreate(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile(casesFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var body datadogV2.CaseNotificationRuleCreateRequest
	if err := json.Unmarshal(data, &body); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewCaseManagementApi(client.V2())
	resp, r, err := api.CreateProjectNotificationRule(client.Context(), args[0], body)
	if err != nil {
		return formatAPIError("create notification rule", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runCasesProjectsNotificationRulesUpdate(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile(casesFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var body datadogV2.CaseNotificationRuleUpdateRequest
	if err := json.Unmarshal(data, &body); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewCaseManagementApi(client.V2())
	r, err := api.UpdateProjectNotificationRule(client.Context(), args[0], args[1], body)
	if err != nil {
		return formatAPIError("update notification rule", err, r)
	}

	printOutput("Notification rule '%s' updated successfully.\n", args[1])
	return nil
}

func runCasesProjectsNotificationRulesDelete(cmd *cobra.Command, args []string) error {
	if !cfg.AutoApprove {
		printOutput("WARNING: This will permanently delete notification rule '%s'.\n", args[1])
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

	api := datadogV2.NewCaseManagementApi(client.V2())
	r, err := api.DeleteProjectNotificationRule(client.Context(), args[0], args[1])
	if err != nil {
		return formatAPIError("delete notification rule", err, r)
	}

	printOutput("Notification rule '%s' deleted successfully.\n", args[1])
	return nil
}
