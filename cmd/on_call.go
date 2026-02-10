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

var onCallCmd = &cobra.Command{
	Use:   "on-call",
	Short: "Manage teams and on-call operations",
	Long: `Manage teams, memberships, links, and notification rules.

Teams in Datadog represent groups of users that collaborate on monitoring,
incident response, and on-call duties. Use this command to manage team
structure, members, and notification settings.

CAPABILITIES:
  • Create, update, and delete teams
  • Manage team memberships and roles
  • Configure team links (documentation, runbooks)
  • Set up notification rules for team alerts

EXAMPLES:
  # List all teams
  pup on-call teams list

  # Create a new team
  pup on-call teams create --name="SRE Team" --handle="sre-team"

  # Add a member to a team
  pup on-call teams memberships add <team-id> --user-id=<uuid> --role=member

  # List team members
  pup on-call teams memberships list <team-id>

AUTHENTICATION:
  Requires either OAuth2 authentication (pup auth login) or API keys.`,
}

var onCallTeamsCmd = &cobra.Command{
	Use:   "teams",
	Short: "Manage teams",
	Long: `Create, update, delete, and query teams.

Teams are groups of users in your Datadog organization that collaborate
on monitoring, incident response, and on-call rotations.`,
}

var onCallTeamsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all teams",
	Long: `List all teams in your organization.

Supports pagination and filtering options.`,
	RunE: runOnCallTeamsList,
}

var onCallTeamsGetCmd = &cobra.Command{
	Use:   "get [team-id]",
	Short: "Get team details",
	Long: `Get detailed information about a specific team.

The team-id is the unique identifier for the team.`,
	Args: cobra.ExactArgs(1),
	RunE: runOnCallTeamsGet,
}

var onCallTeamsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new team",
	Long: `Create a new team in your organization.

REQUIRED FLAGS:
  --name          Team display name
  --handle        Team handle (unique identifier, lowercase, no spaces)

OPTIONAL FLAGS:
  --description   Team description
  --avatar        Team avatar URL
  --color         Team color (hex code, e.g., #00FF00)
  --hidden        Hide team from UI

EXAMPLES:
  # Create a basic team
  pup on-call teams create --name="SRE Team" --handle="sre-team"

  # Create with full details
  pup on-call teams create --name="Platform Engineering" --handle="platform-eng" \
    --description="Platform infrastructure team" --color="#FF5733"`,
	RunE: runOnCallTeamsCreate,
}

var onCallTeamsUpdateCmd = &cobra.Command{
	Use:   "update [team-id]",
	Short: "Update team details",
	Long: `Update an existing team's attributes.

REQUIRED FLAGS:
  --name          Team display name
  --handle        Team handle

OPTIONAL FLAGS:
  --description   Team description
  --avatar        Team avatar URL
  --hidden        Hide team from UI

EXAMPLES:
  # Update team name and handle
  pup on-call teams update abc-123 --name="New Team Name" --handle="new-team"

  # Update with description
  pup on-call teams update abc-123 --name="SRE" --handle="sre" --description="Site Reliability"`,
	Args: cobra.ExactArgs(1),
	RunE: runOnCallTeamsUpdate,
}

var onCallTeamsDeleteCmd = &cobra.Command{
	Use:   "delete [team-id]",
	Short: "Delete a team",
	Long: `Delete a team from your organization.

WARNING: This is a destructive operation. Team deletion will:
  • Remove all team memberships
  • Delete all team links and notification rules
  • Remove team associations from resources

Use --yes to skip confirmation prompt.

EXAMPLES:
  # Delete with confirmation
  pup on-call teams delete abc-123

  # Delete without confirmation
  pup on-call teams delete abc-123 --yes`,
	Args: cobra.ExactArgs(1),
	RunE: runOnCallTeamsDelete,
}

// Team Memberships subcommand
var onCallTeamsMembershipsCmd = &cobra.Command{
	Use:   "memberships",
	Short: "Manage team memberships",
	Long: `Add, update, remove, and list team members.

Team members can have different roles:
  • member: Regular team member with standard permissions
  • admin: Team administrator with full permissions`,
}

