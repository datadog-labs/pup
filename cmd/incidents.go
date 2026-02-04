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

var incidentsCmd = &cobra.Command{
	Use:   "incidents",
	Short: "Manage incidents",
	Long: `Manage Datadog incidents for incident response and tracking.

Incidents provide a centralized place to track, communicate, and resolve issues
affecting your services. They integrate with monitors, timelines, tasks, and
postmortems.

CAPABILITIES:
  • List all incidents with filtering and pagination
  • Get detailed incident information including timeline, tasks, and attachments
  • View incident severity, status, and customer impact
  • Track incident response and resolution

INCIDENT SEVERITIES:
  • SEV-1: Critical impact - complete service outage
  • SEV-2: High impact - major functionality unavailable
  • SEV-3: Moderate impact - partial functionality affected
  • SEV-4: Low impact - minor issues
  • SEV-5: Minimal impact - cosmetic issues

INCIDENT STATES:
  • active: Incident is ongoing, actively being worked
  • stable: Incident is under control but not fully resolved
  • resolved: Incident has been resolved
  • completed: Post-incident tasks completed (postmortem, etc.)

EXAMPLES:
  # List all incidents
  pup incidents list

  # Get detailed incident information
  pup incidents get abc-123-def

  # Get incident and view timeline
  pup incidents get abc-123-def | jq '.data.timeline'

  # Check incident status
  pup incidents get abc-123-def | jq '{status: .data.status, severity: .data.severity}'

INCIDENT FIELDS:
  • id: Incident ID
  • title: Incident title
  • description: Detailed description
  • severity: Severity level (SEV-1 through SEV-5)
  • state: Incident state (active, stable, resolved, completed)
  • customer_impacted: Whether customers are affected
  • customer_impact_scope: Description of customer impact
  • detected_at: When incident was detected
  • created_at: When incident was created in Datadog
  • resolved_at: When incident was resolved
  • commander: Incident commander (user)
  • responders: Team members responding
  • attachments: Related documents, runbooks, etc.

AUTHENTICATION:
  Requires either OAuth2 authentication (pup auth login) or API keys
  (DD_API_KEY and DD_APP_KEY environment variables).`,
}

var incidentsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all incidents",
	Long: `List all incidents with optional filtering.

This command retrieves all incidents from your Datadog account. Results can be
filtered by state, severity, and other criteria.

EXAMPLES:
  # List all incidents
  pup incidents list

  # List incidents with table output
  pup incidents list --output=table

  # Save incidents to file
  pup incidents list > incidents.json

  # Filter active incidents with jq
  pup incidents list | jq '.data[] | select(.state == "active")'

  # Find SEV-1 incidents
  pup incidents list | jq '.data[] | select(.severity == "SEV-1")'

  # Find customer-impacting incidents
  pup incidents list | jq '.data[] | select(.customer_impacted == true)'

OUTPUT FIELDS:
  • id: Incident ID
  • title: Incident title
  • description: Incident description
  • severity: Severity level
  • state: Current state
  • customer_impacted: Boolean flag
  • customer_impact_scope: Impact description
  • customer_impact_start: When impact started
  • customer_impact_end: When impact ended
  • detected_at: Detection timestamp
  • created_at: Creation timestamp
  • modified_at: Last modification timestamp
  • resolved_at: Resolution timestamp (if resolved)
  • commander: Incident commander details
    - name: Commander name
    - email: Commander email
    - handle: Commander handle
  • created_by: User who created incident
  • last_modified_by: User who last modified incident
  • team: Team owning the incident
  • notification_handles: Users/teams to notify

FILTERING:
  Use jq to filter results programmatically:
  • Active only: pup incidents list | jq '.data[] | select(.state == "active")'
  • By severity: pup incidents list | jq '.data[] | select(.severity == "SEV-1")'
  • Customer impact: pup incidents list | jq '.data[] | select(.customer_impacted)'
  • Recent: pup incidents list | jq '.data[] | select(.created_at > "2024-01-01")'

SORTING:
  Incidents are returned sorted by creation time (most recent first).`,
	RunE:  runIncidentsList,
}

var incidentsGetCmd = &cobra.Command{
	Use:   "get [incident-id]",
	Short: "Get incident details",
	Long: `Get complete details for a specific incident.

This command retrieves full incident information including timeline entries,
tasks, attachments, and all metadata.

ARGUMENTS:
  incident-id    The incident ID (format: xxx-xxx-xxx)

EXAMPLES:
  # Get incident details
  pup incidents get abc-123-def

  # Get incident and save to file
  pup incidents get abc-123-def > incident.json

  # View incident timeline
  pup incidents get abc-123-def | jq '.data.timeline'

  # View incident tasks
  pup incidents get abc-123-def | jq '.data.tasks'

  # Check incident status
  pup incidents get abc-123-def | jq '{state: .data.state, severity: .data.severity, customer_impacted: .data.customer_impacted}'

  # Get incident duration
  pup incidents get abc-123-def | jq '{detected: .data.detected_at, resolved: .data.resolved_at}'

OUTPUT STRUCTURE:
  • id: Incident ID
  • title: Incident title
  • description: Detailed description
  • severity: Severity level (SEV-1 through SEV-5)
  • state: Current state
  • customer_impacted: Whether customers affected
  • customer_impact_scope: Description of impact
  • customer_impact_duration: Duration of impact (seconds)
  • detected_at: Detection timestamp (ISO 8601)
  • created_at: Creation timestamp (ISO 8601)
  • modified_at: Last modification timestamp (ISO 8601)
  • resolved_at: Resolution timestamp (ISO 8601, if resolved)
  • time_to_detect: Time from occurrence to detection (seconds)
  • time_to_resolve: Time from detection to resolution (seconds)
  • commander: Incident commander
    - uuid: User UUID
    - name: Full name
    - email: Email address
    - handle: User handle
    - icon: Profile icon URL
  • responders: Array of responding users
  • attachments: Related documents
    - attachment_type: Type (link, postmortem, etc.)
    - attachment: Content/URL
  • timeline: Array of timeline entries
    - timestamp: When event occurred
    - content: Event description
    - creator: User who added entry
  • tasks: Array of incident tasks
    - description: Task description
    - assignee: Assigned user
    - completed_at: Completion timestamp
  • postmortem: Postmortem information
    - published_at: When postmortem was published
    - url: Postmortem URL
  • integration_metadata: Integration data

USE CASES:
  • Track incident progress and timeline
  • Generate incident reports
  • Analyze incident response times
  • Review incident tasks and completion
  • Export incident data for postmortems
  • Monitor customer impact duration`,
	Args:  cobra.ExactArgs(1),
	RunE:  runIncidentsGet,
}

func init() {
	incidentsCmd.AddCommand(incidentsListCmd)
	incidentsCmd.AddCommand(incidentsGetCmd)
}

func runIncidentsList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewIncidentsApi(client.V2())

	resp, r, err := api.ListIncidents(client.Context())
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to list incidents: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to list incidents: %w", err)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}

	fmt.Println(output)
	return nil
}

func runIncidentsGet(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	incidentID := args[0]
	api := datadogV2.NewIncidentsApi(client.V2())

	resp, r, err := api.GetIncident(client.Context(), incidentID)
	if err != nil {
		if r != nil {
			return fmt.Errorf("failed to get incident: %w (status: %d)", err, r.StatusCode)
		}
		return fmt.Errorf("failed to get incident: %w", err)
	}

	output, err := formatter.ToJSON(resp)
	if err != nil {
		return err
	}

	fmt.Println(output)
	return nil
}
