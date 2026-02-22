use anyhow::Result;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::api_code_coverage::CodeCoverageAPI;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::model::{
    BranchCoverageSummaryRequest, BranchCoverageSummaryRequestAttributes,
    BranchCoverageSummaryRequestData, BranchCoverageSummaryRequestType,
    CommitCoverageSummaryRequest, CommitCoverageSummaryRequestAttributes,
    CommitCoverageSummaryRequestData, CommitCoverageSummaryRequestType,
};

#[cfg(not(target_arch = "wasm32"))]
use crate::client;
use crate::config::Config;
use crate::formatter;

#[cfg(not(target_arch = "wasm32"))]
pub async fn branch_summary(cfg: &Config, repo: String, branch: String) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => CodeCoverageAPI::with_client_and_config(dd_cfg, c),
        None => CodeCoverageAPI::with_config(dd_cfg),
    };
    let body = BranchCoverageSummaryRequest::new(BranchCoverageSummaryRequestData::new(
        BranchCoverageSummaryRequestAttributes::new(branch, repo),
        BranchCoverageSummaryRequestType::CI_APP_COVERAGE_BRANCH_SUMMARY_REQUEST,
    ));
    let resp = api
        .get_code_coverage_branch_summary(body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to get branch summary: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn branch_summary(cfg: &Config, repo: String, branch: String) -> Result<()> {
    let body = serde_json::json!({
        "data": {
            "type": "ci_app_coverage_branch_summary_request",
            "attributes": {
                "branch": branch,
                "repository_url": repo,
            }
        }
    });
    let data = crate::api::post(cfg, "/api/v2/ci/code-coverage/branch-summary", &body).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn commit_summary(cfg: &Config, repo: String, commit: String) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => CodeCoverageAPI::with_client_and_config(dd_cfg, c),
        None => CodeCoverageAPI::with_config(dd_cfg),
    };
    let body = CommitCoverageSummaryRequest::new(CommitCoverageSummaryRequestData::new(
        CommitCoverageSummaryRequestAttributes::new(commit, repo),
        CommitCoverageSummaryRequestType::CI_APP_COVERAGE_COMMIT_SUMMARY_REQUEST,
    ));
    let resp = api
        .get_code_coverage_commit_summary(body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to get commit summary: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn commit_summary(cfg: &Config, repo: String, commit: String) -> Result<()> {
    let body = serde_json::json!({
        "data": {
            "type": "ci_app_coverage_commit_summary_request",
            "attributes": {
                "commit_sha": commit,
                "repository_url": repo,
            }
        }
    });
    let data = crate::api::post(cfg, "/api/v2/ci/code-coverage/commit-summary", &body).await?;
    crate::formatter::output(cfg, &data)
}
