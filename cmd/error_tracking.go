// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/DataDog/pup/pkg/util"
	"github.com/spf13/cobra"
)

var errorTrackingCmd = &cobra.Command{
	Use:   "error-tracking",
	Short: "Manage error tracking",
	Long: `Manage error tracking for application errors and crashes.

Error tracking automatically groups and prioritizes errors from
your applications to help you identify and fix critical issues.

CAPABILITIES:
  • Search error issues with filtering and sorting
  • Get detailed information about a specific issue

EXAMPLES:
  # Search error issues
  pup error-tracking issues search

  # Get issue details
  pup error-tracking issues get issue-id

AUTHENTICATION:
  Requires either OAuth2 authentication or API keys.`,
}

var errorTrackingIssuesCmd = &cobra.Command{
	Use:   "issues",
	Short: "Manage error issues",
}

var errorTrackingIssuesSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search error issues",
	Long: `Search error tracking issues with optional filtering.

Search for error tracking issues across your applications. Results are sorted
by the specified order (default: total count).

FLAGS:
  --query      Search query to filter issues (default: "*")
  --from       Start time: relative (e.g., "1h", "1d", "7d") or absolute
               (ISO 8601, Unix timestamp) (default: "1d")
  --to         End time: "now", relative, or absolute (default: "now")
  --order-by   Sort order: TOTAL_COUNT, FIRST_SEEN, IMPACTED_SESSIONS,
               PRIORITY (default: "TOTAL_COUNT")
  --limit      Maximum number of issues to return (default: 10)

EXAMPLES:
  # Search all issues from the last 24 hours
  pup error-tracking issues search

  # Search for specific errors
  pup error-tracking issues search --query="NullPointerException"

  # Search issues from the last 7 days
  pup error-tracking issues search --from=7d

  # Sort by most recently seen
  pup error-tracking issues search --order-by=FIRST_SEEN

  # Combine filters
  pup error-tracking issues search --query="timeout" --from=3d --order-by=PRIORITY`,
	RunE: runErrorTrackingIssuesSearch,
}

var errorTrackingIssuesGetCmd = &cobra.Command{
	Use:   "get [issue-id]",
	Short: "Get issue details",
	Long: `Get detailed information about a specific error tracking issue.

Retrieves full details for an issue including assignee, case, and team owners.

ARGUMENTS:
  issue-id    The ID of the error tracking issue

EXAMPLES:
  # Get issue details
  pup error-tracking issues get abc123

  # Get issue details as YAML
  pup error-tracking issues get abc123 --output=yaml`,
	Args: cobra.ExactArgs(1),
	RunE: runErrorTrackingIssuesGet,
}

var (
	etQuery   string
	etFrom    string
	etTo      string
	etOrderBy string
	etLimit   int
)

func init() {
	errorTrackingIssuesSearchCmd.Flags().StringVar(&etQuery, "query", "*", "Search query to filter issues")
	errorTrackingIssuesSearchCmd.Flags().StringVar(&etFrom, "from", "1d", "Start time (relative or absolute)")
	errorTrackingIssuesSearchCmd.Flags().StringVar(&etTo, "to", "now", "End time (relative or absolute)")
	errorTrackingIssuesSearchCmd.Flags().StringVar(&etOrderBy, "order-by", "TOTAL_COUNT", "Sort order: TOTAL_COUNT, FIRST_SEEN, IMPACTED_SESSIONS, PRIORITY")
	errorTrackingIssuesSearchCmd.Flags().IntVar(&etLimit, "limit", 10, "Maximum number of issues to return")

	errorTrackingIssuesCmd.AddCommand(errorTrackingIssuesSearchCmd, errorTrackingIssuesGetCmd)
	errorTrackingCmd.AddCommand(errorTrackingIssuesCmd)
}

func runErrorTrackingIssuesSearch(cmd *cobra.Command, args []string) error {
	// Error Tracking API doesn't support OAuth, use API keys
	client, err := getClientForEndpoint("POST", "/api/v2/error_tracking/issues/search")
	if err != nil {
		return err
	}

	fromMs, err := util.ParseTimeToUnixMilli(etFrom)
	if err != nil {
		return err
	}

	toMs, err := util.ParseTimeToUnixMilli(etTo)
	if err != nil {
		return err
	}

	api := datadogV2.NewErrorTrackingApi(client.V2())

	orderBy := datadogV2.IssuesSearchRequestDataAttributesOrderBy(etOrderBy)
	body := datadogV2.IssuesSearchRequest{
		Data: datadogV2.IssuesSearchRequestData{
			Attributes: datadogV2.IssuesSearchRequestDataAttributes{
				From:    fromMs,
				To:      toMs,
				Query:   etQuery,
				OrderBy: &orderBy,
				Persona: datadogV2.ISSUESSEARCHREQUESTDATAATTRIBUTESPERSONA_ALL.Ptr(),
			},
			Type: datadogV2.ISSUESSEARCHREQUESTDATATYPE_SEARCH_REQUEST,
		},
	}

	opts := *datadogV2.NewSearchIssuesOptionalParameters().WithInclude(
		[]datadogV2.SearchIssuesIncludeQueryParameterItem{
			datadogV2.SEARCHISSUESINCLUDEQUERYPARAMETERITEM_ISSUE,
		},
	)

	resp, r, err := api.SearchIssues(client.Context(), body, opts)
	if err != nil {
		return formatAPIError("search error tracking issues", err, r)
	}

	if len(resp.Data) == 0 {
		printOutput("No error tracking issues found matching the specified criteria.\n")
		return nil
	}

	if etLimit > 0 && len(resp.Data) > etLimit {
		resp.Data = resp.Data[:etLimit]
	}

	return formatAndPrint(resp, nil)
}

func runErrorTrackingIssuesGet(cmd *cobra.Command, args []string) error {
	// Error Tracking API doesn't support OAuth, use API keys
	client, err := getClientForEndpoint("GET", "/api/v2/error_tracking/issues/")
	if err != nil {
		return err
	}

	issueID := args[0]
	api := datadogV2.NewErrorTrackingApi(client.V2())

	opts := *datadogV2.NewGetIssueOptionalParameters().WithInclude(
		[]datadogV2.GetIssueIncludeQueryParameterItem{
			datadogV2.GETISSUEINCLUDEQUERYPARAMETERITEM_ASSIGNEE,
			datadogV2.GETISSUEINCLUDEQUERYPARAMETERITEM_CASE,
			datadogV2.GETISSUEINCLUDEQUERYPARAMETERITEM_TEAM_OWNERS,
		},
	)

	resp, r, err := api.GetIssue(client.Context(), issueID, opts)
	if err != nil {
		return formatAPIError("get error tracking issue", err, r)
	}

	return formatAndPrint(resp, nil)
}
