// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"fmt"

	"github.com/DataDog/pup/pkg/formatter"
	"github.com/spf13/cobra"
)

var errorTrackingCmd = &cobra.Command{
	Use:   "error-tracking",
	Short: "Manage error tracking",
	Long: `Manage error tracking for application errors and crashes.

Error tracking automatically groups and prioritizes errors from
your applications to help you identify and fix critical issues.

CAPABILITIES:
  • List error issues
  • Get error details
  • View error trends
  • Manage error status

EXAMPLES:
  # List error issues
  pup error-tracking issues list

  # Get issue details
  pup error-tracking issues get issue-id

AUTHENTICATION:
  Requires either OAuth2 authentication or API keys.`,
}

var errorTrackingIssuesCmd = &cobra.Command{
	Use:   "issues",
	Short: "Manage error issues",
}

var errorTrackingIssuesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List error issues",
	RunE:  runErrorTrackingIssuesList,
}

var errorTrackingIssuesGetCmd = &cobra.Command{
	Use:   "get [issue-id]",
	Short: "Get issue details",
	Args:  cobra.ExactArgs(1),
	RunE:  runErrorTrackingIssuesGet,
}

func init() {
	errorTrackingIssuesCmd.AddCommand(errorTrackingIssuesListCmd, errorTrackingIssuesGetCmd)
	errorTrackingCmd.AddCommand(errorTrackingIssuesCmd)
}

func runErrorTrackingIssuesList(cmd *cobra.Command, args []string) error {
	result := map[string]interface{}{
		"data": []map[string]interface{}{},
		"meta": map[string]interface{}{
			"message": "Error tracking list - API endpoint implementation pending",
		},
	}

	output, err := formatter.ToJSON(result)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}

func runErrorTrackingIssuesGet(cmd *cobra.Command, args []string) error {
	issueID := args[0]
	result := map[string]interface{}{
		"data": map[string]interface{}{
			"id":   issueID,
			"type": "error_tracking_issue",
			"attributes": map[string]interface{}{
				"message": "Error tracking details - API endpoint implementation pending",
			},
		},
	}

	output, err := formatter.ToJSON(result)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}
