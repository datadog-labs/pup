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
	Short: "Manage on-call teams and schedules",
	Long: `Manage on-call teams, schedules, and rotations.

CAPABILITIES:
  • Manage on-call teams
  • View and manage schedules
  • Handle schedule overrides
  • Track incidents and escalations

EXAMPLES:
  # List all teams
  pup on-call teams list

  # Get team details
  pup on-call teams get team-id

AUTHENTICATION:
  Requires either OAuth2 authentication or API keys.`,
}

var onCallTeamsCmd = &cobra.Command{
	Use:   "teams",
	Short: "Manage on-call teams",
}

var onCallTeamsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List on-call teams",
	RunE:  runOnCallTeamsList,
}

var onCallTeamsGetCmd = &cobra.Command{
	Use:   "get [team-id]",
	Short: "Get team details",
	Args:  cobra.ExactArgs(1),
	RunE:  runOnCallTeamsGet,
}

func init() {
	onCallTeamsCmd.AddCommand(onCallTeamsListCmd, onCallTeamsGetCmd)
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
		if r != nil {
			return fmt.Errorf("failed to list teams: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to list teams: %w", err)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}
	fmt.Println(output)
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
		if r != nil {
			return fmt.Errorf("failed to get team: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to get team: %w", err)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}