var onCallTeamsMembershipsListCmd = &cobra.Command{
	Use:   "list [team-id]",
	Short: "List team members",
	Long: `List all members of a team.

Supports pagination for teams with many members.

FLAGS:
  --page-size     Number of results per page (default: 100)
  --page-number   Page number to retrieve (default: 0)
  --sort          Sort order: name, email (default: name)

EXAMPLES:
  # List all members
  pup on-call teams memberships list abc-123

  # List with pagination
  pup on-call teams memberships list abc-123 --page-size=50 --page-number=1`,
	Args: cobra.ExactArgs(1),
	RunE: runOnCallTeamsMembershipsList,
}

var onCallTeamsMembershipsAddCmd = &cobra.Command{
	Use:   "add [team-id]",
	Short: "Add a member to team",
	Long: `Add a user to a team with a specified role.

REQUIRED FLAGS:
  --user-id       User UUID to add to team
  --role          Member role: member or admin (default: member)

EXAMPLES:
  # Add as regular member
  pup on-call teams memberships add abc-123 --user-id=user-uuid-here --role=member

  # Add as admin
  pup on-call teams memberships add abc-123 --user-id=user-uuid-here --role=admin`,
	Args: cobra.ExactArgs(1),
	RunE: runOnCallTeamsMembershipsAdd,
}

var onCallTeamsMembershipsUpdateCmd = &cobra.Command{
	Use:   "update [team-id] [user-id]",
	Short: "Update member role",
	Long: `Update a team member's role.

REQUIRED FLAGS:
  --role          New role: member or admin

EXAMPLES:
  # Promote to admin
  pup on-call teams memberships update abc-123 user-uuid --role=admin

  # Demote to member
  pup on-call teams memberships update abc-123 user-uuid --role=member`,
	Args: cobra.ExactArgs(2),
	RunE: runOnCallTeamsMembershipsUpdate,
}

var onCallTeamsMembershipsRemoveCmd = &cobra.Command{
	Use:   "remove [team-id] [user-id]",
	Short: "Remove member from team",
	Long: `Remove a user from a team.

Use --yes to skip confirmation prompt.

EXAMPLES:
  # Remove with confirmation
  pup on-call teams memberships remove abc-123 user-uuid

  # Remove without confirmation
  pup on-call teams memberships remove abc-123 user-uuid --yes`,
	Args: cobra.ExactArgs(2),
	RunE: runOnCallTeamsMembershipsRemove,
}

var (
	// Team flags
	teamName        string
	teamHandle      string
	teamDescription string
	teamAvatar      string
	teamHidden      bool

	// Membership flags
	memberUserID   string
	memberRole     string
	memberPageSize int64
	memberPageNum  int64
	memberSort     string
)

