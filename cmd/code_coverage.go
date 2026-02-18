// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/spf13/cobra"
)

var codeCoverageCmd = &cobra.Command{
	Use:   "code-coverage",
	Short: "Query code coverage data",
	Long: `Query code coverage summaries from Datadog Test Optimization.

Code coverage provides branch-level and commit-level coverage summaries
for your repositories.

EXAMPLES:
  # Get branch coverage summary
  pup code-coverage branch-summary --repo="github.com/org/repo" --branch="main"

  # Get commit coverage summary
  pup code-coverage commit-summary --repo="github.com/org/repo" --commit="abc123"

AUTHENTICATION:
  Requires either OAuth2 authentication or API keys.`,
}

var codeCoverageBranchSummaryCmd = &cobra.Command{
	Use:   "branch-summary",
	Short: "Get branch coverage summary",
	RunE:  runCodeCoverageBranchSummary,
}

var codeCoverageCommitSummaryCmd = &cobra.Command{
	Use:   "commit-summary",
	Short: "Get commit coverage summary",
	RunE:  runCodeCoverageCommitSummary,
}

var (
	codeCoverageRepo   string
	codeCoverageBranch string
	codeCoverageCommit string
)

func init() {
	codeCoverageBranchSummaryCmd.Flags().StringVar(&codeCoverageRepo, "repo", "", "Repository name (required)")
	codeCoverageBranchSummaryCmd.Flags().StringVar(&codeCoverageBranch, "branch", "", "Branch name (required)")
	_ = codeCoverageBranchSummaryCmd.MarkFlagRequired("repo")
	_ = codeCoverageBranchSummaryCmd.MarkFlagRequired("branch")

	codeCoverageCommitSummaryCmd.Flags().StringVar(&codeCoverageRepo, "repo", "", "Repository name (required)")
	codeCoverageCommitSummaryCmd.Flags().StringVar(&codeCoverageCommit, "commit", "", "Commit SHA (required)")
	_ = codeCoverageCommitSummaryCmd.MarkFlagRequired("repo")
	_ = codeCoverageCommitSummaryCmd.MarkFlagRequired("commit")

	codeCoverageCmd.AddCommand(codeCoverageBranchSummaryCmd, codeCoverageCommitSummaryCmd)
}

func runCodeCoverageBranchSummary(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewCodeCoverageApi(client.V2())
	attrs := *datadogV2.NewBranchCoverageSummaryRequestAttributes(codeCoverageBranch, codeCoverageRepo)
	data := *datadogV2.NewBranchCoverageSummaryRequestData(attrs, datadogV2.BRANCHCOVERAGESUMMARYREQUESTTYPE_CI_APP_COVERAGE_BRANCH_SUMMARY_REQUEST)
	body := *datadogV2.NewBranchCoverageSummaryRequest(data)

	resp, r, err := api.GetCodeCoverageBranchSummary(client.Context(), body)
	if err != nil {
		return formatAPIError("get branch coverage summary", err, r)
	}

	return formatAndPrint(resp, nil)
}

func runCodeCoverageCommitSummary(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	api := datadogV2.NewCodeCoverageApi(client.V2())
	attrs := *datadogV2.NewCommitCoverageSummaryRequestAttributes(codeCoverageCommit, codeCoverageRepo)
	data := *datadogV2.NewCommitCoverageSummaryRequestData(attrs, datadogV2.COMMITCOVERAGESUMMARYREQUESTTYPE_CI_APP_COVERAGE_COMMIT_SUMMARY_REQUEST)
	body := *datadogV2.NewCommitCoverageSummaryRequest(data)

	resp, r, err := api.GetCodeCoverageCommitSummary(client.Context(), body)
	if err != nil {
		return formatAPIError("get commit coverage summary", err, r)
	}

	return formatAndPrint(resp, nil)
}
