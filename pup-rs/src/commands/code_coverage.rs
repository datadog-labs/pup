use anyhow::Result;
use datadog_api_client::datadogV2::api_code_coverage::CodeCoverageAPI;
use datadog_api_client::datadogV2::model::{
    BranchCoverageSummaryRequest, BranchCoverageSummaryRequestAttributes,
    BranchCoverageSummaryRequestData, BranchCoverageSummaryRequestType,
    CommitCoverageSummaryRequest, CommitCoverageSummaryRequestAttributes,
    CommitCoverageSummaryRequestData, CommitCoverageSummaryRequestType,
};

use crate::client;
use crate::config::Config;
use crate::formatter;

pub async fn branch_summary(cfg: &Config, repo: String, branch: String) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => CodeCoverageAPI::with_client_and_config(dd_cfg, c),
        None => CodeCoverageAPI::with_config(dd_cfg),
    };
    let body = BranchCoverageSummaryRequest::new(
        BranchCoverageSummaryRequestData::new(
            BranchCoverageSummaryRequestAttributes::new(branch, repo),
            BranchCoverageSummaryRequestType::CI_APP_COVERAGE_BRANCH_SUMMARY_REQUEST,
        ),
    );
    let resp = api
        .get_code_coverage_branch_summary(body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to get branch summary: {e:?}"))?;
    formatter::print_json(&resp)
}

pub async fn commit_summary(cfg: &Config, repo: String, commit: String) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => CodeCoverageAPI::with_client_and_config(dd_cfg, c),
        None => CodeCoverageAPI::with_config(dd_cfg),
    };
    let body = CommitCoverageSummaryRequest::new(
        CommitCoverageSummaryRequestData::new(
            CommitCoverageSummaryRequestAttributes::new(commit, repo),
            CommitCoverageSummaryRequestType::CI_APP_COVERAGE_COMMIT_SUMMARY_REQUEST,
        ),
    );
    let resp = api
        .get_code_coverage_commit_summary(body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to get commit summary: {e:?}"))?;
    formatter::print_json(&resp)
}