func init() {
	// Team create flags
	onCallTeamsCreateCmd.Flags().StringVar(&teamName, "name", "", "Team display name (required)")
	onCallTeamsCreateCmd.Flags().StringVar(&teamHandle, "handle", "", "Team handle (required)")
	onCallTeamsCreateCmd.Flags().StringVar(&teamDescription, "description", "", "Team description")
	onCallTeamsCreateCmd.Flags().StringVar(&teamAvatar, "avatar", "", "Team avatar URL")
	onCallTeamsCreateCmd.Flags().BoolVar(&teamHidden, "hidden", false, "Hide team from UI")
	_ = onCallTeamsCreateCmd.MarkFlagRequired("name")
	_ = onCallTeamsCreateCmd.MarkFlagRequired("handle")

	// Team update flags
	onCallTeamsUpdateCmd.Flags().StringVar(&teamName, "name", "", "Team display name (required)")
	onCallTeamsUpdateCmd.Flags().StringVar(&teamHandle, "handle", "", "Team handle (required)")
	onCallTeamsUpdateCmd.Flags().StringVar(&teamDescription, "description", "", "Team description")
	onCallTeamsUpdateCmd.Flags().StringVar(&teamAvatar, "avatar", "", "Team avatar URL")
	onCallTeamsUpdateCmd.Flags().BoolVar(&teamHidden, "hidden", false, "Hide team from UI")
	_ = onCallTeamsUpdateCmd.MarkFlagRequired("name")
	_ = onCallTeamsUpdateCmd.MarkFlagRequired("handle")

	// Membership list flags
	onCallTeamsMembershipsListCmd.Flags().Int64Var(&memberPageSize, "page-size", 100, "Results per page")
	onCallTeamsMembershipsListCmd.Flags().Int64Var(&memberPageNum, "page-number", 0, "Page number")
	onCallTeamsMembershipsListCmd.Flags().StringVar(&memberSort, "sort", "name", "Sort order: name, email")

	// Membership add flags
	onCallTeamsMembershipsAddCmd.Flags().StringVar(&memberUserID, "user-id", "", "User UUID (required)")
	onCallTeamsMembershipsAddCmd.Flags().StringVar(&memberRole, "role", "member", "Role: member or admin")
	_ = onCallTeamsMembershipsAddCmd.MarkFlagRequired("user-id")

	// Membership update flags
	onCallTeamsMembershipsUpdateCmd.Flags().StringVar(&memberRole, "role", "", "Role: member or admin (required)")
	_ = onCallTeamsMembershipsUpdateCmd.MarkFlagRequired("role")

	// Build command hierarchy
	onCallTeamsMembershipsCmd.AddCommand(
		onCallTeamsMembershipsListCmd,
		onCallTeamsMembershipsAddCmd,
		onCallTeamsMembershipsUpdateCmd,
		onCallTeamsMembershipsRemoveCmd,
	)

	onCallTeamsCmd.AddCommand(
		onCallTeamsListCmd,
		onCallTeamsGetCmd,
		onCallTeamsCreateCmd,
		onCallTeamsUpdateCmd,
		onCallTeamsDeleteCmd,
		onCallTeamsMembershipsCmd,
	)

	onCallCmd.AddCommand(onCallTeamsCmd)
}

func runOnCallTeamsList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewTeamsApi(client.V2())
	resp, r, err := api.ListTeams(client.Context())
	if err != nil {
		return formatAPIError("list teams", err, r)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}
	printOutput("%s\n", output)
	return nil
}

func runOnCallTeamsGet(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	teamID := args[0]
	api := datadogV2.NewTeamsApi(client.V2())
	resp, r, err := api.GetTeam(client.Context(), teamID)
	if err != nil {
		return formatAPIError("get team", err, r)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}
	printOutput("%s\n", output)
	return nil
}

func runOnCallTeamsCreate(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	// Build team create request
	attributes := datadogV2.NewTeamCreateAttributes(teamHandle, teamName)

	if teamDescription != "" {
		attributes.SetDescription(teamDescription)
	}
	if teamAvatar != "" {
		attributes.SetAvatar(teamAvatar)
	}
	if teamHidden {
		attributes.SetHiddenModules([]string{"hidden"})
	}

	teamData := datadogV2.NewTeamCreate(*attributes, datadogV2.TEAMTYPE_TEAM)
	body := datadogV2.NewTeamCreateRequest(*teamData)

	api := datadogV2.NewTeamsApi(client.V2())
	resp, r, err := api.CreateTeam(client.Context(), *body)
	if err != nil {
		return formatAPIError("create team", err, r)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}
	printOutput("%s\n", output)
	return nil
}

func runOnCallTeamsUpdate(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	teamID := args[0]

	// Build team update request (handle and name are required)
	attributes := datadogV2.NewTeamUpdateAttributes(teamHandle, teamName)

	if teamDescription != "" {
		attributes.SetDescription(teamDescription)
	}
	if teamAvatar != "" {
		attributes.SetAvatar(teamAvatar)
	}
	if teamHidden {
		attributes.SetHiddenModules([]string{"hidden"})
	}

	teamData := datadogV2.NewTeamUpdate(*attributes, datadogV2.TEAMTYPE_TEAM)
	body := datadogV2.NewTeamUpdateRequest(*teamData)

	api := datadogV2.NewTeamsApi(client.V2())
	resp, r, err := api.UpdateTeam(client.Context(), teamID, *body)
	if err != nil {
		return formatAPIError("update team", err, r)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}
	printOutput("%s\n", output)
	return nil
}

func runOnCallTeamsDelete(cmd *cobra.Command, args []string) error {
	teamID := args[0]

	// Confirmation prompt unless --yes flag is set
	if !cfg.AutoApprove {
		printOutput("WARNING: This will permanently delete team '%s' and all associated data.\n", teamID)
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

	api := datadogV2.NewTeamsApi(client.V2())
	r, err := api.DeleteTeam(client.Context(), teamID)
	if err != nil {
		return formatAPIError("delete team", err, r)
	}

	printOutput("Team '%s' deleted successfully.\n", teamID)
	return nil
}

// Team Memberships implementations
func runOnCallTeamsMembershipsList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	teamID := args[0]
	api := datadogV2.NewTeamsApi(client.V2())

	opts := datadogV2.GetTeamMembershipsOptionalParameters{}
	if memberPageSize > 0 {
		opts.WithPageSize(memberPageSize)
	}
	if memberPageNum > 0 {
		opts.WithPageNumber(memberPageNum)
	}
	if memberSort != "" {
		sortVal, err := datadogV2.NewGetTeamMembershipsSortFromValue(memberSort)
		if err != nil {
			return fmt.Errorf("invalid sort value: %w", err)
		}
		opts.WithSort(*sortVal)
	}

	resp, r, err := api.GetTeamMemberships(client.Context(), teamID, opts)
	if err != nil {
		return formatAPIError("list team memberships", err, r)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}
	printOutput("%s\n", output)
	return nil
}

func runOnCallTeamsMembershipsAdd(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	teamID := args[0]

	// Build membership request
	userData := datadogV2.NewRelationshipToUserTeamUserData(memberUserID, datadogV2.USERTEAMUSERTYPE_USERS)
	user := datadogV2.NewRelationshipToUserTeamUser(*userData)

	teamData := datadogV2.NewRelationshipToUserTeamTeamData(teamID, datadogV2.USERTEAMTEAMTYPE_TEAM)
	team := datadogV2.NewRelationshipToUserTeamTeam(*teamData)

	relationships := datadogV2.NewUserTeamRelationships()
	relationships.SetUser(*user)
	relationships.SetTeam(*team)

	userTeam := datadogV2.NewUserTeamCreate(datadogV2.USERTEAMTYPE_TEAM_MEMBERSHIPS)
	userTeam.SetRelationships(*relationships)

	if memberRole != "" {
		attributes := datadogV2.NewUserTeamAttributes()
		role := datadogV2.UserTeamRole(memberRole)
		attributes.SetRole(role)
		userTeam.SetAttributes(*attributes)
	}

	body := datadogV2.NewUserTeamRequest(*userTeam)

	api := datadogV2.NewTeamsApi(client.V2())
	resp, r, err := api.CreateTeamMembership(client.Context(), teamID, *body)
	if err != nil {
		return formatAPIError("add team member", err, r)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}
	printOutput("%s\n", output)
	return nil
}

func runOnCallTeamsMembershipsUpdate(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	teamID := args[0]
	userID := args[1]

	// Build membership update request
	attributes := datadogV2.NewUserTeamAttributes()
	if memberRole != "" {
		role := datadogV2.UserTeamRole(memberRole)
		attributes.SetRole(role)
	}

	userTeam := datadogV2.NewUserTeamUpdate(datadogV2.USERTEAMTYPE_TEAM_MEMBERSHIPS)
	userTeam.SetAttributes(*attributes)
	body := datadogV2.NewUserTeamUpdateRequest(*userTeam)

	api := datadogV2.NewTeamsApi(client.V2())
	resp, r, err := api.UpdateTeamMembership(client.Context(), teamID, userID, *body)
	if err != nil {
		return formatAPIError("update team membership", err, r)
	}

	output, err := formatter.FormatOutput(resp, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}
	printOutput("%s\n", output)
	return nil
}

func runOnCallTeamsMembershipsRemove(cmd *cobra.Command, args []string) error {
	teamID := args[0]
	userID := args[1]

	// Confirmation prompt unless --yes flag is set
	if !cfg.AutoApprove {
		printOutput("WARNING: This will remove user '%s' from team '%s'.\n", userID, teamID)
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

	api := datadogV2.NewTeamsApi(client.V2())
	r, err := api.DeleteTeamMembership(client.Context(), teamID, userID)
	if err != nil {
		return formatAPIError("remove team member", err, r)
	}

	printOutput("User '%s' removed from team '%s' successfully.\n", userID, teamID)
	return nil
}
